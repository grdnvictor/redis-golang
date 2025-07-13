package commands

import (
	"strconv"
	"time"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleSetNxCommand implémente SETNX key value
func (commandRegistry *RedisCommandRegistry) handleSetNxCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SETNX' (attendu: SETNX clé valeur)")
	}

	storageKey := commandArguments[0]
	storageValue := commandArguments[1]

	// Utiliser la nouvelle méthode SetKeyValueIfNotExists
	keyWasCreated := redisStorage.SetKeyValueIfNotExists(storageKey, storageValue, storage.RedisStringType)

	if keyWasCreated {
		return protocolEncoder.WriteIntegerResponse(1) // Clé créée
	}
	return protocolEncoder.WriteIntegerResponse(0) // Clé existe déjà
}

// handleSetExCommand implémente SETEX key seconds value
func (commandRegistry *RedisCommandRegistry) handleSetExCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SETEX' (attendu: SETEX clé secondes valeur)")
	}

	storageKey := commandArguments[0]
	secondsString := commandArguments[1]
	storageValue := commandArguments[2]

	// Valider et convertir les secondes
	expirationSeconds, parseError := strconv.Atoi(secondsString)
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : la valeur des secondes doit être un nombre entier")
	}

	if expirationSeconds <= 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : le délai d'expiration doit être positif")
	}

	// Créer la durée TTL
	timeToLive := time.Duration(expirationSeconds) * time.Second

	// Utiliser la méthode existante avec TTL
	redisStorage.SetKeyValue(storageKey, storageValue, storage.RedisStringType, &timeToLive)

	return protocolEncoder.WriteSimpleStringResponse("OK")
}
