package storage

type Storage interface {
	Set(key string, value interface{})
	Get(key string, result interface{}) bool
	Delete(key string)
}
