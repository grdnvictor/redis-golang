package commands

import (
	"strconv"
	"time"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleTtlCommand implémente TTL key (retourne TTL en secondes)
func (commandRegistry *RedisCommandRegistry) handleTtlCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'TTL' (attendu: TTL clé)")
	}

	storageKey := commandArguments[0]
	ttlSeconds := redisStorage.GetKeyTTL(storageKey, false) // false = secondes

	return protocolEncoder.WriteIntegerResponse(ttlSeconds)
}

// handlePttlCommand implémente PTTL key (retourne TTL en millisecondes)
func (commandRegistry *RedisCommandRegistry) handlePttlCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'PTTL' (attendu: PTTL clé)")
	}

	storageKey := commandArguments[0]
	ttlMilliseconds := redisStorage.GetKeyTTL(storageKey, true) // true = millisecondes

	return protocolEncoder.WriteIntegerResponse(ttlMilliseconds)
}

// handleExpireCommand implémente EXPIRE key seconds
func (commandRegistry *RedisCommandRegistry) handleExpireCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'EXPIRE' (attendu: EXPIRE clé secondes)")
	}

	storageKey := commandArguments[0]
	expirationSeconds, parseError := strconv.ParseInt(commandArguments[1], 10, 64)
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : la valeur des secondes doit être un nombre entier")
	}

	if expirationSeconds <= 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : le délai d'expiration doit être positif")
	}

	// Tenter de définir l'expiration
	success := redisStorage.SetKeyExpiration(storageKey, time.Duration(expirationSeconds)*time.Second)
	if success {
		return protocolEncoder.WriteIntegerResponse(1)
	}
	return protocolEncoder.WriteIntegerResponse(0) // Clé n'existe pas
}

// handlePexpireCommand implémente PEXPIRE key milliseconds
func (commandRegistry *RedisCommandRegistry) handlePexpireCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'PEXPIRE' (attendu: PEXPIRE clé millisecondes)")
	}

	storageKey := commandArguments[0]
	expirationMilliseconds, parseError := strconv.ParseInt(commandArguments[1], 10, 64)
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : la valeur des millisecondes doit être un nombre entier")
	}

	if expirationMilliseconds <= 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : le délai d'expiration doit être positif")
	}

	// Tenter de définir l'expiration
	success := redisStorage.SetKeyExpiration(storageKey, time.Duration(expirationMilliseconds)*time.Millisecond)
	if success {
		return protocolEncoder.WriteIntegerResponse(1)
	}
	return protocolEncoder.WriteIntegerResponse(0) // Clé n'existe pas
}

// handlePersistCommand implémente PERSIST key (supprime le TTL)
func (commandRegistry *RedisCommandRegistry) handlePersistCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'PERSIST' (attendu: PERSIST clé)")
	}

	storageKey := commandArguments[0]
	success := redisStorage.RemoveKeyExpiration(storageKey)

	if success {
		return protocolEncoder.WriteIntegerResponse(1) // TTL supprimé
	}
	return protocolEncoder.WriteIntegerResponse(0) // Clé n'existe pas ou n'avait pas de TTL
}
