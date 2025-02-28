package storage

import (
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/redis"
	"encoding/json"
	"time"
)

const defaultExpiration = 24 * time.Hour

func Set(key string, value interface{}) {
	result, err := json.Marshal(value)
	if err != nil {
		logger.Instance.Error("Ошибка сериализации в Redis", "error", err.Error())
	}

	err = redis.GetClient().Set(key, string(result), defaultExpiration)
	if err != nil {
		logger.Instance.Error("Ошибка сохранения в БД", "error", err)
	}

	logger.Instance.Debugw("Установлено значение в БД", "key", key, "value", value)
}

func Get(key string, result interface{}) bool {
	value, err := redis.GetClient().Get(key)

	if err != nil {
		logger.Instance.Debugw("Ключ не найден в БД:", "key", key, "error", err.Error())
		return false
	}

	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		logger.Instance.Error(err.Error())
	}

	logger.Instance.Debugw("Получение unmarshalled значения из БД", "key", key, "value", value)
	return true
}

func Delete(key string) {
	err := redis.GetClient().Delete(key)

	if err != nil {
		logger.Instance.Error("Ошибка удаления из Redis", "error", err.Error())
	}
}
