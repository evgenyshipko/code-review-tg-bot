package memoryStorage

import (
	"encoding/json"
	"sync"

	"code-review-tg-bot/internal/logger"
)

type MemoryStorage struct {
	data  map[string]string
	mutex sync.Mutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (s *MemoryStorage) Set(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result, err := json.Marshal(value)
	if err != nil {
		logger.Instance.Error("Ошибка сериализации", err.Error())
	}

	s.data[key] = string(result)
	logger.Instance.Debugw("Значение сохранено в Memory", "key", key)
}

func (s *MemoryStorage) Get(key string, result interface{}) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, exists := s.data[key]
	if !exists {
		logger.Instance.Debugw("Ключ не найден в Memory", "key", key)
		return false
	}

	err := json.Unmarshal([]byte(value), &result)
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
