package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) (done bool, err error) {
	data, err := r.serializer.Marshal(value)
	if err != nil {
		log.Printf("[Cache] Failed to marshal value for key %s: %v\n", key, err)
		return false, err
	}

	if err = r.Client.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("[Cache] Failed to set key %s in Redis: %v\n", key, err)
		return false, err
	}

	return true, nil
}

func (r *Redis) Get(ctx context.Context, key string, out interface{}) (found bool, err error) {
	cmd := r.Client.Get(ctx, key)
	if err = cmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // key doesn't exist
		}
		log.Printf("[Cache] Failed to get key %s from Redis: %v\n", key, err)
		return false, err
	}

	str, err := cmd.Bytes()
	if err != nil {
		log.Printf("[Cache] Failed to read Redis bytes for key %s: %v\n", key, err)
		return false, err
	}

	if err = r.serializer.Unmarshal(str, out); err != nil {
		log.Printf("[Cache] Failed to unmarshal value for key %s: %v\n", key, err)
		return false, err
	}

	return true, nil
}

func (r *Redis) Unlink(ctx context.Context, keys []string) (int64, error) {
	result := r.Client.Unlink(ctx, keys...)
	if err := result.Err(); err != nil {
		log.Printf("[Cache] Failed to unlink keys %v: %v\n", keys, err)
		return 0, err
	}

	return result.Val(), nil
}

func (r *Redis) MSet(ctx context.Context, keyValue map[string]any) error {
	values := make([]any, 0)
	for k, v := range keyValue {
		b, err := r.serializer.Marshal(v)
		if err != nil {
			log.Printf("[Cache] MSet Serialise Cache key %v failed. err: %s", v, err)

			return err
		}

		values = append(values, k, string(b))
	}

	err := r.Client.MSet(ctx, values...).Err()
	if err != nil {
		log.Printf("[Cache] MSet Cache keys %v failed. err: %s", values, err)

		return err
	}

	return nil
}

func (r *Redis) MGet(ctx context.Context, keys []string, model any) ([]interface{}, error) {
	//output := make([]interface{}, 0)
	values, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		log.Printf("[Cache] MGet Multi Redis keys %v failed. err: %s", keys, err)

		return nil, err
	}

	for _, v := range values {
		if v == nil {
			continue
		}

		val, ok := v.(string)
		if !ok {
			continue
		}

		data, marshalErr := r.serializer.Marshal([]byte(val))
		if marshalErr != nil {
			log.Printf("[Cache] MGet Serialise Cache key %v failed. err: %s", val, err)

			return nil, marshalErr
		}

		r.serializer.Unmarshal(data, model)
	}

	return values, nil
}

//func (r *Redis) PipedMSet(ctx context.Context, kvArr []KVIn, d time.Duration) error {
//	p := r.Client.Pipeline()
//
//	for _, kv := range kvArr {
//		err := p.Set(ctx, kv.Key, kv.Val, d).Err()
//		if err != nil {
//			log.Printf("[Cache] PipedMSet failed. key: %s, err: %v", kv.Key, err)
//
//			return err
//		}
//	}
//
//	_, err := p.Exec(ctx)
//	if err != nil {
//		log.Printf("[Cache] PipedMSet failed. err: %v", err)
//
//		return err
//	}
//
//	return nil
//}
//
//func (r *Redis) PipedMGet(ctx context.Context, kvArr []*KVOut) (err error) {
//	p := r.Client.Pipeline()
//
//	cmds := make([]*redis.StringCmd, len(kvArr))
//	for i, kv := range kvArr {
//		cmds[i] = p.Get(ctx, kv.Key)
//	}
//
//	_, err = p.Exec(ctx)
//	if err != nil && !errors.Is(err, redis.Nil) {
//		log.Printf("[Cache] PipedExec failed. err: %v", err)
//
//		return err
//	}
//
//	for i, cmd := range cmds {
//		kvArr[i].exists, kvArr[i].err = r.result(cmd, kvArr[i].Val)
//	}
//
//	return nil
//}
//

func (s *Redis) result(cmd *redis.StringCmd, val interface{}) (ok bool, err error) {
	if err = cmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}

		return false, err
	}

	bytes, err := cmd.Bytes()
	if err != nil {
		return false, err
	}

	if err = s.serializer.Unmarshal(bytes, val); err != nil {
		return false, fmt.Errorf("deserialize error: %w", err)
	}

	return true, nil
}

func (r *Redis) PipedMGet(ctx context.Context, kvArr []*KVOut) error {
	pipe := r.Client.Pipeline()
	cmds := make([]*redis.StringCmd, len(kvArr))

	for i, kv := range kvArr {
		cmds[i] = pipe.Get(ctx, kv.Key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Printf("[Cache] Pipe Redis keys %v failed. err: %s", kvArr, err)

		return err
	}

	for i, cmd := range cmds {
		kvArr[i].exists, kvArr[i].err = r.result(cmd, kvArr[i].Val)
	}

	return nil
}

func (r *Redis) PipedMSet(ctx context.Context, kvArr []KVIn, expiry time.Duration) error {
	pipe := r.Client.Pipeline()

	for _, kv := range kvArr {
		bytes, err := r.serializer.Marshal(kv.Val)
		if err != nil {
			return fmt.Errorf("marshal error for key %s: %w", kv.Key, err)
		}
		pipe.Set(ctx, kv.Key, bytes, expiry)
	}

	_, err := pipe.Exec(ctx)
	return err
}
