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
		logger.Instance.Error(err.Error())
		panic(err)
	}

	err = redis.GetClient().Set(key, string(result), defaultExpiration)
	if err != nil {
		logger.Instance.Error("Ошибка сохранения в Redis", "error", err)
		panic(err)
	}

	logger.Instance.Debugw("STORAGE SET", "key", key, "value", value)
}

func Get(key string, result interface{}) bool {
	value, err := redis.GetClient().Get(key)

	if err != nil {
		logger.Instance.Debugw("GET RAW FROM STORAGE", "key", key, "error", err.Error())
		return false
	}

	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		logger.Instance.Error(err.Error())
		panic(err)
	}

	logger.Instance.Debugw("GET UNMARSHALLED FROM STORAGE", "key", key, "value", value)
	return true
}

func Delete(key string) {
	err := redis.GetClient().Delete(key)

	if err != nil {
		logger.Instance.Error("Ошибка удаления из Redis", "error", err.Error())
	}
}
