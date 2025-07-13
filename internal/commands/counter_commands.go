package commands

import (
	"strconv"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleIncrementCommand implémente INCR key
func (commandRegistry *RedisCommandRegistry) handleIncrementCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'INCR' (attendu: INCR clé)")
	}

	counterKey := commandArguments[0]
	storageValue := redisStorage.GetKeyValue(counterKey)

	var currentCounterValue int64 = 0
	if storageValue != nil {
		if storageValue.DataType != storage.RedisStringType {
			return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}

		var parseError error
		currentCounterValue, parseError = strconv.ParseInt(storageValue.StoredData.(string), 10, 64)
		if parseError != nil {
			return protocolEncoder.WriteErrorResponse("ERREUR : la valeur n'est pas un nombre entier")
		}
	}

	currentCounterValue++
	redisStorage.SetKeyValue(counterKey, strconv.FormatInt(currentCounterValue, 10), storage.RedisStringType, nil)
	return protocolEncoder.WriteIntegerResponse(currentCounterValue)
}

// handleDecrementCommand implémente DECR key
func (commandRegistry *RedisCommandRegistry) handleDecrementCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'DECR' (attendu: DECR clé)")
	}

	counterKey := commandArguments[0]
	storageValue := redisStorage.GetKeyValue(counterKey)

	var currentCounterValue int64 = 0
	if storageValue != nil {
		if storageValue.DataType != storage.RedisStringType {
			return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}

		var parseError error
		currentCounterValue, parseError = strconv.ParseInt(storageValue.StoredData.(string), 10, 64)
		if parseError != nil {
			return protocolEncoder.WriteErrorResponse("ERREUR : la valeur n'est pas un nombre entier")
		}
	}

	currentCounterValue--
	redisStorage.SetKeyValue(counterKey, strconv.FormatInt(currentCounterValue, 10), storage.RedisStringType, nil)
	return protocolEncoder.WriteIntegerResponse(currentCounterValue)
}

// handleIncrementByCommand implémente INCRBY key increment
func (commandRegistry *RedisCommandRegistry) handleIncrementByCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'INCRBY' (attendu: INCRBY clé incrément)")
	}

	counterKey := commandArguments[0]
	incrementValue, parseError := strconv.ParseInt(commandArguments[1], 10, 64)
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'incrément doit être un nombre entier")
	}

	storageValue := redisStorage.GetKeyValue(counterKey)
	var currentCounterValue int64 = 0
	if storageValue != nil {
		if storageValue.DataType != storage.RedisStringType {
			return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}

		currentCounterValue, parseError = strconv.ParseInt(storageValue.StoredData.(string), 10, 64)
		if parseError != nil {
			return protocolEncoder.WriteErrorResponse("ERREUR : la valeur n'est pas un nombre entier")
		}
	}

	currentCounterValue += incrementValue
	redisStorage.SetKeyValue(counterKey, strconv.FormatInt(currentCounterValue, 10), storage.RedisStringType, nil)
	return protocolEncoder.WriteIntegerResponse(currentCounterValue)
}

// handleDecrementByCommand implémente DECRBY key decrement
func (commandRegistry *RedisCommandRegistry) handleDecrementByCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'DECRBY' (attendu: DECRBY clé décrément)")
	}

	counterKey := commandArguments[0]
	decrementValue, parseError := strconv.ParseInt(commandArguments[1], 10, 64)
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : le décrément doit être un nombre entier")
	}

	storageValue := redisStorage.GetKeyValue(counterKey)
	var currentCounterValue int64 = 0
	if storageValue != nil {
		if storageValue.DataType != storage.RedisStringType {
			return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}

		currentCounterValue, parseError = strconv.ParseInt(storageValue.StoredData.(string), 10, 64)
		if parseError != nil {
			return protocolEncoder.WriteErrorResponse("ERREUR : la valeur n'est pas un nombre entier")
		}
	}

	currentCounterValue -= decrementValue
	redisStorage.SetKeyValue(counterKey, strconv.FormatInt(currentCounterValue, 10), storage.RedisStringType, nil)
	return protocolEncoder.WriteIntegerResponse(currentCounterValue)
}
