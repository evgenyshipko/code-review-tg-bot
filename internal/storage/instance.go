package storage

import (
	"code-review-tg-bot/internal/memoryStorage"
	"code-review-tg-bot/internal/redis"
	"code-review-tg-bot/internal/redisStorage"
)

// Глобальный экземпляр хранилища
var instance Storage

// Инициализация хранилища
func InitStorage() error {
	if instance == nil {
		var err error
		instance, err = NewStorage()

		if err != nil {
			return err
		}
	}

	return nil
}

// Создание нового хранилища
func NewStorage() (Storage, error) {
	err := redis.Init()
	if err != nil {
		return memoryStorage.NewMemoryStorage(), err
	}

	return redisStorage.NewRedisStorage(redis.GetClient()), nil
}

// Получение экземпляра хранилища
func GetStorage() Storage {
	return instance
}
