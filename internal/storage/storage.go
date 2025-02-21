package storage

import (
	"code-review-tg-bot/internal/logger"
	"encoding/json"
	"sync"
)

var (
	dataStore  = make(map[string]string)
	storeMutex sync.Mutex
)

func Set(key string, value interface{}) {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	result, err := json.Marshal(value)
	if err != nil {
		logger.Instance.Error(err.Error())
		panic(err)
	}

	dataStore[key] = string(result)

	logger.Instance.Debugw("STORAGE", "dataStore", dataStore)
}

func Get(key string, result interface{}) bool {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	value, exists := dataStore[key]

	logger.Instance.Debugw("GET RAW FROM STORAGE", "key", key, "exists", exists, "value", value)

	if exists {
		err := json.Unmarshal([]byte(value), &result)
		if err != nil {
			logger.Instance.Error(err.Error())
			panic(err)
		}
	}

	logger.Instance.Debugw("GET UNMARSHALLED FROM STORAGE", "key", key, "value", value)

	return exists
}

func Delete(key string) {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	delete(dataStore, key)
}

func Show() {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	logger.Instance.Debugw("Store contents:")
	for key, value := range dataStore {
		logger.Instance.Debugw("%s: %s\n", key, value)
	}
}
