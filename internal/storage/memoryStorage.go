package storage

import (
	"encoding/json"
	"sync"
	"time"

	"code-review-tg-bot/internal/logger"
)

type memoryData struct {
	Value    string
	ExpireAt *time.Time // время истечения
}

type MemoryStorage struct {
	data  map[string]memoryData
	mutex sync.Mutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]memoryData),
	}
}

func (s *MemoryStorage) Set(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result, err := json.Marshal(value)
	if err != nil {
		logger.Instance.Error("Ошибка сериализации", err.Error())
		return
	}

	s.data[key] = memoryData{
		Value: string(result),
	}
	logger.Instance.Debugw("Значение сохранено в Memory", "key", key)
}

func (s *MemoryStorage) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result, err := json.Marshal(value)
	if err != nil {
		return err
	}

	expireAt := time.Now().Add(ttl)
	s.data[key] = memoryData{
		Value:    string(result),
		ExpireAt: &expireAt,
	}

	logger.Instance.Debugw("Значение сохранено в Memory с TTL",
		"key", key,
		"expire_at", expireAt.Format(time.RFC3339))
	return nil
}

func (s *MemoryStorage) Get(key string, result interface{}) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, exists := s.data[key]
	if !exists {
		logger.Instance.Debugw("Ключ не найден в Memory", "key", key)
		return false
	}

	// Проверка срока истечения
	if data.ExpireAt != nil && time.Now().After(*data.ExpireAt) {
		delete(s.data, key)
		logger.Instance.Debugw("Значение удалено по истечению TTL", "key", key)
		return false
	}

	err := json.Unmarshal([]byte(data.Value), &result)
	if err != nil {
		logger.Instance.Error("Ошибка десериализации", "error", err)
		return false
	}

	logger.Instance.Debugw("Значение получено из Memory", "key", key)
	return true
}

func (s *MemoryStorage) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, key)
	logger.Instance.Debugw("Значение удалено из Memory", "key", key)
}
