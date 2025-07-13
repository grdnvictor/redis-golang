package commands

import (
	"strconv"
	"strings"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleLeftPushCommand implémente LPUSH key element [element ...]
func (commandRegistry *RedisCommandRegistry) handleLeftPushCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'LPUSH' (attendu: LPUSH clé élément [élément ...])")
	}

	listKey := commandArguments[0]
	elementsToAdd := commandArguments[1:]

	listLength := redisStorage.PushElementsToList(listKey, elementsToAdd, true) // true = left
	if listLength == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une liste")
	}

	return protocolEncoder.WriteIntegerResponse(int64(listLength))
}

// handleRightPushCommand implémente RPUSH key element [element ...]
func (commandRegistry *RedisCommandRegistry) handleRightPushCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'RPUSH' (attendu: RPUSH clé élément [élément ...])")
	}

	listKey := commandArguments[0]
	elementsToAdd := commandArguments[1:]

	listLength := redisStorage.PushElementsToList(listKey, elementsToAdd, false) // false = right
	if listLength == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une liste")
	}

	return protocolEncoder.WriteIntegerResponse(int64(listLength))
}

// handleLeftPopCommand implémente LPOP key
func (commandRegistry *RedisCommandRegistry) handleLeftPopCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'LPOP' (attendu: LPOP clé)")
	}

	listKey := commandArguments[0]
	poppedElement, elementExists := redisStorage.PopElementFromList(listKey, true) // true = left
	if !elementExists {
		return protocolEncoder.WriteBulkStringResponse("(nil)")
	}

	return protocolEncoder.WriteBulkStringResponse(poppedElement)
}

// handleRightPopCommand implémente RPOP key
func (commandRegistry *RedisCommandRegistry) handleRightPopCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'RPOP' (attendu: RPOP clé)")
	}

	listKey := commandArguments[0]
	poppedElement, elementExists := redisStorage.PopElementFromList(listKey, false) // false = right
	if !elementExists {
		return protocolEncoder.WriteBulkStringResponse("(nil)")
	}

	return protocolEncoder.WriteBulkStringResponse(poppedElement)
}

// handleListLengthCommand implémente LLEN key
func (commandRegistry *RedisCommandRegistry) handleListLengthCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'LLEN' (attendu: LLEN clé)")
	}

	listKey := commandArguments[0]
	listLength := redisStorage.GetListLength(listKey)
	if listLength == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une liste")
	}

	return protocolEncoder.WriteIntegerResponse(int64(listLength))
}

// handleListRangeCommand implémente LRANGE key start stop
func (commandRegistry *RedisCommandRegistry) handleListRangeCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'LRANGE' (attendu: LRANGE clé début fin)")
	}

	listKey := commandArguments[0]
	startIndex, parseError := strconv.Atoi(commandArguments[1])
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'index de début doit être un nombre entier")
	}

	stopIndex, parseError := strconv.Atoi(commandArguments[2])
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'index de fin doit être un nombre entier")
	}

	listElements := redisStorage.GetListElementsInRange(listKey, startIndex, stopIndex)
	if listElements == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une liste")
	}

	return protocolEncoder.WriteArrayResponse(listElements)
}

// handleListSetCommand implémente LSET key index element
func (commandRegistry *RedisCommandRegistry) handleListSetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'LSET' (attendu: LSET clé index élément)")
	}

	listKey := commandArguments[0]
	elementIndex, parseError := strconv.Atoi(commandArguments[1])
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'index doit être un nombre entier")
	}

	newElement := commandArguments[2]

	success := redisStorage.SetListElement(listKey, elementIndex, newElement)
	if success == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une liste")
	}
	if success == 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : index hors limites")
	}

	return protocolEncoder.WriteSimpleStringResponse("OK")
}

// handleListRemoveCommand implémente LREM key count element
func (commandRegistry *RedisCommandRegistry) handleListRemoveCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'LREM' (attendu: LREM clé count élément)")
	}

	listKey := commandArguments[0]
	removeCount, parseError := strconv.Atoi(commandArguments[1])
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : count doit être un nombre entier")
	}

	elementToRemove := commandArguments[2]

	removedCount := redisStorage.RemoveListElements(listKey, removeCount, elementToRemove)
	if removedCount == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une liste")
	}

	return protocolEncoder.WriteIntegerResponse(int64(removedCount))
}

// handleListInsertCommand implémente LINSERT key BEFORE|AFTER pivot element
func (commandRegistry *RedisCommandRegistry) handleListInsertCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 4 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'LINSERT' (attendu: LINSERT clé BEFORE|AFTER pivot élément)")
	}

	listKey := commandArguments[0]
	direction := strings.ToUpper(commandArguments[1])
	pivotElement := commandArguments[2]
	newElement := commandArguments[3]

	if direction != "BEFORE" && direction != "AFTER" {
		return protocolEncoder.WriteErrorResponse("ERREUR : direction doit être BEFORE ou AFTER")
	}

	insertBefore := direction == "BEFORE"
	resultLength := redisStorage.InsertIntoList(listKey, insertBefore, pivotElement, newElement)

	if resultLength == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une liste")
	}
	if resultLength == -2 {
		return protocolEncoder.WriteIntegerResponse(-1) // Pivot non trouvé
	}

	return protocolEncoder.WriteIntegerResponse(int64(resultLength))
}

// handleListTrimCommand implémente LTRIM key start stop
func (commandRegistry *RedisCommandRegistry) handleListTrimCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'LTRIM' (attendu: LTRIM clé début fin)")
	}

	listKey := commandArguments[0]
	startIndex, parseError := strconv.Atoi(commandArguments[1])
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'index de début doit être un nombre entier")
	}

	stopIndex, parseError := strconv.Atoi(commandArguments[2])
	if parseError != nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : l'index de fin doit être un nombre entier")
	}

	success := redisStorage.TrimList(listKey, startIndex, stopIndex)
	if success == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas une liste")
	}

	return protocolEncoder.WriteSimpleStringResponse("OK")
}
