package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleSetCommand implémente SET key value [EX seconds]
func (commandRegistry *RedisCommandRegistry) handleSetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SET' (attendu: SET clé valeur [EX secondes])")
	}

	storageKey := commandArguments[0]
	storageValue := commandArguments[1]

	// Parsing des options (EX pour TTL)
	for argumentIndex := 2; argumentIndex < len(commandArguments); argumentIndex++ {
		switch strings.ToUpper(commandArguments[argumentIndex]) {
		case "EX":
			if argumentIndex+1 >= len(commandArguments) {
				return protocolEncoder.WriteErrorResponse("ERREUR : valeur manquante après 'EX'")
			}
			expirationSeconds, parseError := strconv.Atoi(commandArguments[argumentIndex+1])
			if parseError != nil {
				return protocolEncoder.WriteErrorResponse("ERREUR : la valeur après 'EX' doit être un nombre entier")
			}
			if expirationSeconds <= 0 {
				return protocolEncoder.WriteErrorResponse("ERREUR : le délai d'expiration doit être positif")
			}
			timeToLive := time.Duration(expirationSeconds) * time.Second
			redisStorage.SetKeyValue(storageKey, storageValue, storage.RedisStringType, &timeToLive)
			return protocolEncoder.WriteSimpleStringResponse("OK")
		default:
			return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : option inconnue '%s' pour SET", commandArguments[argumentIndex]))
		}
	}

	// SET sans TTL
	redisStorage.SetKeyValue(storageKey, storageValue, storage.RedisStringType, nil)
	return protocolEncoder.WriteSimpleStringResponse("OK")
}

// handleGetCommand implémente GET key
func (commandRegistry *RedisCommandRegistry) handleGetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'GET' (attendu: GET clé)")
	}

	storageKey := commandArguments[0]
	storageValue := redisStorage.GetKeyValue(storageKey)

	if storageValue == nil {
		return protocolEncoder.WriteBulkStringResponse("(nil)")
	}

	if storageValue.DataType != storage.RedisStringType {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
	}

	return protocolEncoder.WriteBulkStringResponse(storageValue.StoredData.(string))
}

// handleDeleteCommand implémente DEL key [key ...]
func (commandRegistry *RedisCommandRegistry) handleDeleteCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'DEL' (attendu: DEL clé [clé ...])")
	}

	deletedKeyCount := int64(0)
	for _, keyToDelete := range commandArguments {
		if redisStorage.DeleteKeyValue(keyToDelete) {
			deletedKeyCount++
		}
	}

	return protocolEncoder.WriteIntegerResponse(deletedKeyCount)
}

// handleExistsCommand implémente EXISTS key [key ...]
func (commandRegistry *RedisCommandRegistry) handleExistsCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'EXISTS' (attendu: EXISTS clé [clé ...])")
	}

	existingKeyCount := int64(0)
	for _, keyToCheck := range commandArguments {
		if redisStorage.CheckKeyExists(keyToCheck) {
			existingKeyCount++
		}
	}

	return protocolEncoder.WriteIntegerResponse(existingKeyCount)
}

// handleKeysCommand implémente KEYS <pattern>
func (commandRegistry *RedisCommandRegistry) handleKeysCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'KEYS' (attendu: KEYS motif)")
	}

	searchPattern := commandArguments[0]
	matchingKeys := redisStorage.FindKeysByPattern(searchPattern)

	// Si aucune clé trouvée, afficher un message explicite
	if len(matchingKeys) == 0 {
		return protocolEncoder.WriteBulkStringResponse("(empty list or set)")
	}

	// Si des clés trouvées, utiliser WriteArray pour avoir les numéros 1), 2), etc.
	return protocolEncoder.WriteArrayResponse(matchingKeys)
}

// handleTypeCommand implémente TYPE key
func (commandRegistry *RedisCommandRegistry) handleTypeCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'TYPE' (attendu: TYPE clé)")
	}

	storageKey := commandArguments[0]
	keyDataType := redisStorage.GetKeyDataType(storageKey)

	var dataTypeString string
	switch keyDataType {
	case storage.RedisStringType:
		dataTypeString = "string"
	case storage.RedisListType:
		dataTypeString = "list"
	case storage.RedisSetType:
		dataTypeString = "set"
	case storage.RedisHashType:
		dataTypeString = "hash"
	case storage.RedisZSetType:
		dataTypeString = "zset"
	default:
		dataTypeString = "none"
	}

	return protocolEncoder.WriteSimpleStringResponse(dataTypeString)
}
