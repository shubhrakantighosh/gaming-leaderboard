package redis

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) (done bool, err error)
	Get(ctx context.Context, key string, out interface{}) (found bool, err error)
	MSet(ctx context.Context, keyValue map[string]any) error
	//MGet(ctx context.Context, keys []string) ([]interface{}, error)
	PipedMSet(ctx context.Context, kvArr []KVIn, d time.Duration) error
	PipedMGet(ctx context.Context, kvArr []*KVOut) (err error)
	Unlink(ctx context.Context, keys []string) (int64, error)
}

type KVIn struct {
	Key string
	Val interface{}
}

type KVOut struct {
	Key    string
	Val    interface{}
	err    error
	exists bool
}

func (ko KVOut) OK() bool {
	return ko.exists && ko.err == nil
}
