package commands

import (
	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleSetAddCommand implémente SADD key member [member ...]
func (commandRegistry *RedisCommandRegistry) handleSetAddCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SADD' (attendu: SADD clé membre [membre ...])")
	}

	setKey := commandArguments[0]
	membersToAdd := commandArguments[1:]

	addedMemberCount := redisStorage.AddMembersToSet(setKey, membersToAdd)
	if addedMemberCount == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un ensemble")
	}

	return protocolEncoder.WriteIntegerResponse(int64(addedMemberCount))
}

// handleSetMembersCommand implémente SMEMBERS key
func (commandRegistry *RedisCommandRegistry) handleSetMembersCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SMEMBERS' (attendu: SMEMBERS clé)")
	}

	setKey := commandArguments[0]
	setMembers := redisStorage.GetAllSetMembers(setKey)
	if setMembers == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un ensemble")
	}

	return protocolEncoder.WriteArrayResponse(setMembers)
}

// handleSetIsMemberCommand implémente SISMEMBER key member
func (commandRegistry *RedisCommandRegistry) handleSetIsMemberCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SISMEMBER' (attendu: SISMEMBER clé membre)")
	}

	setKey := commandArguments[0]
	memberToCheck := commandArguments[1]

	memberExists := redisStorage.CheckSetMemberExists(setKey, memberToCheck)
	if memberExists {
		return protocolEncoder.WriteIntegerResponse(1)
	}
	return protocolEncoder.WriteIntegerResponse(0)
}

// handleSetRemoveCommand implémente SREM key member [member ...]
func (commandRegistry *RedisCommandRegistry) handleSetRemoveCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SREM' (attendu: SREM clé membre [membre ...])")
	}

	setKey := commandArguments[0]
	membersToRemove := commandArguments[1:]

	removedMemberCount := redisStorage.RemoveMembersFromSet(setKey, membersToRemove)
	if removedMemberCount == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un ensemble")
	}

	return protocolEncoder.WriteIntegerResponse(int64(removedMemberCount))
}

// handleSetCardinalityCommand implémente SCARD key
func (commandRegistry *RedisCommandRegistry) handleSetCardinalityCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SCARD' (attendu: SCARD clé)")
	}

	setKey := commandArguments[0]
	setCardinality := redisStorage.GetSetCardinality(setKey)
	if setCardinality == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un ensemble")
	}

	return protocolEncoder.WriteIntegerResponse(int64(setCardinality))
}

// handleSetDifferenceCommand implémente SDIFF key [key ...]
func (commandRegistry *RedisCommandRegistry) handleSetDifferenceCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SDIFF' (attendu: SDIFF clé [clé ...])")
	}

	differenceMembers := redisStorage.ComputeSetDifference(commandArguments)
	if differenceMembers == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : au moins une clé ne contient pas un ensemble")
	}

	return protocolEncoder.WriteArrayResponse(differenceMembers)
}

// handleSetIntersectionCommand implémente SINTER key [key ...]
func (commandRegistry *RedisCommandRegistry) handleSetIntersectionCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SINTER' (attendu: SINTER clé [clé ...])")
	}

	intersectionMembers := redisStorage.ComputeSetIntersection(commandArguments)
	if intersectionMembers == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : au moins une clé ne contient pas un ensemble")
	}

	return protocolEncoder.WriteArrayResponse(intersectionMembers)
}

// handleSetUnionCommand implémente SUNION key [key ...]
func (commandRegistry *RedisCommandRegistry) handleSetUnionCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SUNION' (attendu: SUNION clé [clé ...])")
	}

	unionMembers := redisStorage.ComputeSetUnion(commandArguments)
	if unionMembers == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : au moins une clé ne contient pas un ensemble")
	}

	return protocolEncoder.WriteArrayResponse(unionMembers)
}
