package storage

import (
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/redis"
	"os"
	"time"
)

type Storage interface {
	Set(key string, value interface{})
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
	Get(key string, result interface{}) bool
	Delete(key string)
}

func InitStorage() (Storage, error) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	var err error

	if redisHost == "" && redisPort == "" {
		logger.Instance.Debug("Ошибка инициализации Redis, используется Memory")

		return NewMemoryStorage(), nil
	}

	client, err := redis.InitRedisClient()
	if err != nil {
		logger.Instance.Error("Ошибка инициализации Redis", "error", err)
		return nil, err
	}

	logger.Instance.Info("Успешная инициализации Redis")
	return NewRedisStorage(client), nil

}
