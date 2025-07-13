package commands

import (
	"strconv"
	"strings"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleHashSetCommand implémente HSET key field value [field value ...]
func (commandRegistry *RedisCommandRegistry) handleHashSetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 3 || len(commandArguments)%2 == 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HSET' (attendu: HSET clé champ valeur [champ valeur ...])")
	}

	hashKey := commandArguments[0]
	newFieldCount := int64(0)

	// Traiter les paires field/value
	for argumentIndex := 1; argumentIndex < len(commandArguments); argumentIndex += 2 {
		fieldName := commandArguments[argumentIndex]
		fieldValue := commandArguments[argumentIndex+1]

		isNewField := redisStorage.SetHashField(hashKey, fieldName, fieldValue)
		if isNewField {
			newFieldCount++
		}
	}

	return protocolEncoder.WriteIntegerResponse(newFieldCount)
}

// handleHashGetCommand implémente HGET key field
func (commandRegistry *RedisCommandRegistry) handleHashGetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HGET' (attendu: HGET clé champ)")
	}

	hashKey := commandArguments[0]
	fieldName := commandArguments[1]

	fieldValue, fieldExists := redisStorage.GetHashField(hashKey, fieldName)
	if !fieldExists {
		return protocolEncoder.WriteBulkStringResponse("(nil)")
	}

	return protocolEncoder.WriteBulkStringResponse(fieldValue)
}

// handleHashGetAllCommand implémente HGETALL key
func (commandRegistry *RedisCommandRegistry) handleHashGetAllCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HGETALL' (attendu: HGETALL clé)")
	}

	hashKey := commandArguments[0]
	hashFields := redisStorage.GetAllHashFields(hashKey)
	if hashFields == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un hash")
	}

	// Convertir en array alternant field/value
	responseArray := make([]string, 0, len(hashFields)*2)
	for fieldName, fieldValue := range hashFields {
		responseArray = append(responseArray, fieldName, fieldValue)
	}

	return protocolEncoder.WriteArrayResponse(responseArray)
}

// handleHashExistsCommand implémente HEXISTS key field
func (commandRegistry *RedisCommandRegistry) handleHashExistsCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HEXISTS' (attendu: HEXISTS clé champ)")
	}

	hashKey := commandArguments[0]
	fieldName := commandArguments[1]

	_, fieldExists := redisStorage.GetHashField(hashKey, fieldName)
	if fieldExists {
		return protocolEncoder.WriteIntegerResponse(1)
	}
	return protocolEncoder.WriteIntegerResponse(0)
}

// handleHashDeleteCommand implémente HDEL key field [field ...]
func (commandRegistry *RedisCommandRegistry) handleHashDeleteCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HDEL' (attendu: HDEL clé champ [champ ...])")
	}

	hashKey := commandArguments[0]
	fieldsToDelete := commandArguments[1:]

	deletedFieldCount := redisStorage.DeleteHashFields(hashKey, fieldsToDelete)
	if deletedFieldCount == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un hash")
	}

	return protocolEncoder.WriteIntegerResponse(int64(deletedFieldCount))
}

// handleHashLengthCommand implémente HLEN key
func (commandRegistry *RedisCommandRegistry) handleHashLengthCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HLEN' (attendu: HLEN clé)")
	}

	hashKey := commandArguments[0]
	hashLength := redisStorage.GetHashLength(hashKey)
	if hashLength == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un hash")
	}

	return protocolEncoder.WriteIntegerResponse(int64(hashLength))
}

// handleHashKeysCommand implémente HKEYS key
func (commandRegistry *RedisCommandRegistry) handleHashKeysCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HKEYS' (attendu: HKEYS clé)")
	}

	hashKey := commandArguments[0]
	hashKeys := redisStorage.GetHashKeys(hashKey)
	if hashKeys == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un hash")
	}

	return protocolEncoder.WriteArrayResponse(hashKeys)
}

// handleHashValuesCommand implémente HVALS key
func (commandRegistry *RedisCommandRegistry) handleHashValuesCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HVALS' (attendu: HVALS clé)")
	}

	hashKey := commandArguments[0]
	hashValues := redisStorage.GetHashValues(hashKey)
	if hashValues == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un hash")
	}

	return protocolEncoder.WriteArrayResponse(hashValues)
}

// handleHashIncrementByCommand implémente HINCRBY key field increment
func (commandRegistry *RedisCommandRegistry) handleHashIncrementByCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HINCRBY' (attendu: HINCRBY clé champ incrément)")
	}

	hashKey := commandArguments[0]
	fieldName := commandArguments[1]
	incrementValue, parseError := strconv.ParseInt(commandArguments[2], 10, 64)
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'incrément doit être un nombre entier")
	}

	newValue := redisStorage.IncrementHashField(hashKey, fieldName, incrementValue)
	if newValue == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un hash ou le champ n'est pas un nombre")
	}

	return protocolEncoder.WriteIntegerResponse(*newValue)
}

// handleHashIncrementByFloatCommand implémente HINCRBYFLOAT key field increment
func (commandRegistry *RedisCommandRegistry) handleHashIncrementByFloatCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HINCRBYFLOAT' (attendu: HINCRBYFLOAT clé champ incrément)")
	}

	hashKey := commandArguments[0]
	fieldName := commandArguments[1]
	incrementValue, parseError := strconv.ParseFloat(commandArguments[2], 64)
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'incrément doit être un nombre flottant")
	}

	newValue := redisStorage.IncrementHashFieldFloat(hashKey, fieldName, incrementValue)
	if newValue == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un hash ou le champ n'est pas un nombre")
	}

	// Formatter le float pour éviter la notation scientifique
	resultString := strconv.FormatFloat(*newValue, 'f', -1, 64)
	return protocolEncoder.WriteBulkStringResponse(resultString)
}
