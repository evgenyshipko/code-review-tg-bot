package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	client *RedisClient
	ctx    = context.Background()
)

type RedisClient struct {
	client    *redis.Client
	namespace string
}

// Init инициализирует Redis-клиент
func Init() error {
	if client != nil {
		return nil
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" || redisPort == "" {
		return fmt.Errorf("необходимо указать REDIS_HOST и REDIS_PORT")
	}

	// Создаем глобальный клиент
	client = NewRedisClient(fmt.Sprintf("%s:%s", redisHost, redisPort), os.Getenv("APP_ENVIRONMENT"))

	// Проверка подключения
	_, err := client.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	return nil
}

// Фабрика для клиента с namespace
func NewRedisClient(addr, namespace string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{client: rdb, namespace: namespace}
}

// Добавление namespace к ключу
func (r *RedisClient) withNamespace(key string) string {
	return r.namespace + ":" + key
}

// Set записывает значение
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, r.withNamespace(key), value, expiration).Err()
}

// Get получает значение
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(ctx, r.withNamespace(key)).Result()
}

// Delete удаляет ключ
func (r *RedisClient) Delete(key string) error {
	return r.client.Del(ctx, r.withNamespace(key)).Err()
}

// GetClient возвращает глобальный клиент
func GetClient() *RedisClient {
	return client
}
