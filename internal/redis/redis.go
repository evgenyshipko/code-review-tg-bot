package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client    *redis.Client
	namespace string
}

func InitRedisClient() (*RedisClient, error) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" || redisPort == "" {
		return nil, fmt.Errorf("необходимо указать REDIS_HOST и REDIS_PORT")
	}

	client := NewRedisClient(fmt.Sprintf("%s:%s", redisHost, redisPort), os.Getenv("APP_ENVIRONMENT"))

	// Проверка подключения
	_, err := client.client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	return client, nil
}

func NewRedisClient(addr, namespace string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{client: rdb, namespace: namespace}
}

func (r *RedisClient) withNamespace(key string) string {
	return r.namespace + ":" + key
}

func (r *RedisClient) Set(key string, value interface{}) error {
	return r.client.Set(context.Background(), r.withNamespace(key), value, -1).Err()
}

func (r *RedisClient) SetWithTTL(key string, value string, ttl time.Duration) error {
	return r.client.Set(context.Background(), r.withNamespace(key), value, ttl).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(context.Background(), r.withNamespace(key)).Result()
}

func (r *RedisClient) Delete(key string) error {
	return r.client.Del(context.Background(), r.withNamespace(key)).Err()
}
