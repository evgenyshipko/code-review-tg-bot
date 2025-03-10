package storage

import (
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/memoryStorage"
	"code-review-tg-bot/internal/redis"
	"code-review-tg-bot/internal/redisStorage"
	"time"
)

type Storage interface {
	Set(key string, value interface{})
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
	Get(key string, result interface{}) bool
	Delete(key string)
}

func InitStorage() (Storage, error) {
	err := redis.Init()
	if err != nil {
		logger.Instance.Error("Ошибка инициализации Redis", "error", err)
		return memoryStorage.NewMemoryStorage(), nil

	}

	logger.Instance.Info("Успешная инициализации Redis")
	return redisStorage.NewRedisStorage(redis.GetClient()), nil

}
