package storage

import (
	"code-review-tg-bot/internal/memoryStorage"
	"code-review-tg-bot/internal/redis"
	"code-review-tg-bot/internal/redisStorage"
)

type Storage interface {
	Set(key string, value interface{})
	Get(key string, result interface{}) bool
	Delete(key string)
}

func NewStorage() Storage {
	err := redis.Init()
	if err != nil {
		return memoryStorage.NewMemoryStorage()
	}

	return redisStorage.NewRedisStorage(redis.GetClient())
}
