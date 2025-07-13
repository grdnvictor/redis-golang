package protocol

// Constantes pour le protocole RESP (Redis Serialization Protocol)
const (
	RedisSimpleStringType = '+'
	RedisErrorType        = '-'
	RedisIntegerType      = ':'
	RedisBulkStringType   = '$'
	RedisArrayType        = '*'
)
