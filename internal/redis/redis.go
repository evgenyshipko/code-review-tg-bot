package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	ctx    = context.Background()
)

func Init() error {
	if client != nil {
		return nil
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	if redisHost == "" || redisPort == "" {
		return fmt.Errorf("необходимо указать REDIS_HOST и REDIS_PORT")
	}

	// Конфиг
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	// Проверка подключения
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	return nil
}

func Set(key string, value interface{}, expiration time.Duration) error {
	if client == nil {
		return fmt.Errorf("redis клиент не инициализирован")
	}

	return client.Set(ctx, key, value, expiration).Err()
}

func Get(key string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("redis клиент не инициализирован")
	}

	return client.Get(ctx, key).Result()
}

func Delete(key string) error {
	if client == nil {
		return fmt.Errorf("redis клиент не инициализирован")
	}

	return client.Del(ctx, key).Err()
}
