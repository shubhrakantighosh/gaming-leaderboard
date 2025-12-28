package redis

import (
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client     *redis.Client
	serializer ISerializer
}

var redisInstance *Redis

func GetClient() *Redis {
	return redisInstance
}

func SetClient(client *redis.Client) {
	redisInstance = &Redis{Client: client, serializer: NewMsgpackSerializer()}
}

type Cmdable interface {
	redis.Cmdable
	Close() error
}

// Client represents a Redis Client Connection
type Client struct {
	Nil error
	Cmdable
}
