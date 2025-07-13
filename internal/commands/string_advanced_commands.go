package commands

import (
	"strconv"
	"strings"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleAppendCommand implémente APPEND key value
func (commandRegistry *RedisCommandRegistry) handleAppendCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'APPEND' (attendu: APPEND clé valeur)")
	}

	storageKey := commandArguments[0]
	valueToAppend := commandArguments[1]

	// Récupérer la valeur existante ou créer une nouvelle string
	existingValue := redisStorage.GetKeyValue(storageKey)
	var finalValue string

	if existingValue == nil {
		// Clé n'existe pas, créer avec la valeur à ajouter
		finalValue = valueToAppend
	} else {
		// Vérifier que c'est bien une string
		if existingValue.DataType != storage.RedisStringType {
			return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}
		// Concaténer avec la valeur existante
		finalValue = existingValue.StoredData.(string) + valueToAppend
	}

	// Stocker la valeur concaténée
	redisStorage.SetKeyValue(storageKey, finalValue, storage.RedisStringType, nil)

	// Retourner la nouvelle longueur
	return protocolEncoder.WriteIntegerResponse(int64(len(finalValue)))
}

// handleStringLengthCommand implémente STRLEN key
func (commandRegistry *RedisCommandRegistry) handleStringLengthCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'STRLEN' (attendu: STRLEN clé)")
	}

	storageKey := commandArguments[0]
	storageValue := redisStorage.GetKeyValue(storageKey)

	if storageValue == nil {
		return protocolEncoder.WriteIntegerResponse(0)
	}

	if storageValue.DataType != storage.RedisStringType {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
	}

	stringValue := storageValue.StoredData.(string)
	return protocolEncoder.WriteIntegerResponse(int64(len(stringValue)))
}

// handleGetRangeCommand implémente GETRANGE key start end
func (commandRegistry *RedisCommandRegistry) handleGetRangeCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'GETRANGE' (attendu: GETRANGE clé début fin)")
	}

	storageKey := commandArguments[0]
	startIndex, parseError := strconv.Atoi(commandArguments[1])
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'index de début doit être un nombre entier")
	}

	endIndex, parseError := strconv.Atoi(commandArguments[2])
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'index de fin doit être un nombre entier")
	}

	storageValue := redisStorage.GetKeyValue(storageKey)
	if storageValue == nil {
		return protocolEncoder.WriteBulkStringResponse("")
	}

	if storageValue.DataType != storage.RedisStringType {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
	}

	stringValue := storageValue.StoredData.(string)
	stringLength := len(stringValue)

	// Gérer les indices négatifs (comme Redis)
	if startIndex < 0 {
		startIndex = stringLength + startIndex
	}
	if endIndex < 0 {
		endIndex = stringLength + endIndex
	}

	// Limiter aux bornes
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex >= stringLength {
		endIndex = stringLength - 1
	}
	if startIndex > endIndex || stringLength == 0 {
		return protocolEncoder.WriteBulkStringResponse("")
	}

	substring := stringValue[startIndex : endIndex+1]
	return protocolEncoder.WriteBulkStringResponse(substring)
}

// handleSetRangeCommand implémente SETRANGE key offset value
func (commandRegistry *RedisCommandRegistry) handleSetRangeCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SETRANGE' (attendu: SETRANGE clé offset valeur)")
	}

	storageKey := commandArguments[0]
	offset, parseError := strconv.Atoi(commandArguments[1])
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'offset doit être un nombre entier")
	}

	if offset < 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'offset ne peut pas être négatif")
	}

	newValue := commandArguments[2]
	storageValue := redisStorage.GetKeyValue(storageKey)

	var currentString string
	if storageValue == nil {
		// Clé n'existe pas, créer une string vide
		currentString = ""
	} else {
		if storageValue.DataType != storage.RedisStringType {
			return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}
		currentString = storageValue.StoredData.(string)
	}

	// Étendre la string avec des zéros si nécessaire
	requiredLength := offset + len(newValue)
	if len(currentString) < requiredLength {
		currentString += strings.Repeat("\x00", requiredLength-len(currentString))
	}

	// Remplacer la partie désignée
	stringBytes := []byte(currentString)
	copy(stringBytes[offset:], []byte(newValue))
	finalString := string(stringBytes)

	// Stocker la nouvelle valeur
	redisStorage.SetKeyValue(storageKey, finalString, storage.RedisStringType, nil)

	return protocolEncoder.WriteIntegerResponse(int64(len(finalString)))
}

// handleMultiSetCommand implémente MSET key value [key value ...]
func (commandRegistry *RedisCommandRegistry) handleMultiSetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 || len(commandArguments)%2 != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'MSET' (attendu: MSET clé valeur [clé valeur ...])")
	}

	// Traiter toutes les paires clé/valeur
	for argumentIndex := 0; argumentIndex < len(commandArguments); argumentIndex += 2 {
		storageKey := commandArguments[argumentIndex]
		storageValue := commandArguments[argumentIndex+1]
		redisStorage.SetKeyValue(storageKey, storageValue, storage.RedisStringType, nil)
	}

	return protocolEncoder.WriteSimpleStringResponse("OK")
}

// handleMultiGetCommand implémente MGET key [key ...]
func (commandRegistry *RedisCommandRegistry) handleMultiGetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'MGET' (attendu: MGET clé [clé ...])")
	}

	responseValues := make([]string, len(commandArguments))

	for keyIndex, storageKey := range commandArguments {
		storageValue := redisStorage.GetKeyValue(storageKey)
		if storageValue == nil || storageValue.DataType != storage.RedisStringType {
			responseValues[keyIndex] = "(nil)"
		} else {
			responseValues[keyIndex] = storageValue.StoredData.(string)
		}
	}

	return protocolEncoder.WriteArrayResponse(responseValues)
}

// handleGetSetCommand implémente GETSET key value - atomique GET + SET
func (commandRegistry *RedisCommandRegistry) handleGetSetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'GETSET' (attendu: GETSET clé valeur)")
	}

	storageKey := commandArguments[0]
	newValue := commandArguments[1]

	// Récupérer l'ancienne valeur
	oldValue := redisStorage.GetKeyValue(storageKey)
	var oldStringValue string
	var hasOldValue bool

	if oldValue != nil && oldValue.DataType == storage.RedisStringType {
		oldStringValue = oldValue.StoredData.(string)
		hasOldValue = true
	}

	// Définir la nouvelle valeur
	redisStorage.SetKeyValue(storageKey, newValue, storage.RedisStringType, nil)

	// Retourner l'ancienne valeur
	if hasOldValue {
		return protocolEncoder.WriteBulkStringResponse(oldStringValue)
	}
	return protocolEncoder.WriteBulkStringResponse("(nil)")
}

// handleMultiSetNxCommand implémente MSETNX key value [key value ...] - Multi-set si aucune clé existe
func (commandRegistry *RedisCommandRegistry) handleMultiSetNxCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 || len(commandArguments)%2 != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'MSETNX' (attendu: MSETNX clé valeur [clé valeur ...])")
	}

	// Vérifier d'abord que TOUTES les clés n'existent pas
	for argumentIndex := 0; argumentIndex < len(commandArguments); argumentIndex += 2 {
		storageKey := commandArguments[argumentIndex]
		if redisStorage.CheckKeyExists(storageKey) {
			// Si au moins une clé existe, retourner 0 sans rien faire
			return protocolEncoder.WriteIntegerResponse(0)
		}
	}

	// Si aucune clé n'existe, définir toutes les paires
	for argumentIndex := 0; argumentIndex < len(commandArguments); argumentIndex += 2 {
		storageKey := commandArguments[argumentIndex]
		storageValue := commandArguments[argumentIndex+1]
		redisStorage.SetKeyValue(storageKey, storageValue, storage.RedisStringType, nil)
	}

	return protocolEncoder.WriteIntegerResponse(1) // Toutes les clés ont été définies
}

// handleGetDelCommand implémente GETDEL key - GET puis DELETE atomique
func (commandRegistry *RedisCommandRegistry) handleGetDelCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'GETDEL' (attendu: GETDEL clé)")
	}

	storageKey := commandArguments[0]

	// Récupérer la valeur
	storageValue := redisStorage.GetKeyValue(storageKey)
	if storageValue == nil {
		return protocolEncoder.WriteBulkStringResponse("(nil)")
	}

	if storageValue.DataType != storage.RedisStringType {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une chaîne de caractères")
	}

	stringValue := storageValue.StoredData.(string)

	// Supprimer la clé
	redisStorage.DeleteKeyValue(storageKey)

	// Retourner l'ancienne valeur
	return protocolEncoder.WriteBulkStringResponse(stringValue)
}
