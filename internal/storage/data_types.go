package storage

import "time"

// RedisDataType représente le type de données stocké
type RedisDataType int

const (
	RedisStringType RedisDataType = iota
	RedisListType
	RedisSetType
	RedisHashType
	RedisZSetType
)

// RedisStorageValue représente une valeur stockée avec son type et TTL
type RedisStorageValue struct {
	StoredData     interface{}
	DataType       RedisDataType
	ExpirationTime *time.Time
}

// RedisListStructure représente une liste Redis
type RedisListStructure struct {
	ListElements []string
}

// RedisSetStructure représente un set Redis
type RedisSetStructure struct {
	SetElements map[string]bool
}

// RedisHashStructure représente un hash Redis
type RedisHashStructure struct {
	HashFields map[string]string
}
