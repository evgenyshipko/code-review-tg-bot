package storage

import (
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/memoryStorage"
	"code-review-tg-bot/internal/redis"
	"code-review-tg-bot/internal/redisStorage"
)

type Storage interface {
	Set(key string, value interface{})
	Get(key string, result interface{}) bool
	Delete(key string)
}

func InitStorage() Storage {
	err := redis.Init()
	if err != nil {
		logger.Instance.Debugf("Ошибка инициализации Redis, используется Memory", "error", err)
		return memoryStorage.NewMemoryStorage()
	}

	logger.Instance.Info("Успешная инициализации Redis")
	return redisStorage.NewRedisStorage(redis.GetClient())
}
