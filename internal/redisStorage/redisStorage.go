package redisStorage

import (
	"encoding/json"
	"fmt"
	"time"

	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/redis"
)

type RedisStorage struct {
	client *redis.RedisClient
}

func NewRedisStorage(client *redis.RedisClient) *RedisStorage {
	return &RedisStorage{
		client: client,
	}
}

func (s *RedisStorage) Set(key string, value interface{}) {
	result, err := json.Marshal(value)
	if err != nil {
		logger.Instance.Error("Ошибка сериализации в Redis", "error", err.Error())
		return
	}

	err = s.client.Set(key, string(result))
	if err != nil {
		logger.Instance.Error("Ошибка сохранения в БД", "error", err)
		return
	}

	logger.Instance.Debugw("Установлено значение в Redis", "key", key, "value", value)
}

func (s *RedisStorage) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	result, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	err = s.client.SetWithTTL(key, string(result), ttl)
	if err != nil {
		return fmt.Errorf("ошибка сохранения в Redis с TTL: %w", err)
	}

	logger.Instance.Debugw("Значение сохранено в Redis с TTL",
		"key", key,
		"ttl", ttl,
		"ttl_days", ttl.Hours()/24,
		"ttl_hours", ttl.Hours())

	return nil
}

func (s *RedisStorage) Get(key string, result interface{}) bool {
	value, err := s.client.Get(key)
	if err != nil {
		logger.Instance.Debugw("Ключ не найден в Redis", "key", key, "error", err.Error())
		return false
	}

	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		logger.Instance.Error("Ошибка десериализации", "error", err)
		return false
	}

	logger.Instance.Debugw("Получено значение из Redis", "key", key, "value", value)
	return true
}

func (s *RedisStorage) Delete(key string) {
	err := s.client.Delete(key)
	if err != nil {
		logger.Instance.Error("Ошибка удаления из Redis", "error", err.Error())
	}
}
