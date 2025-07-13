package commands

import (
	"fmt"
	"strings"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// RedisCommandHandler représente une fonction qui traite une commande Redis
type RedisCommandHandler func(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error

// RedisCommandRegistry contient toutes les commandes supportées
type RedisCommandRegistry struct {
	registeredCommands map[string]RedisCommandHandler
}

// NewRedisCommandRegistry crée un nouveau registre de commandes
func NewRedisCommandRegistry() *RedisCommandRegistry {
	commandRegistry := &RedisCommandRegistry{
		registeredCommands: make(map[string]RedisCommandHandler),
	}

	// Enregistrement des commandes
	commandRegistry.registerAllCommands()

	return commandRegistry
}

// registerAllCommands enregistre toutes les commandes supportées
func (commandRegistry *RedisCommandRegistry) registerAllCommands() {
	commands := map[string]RedisCommandHandler{
		// Commandes String basiques
		"SET":    commandRegistry.handleSetCommand,
		"GET":    commandRegistry.handleGetCommand,
		"DEL":    commandRegistry.handleDeleteCommand,
		"EXISTS": commandRegistry.handleExistsCommand,
		"KEYS":   commandRegistry.handleKeysCommand,
		"TYPE":   commandRegistry.handleTypeCommand,
		"INCR":   commandRegistry.handleIncrementCommand,
		"DECR":   commandRegistry.handleDecrementCommand,
		"INCRBY": commandRegistry.handleIncrementByCommand,
		"DECRBY": commandRegistry.handleDecrementByCommand,
		// =====================================
		"SETNX": commandRegistry.handleSetNxCommand, // SET if Not eXists
		"SETEX": commandRegistry.handleSetExCommand, // SET with EXpiration
		// =====================================

		// Commandes String avancées
		"APPEND":   commandRegistry.handleAppendCommand,
		"STRLEN":   commandRegistry.handleStringLengthCommand,
		"GETRANGE": commandRegistry.handleGetRangeCommand,
		"SUBSTR":   commandRegistry.handleGetRangeCommand, // Alias pour GETRANGE
		"SETRANGE": commandRegistry.handleSetRangeCommand,
		"MSET":     commandRegistry.handleMultiSetCommand,
		"MGET":     commandRegistry.handleMultiGetCommand,

		// Nouvelles commandes String atomiques
		"GETSET": commandRegistry.handleGetSetCommand,     // Get ancien + Set nouveau
		"MSETNX": commandRegistry.handleMultiSetNxCommand, // Multi-set si aucune clé existe
		"GETDEL": commandRegistry.handleGetDelCommand,     // Get puis delete atomique

		// Commandes TTL
		"TTL":     commandRegistry.handleTtlCommand,
		"PTTL":    commandRegistry.handlePttlCommand,
		"EXPIRE":  commandRegistry.handleExpireCommand,
		"PEXPIRE": commandRegistry.handlePexpireCommand,
		"PERSIST": commandRegistry.handlePersistCommand,

		// Commandes List
		"LPUSH":  commandRegistry.handleLeftPushCommand,
		"RPUSH":  commandRegistry.handleRightPushCommand,
		"LPOP":   commandRegistry.handleLeftPopCommand,
		"RPOP":   commandRegistry.handleRightPopCommand,
		"LLEN":   commandRegistry.handleListLengthCommand,
		"LRANGE": commandRegistry.handleListRangeCommand,

		// Nouvelles commandes List avancées
		"LSET":    commandRegistry.handleListSetCommand,    // Set élément à index
		"LREM":    commandRegistry.handleListRemoveCommand, // Remove éléments
		"LINSERT": commandRegistry.handleListInsertCommand, // Insert avant/après
		"LTRIM":   commandRegistry.handleListTrimCommand,   // Trim liste

		// Commandes Set
		"SADD":      commandRegistry.handleSetAddCommand,
		"SMEMBERS":  commandRegistry.handleSetMembersCommand,
		"SISMEMBER": commandRegistry.handleSetIsMemberCommand,

		// Nouvelles commandes Set avancées
		"SREM":   commandRegistry.handleSetRemoveCommand,       // Remove membres
		"SCARD":  commandRegistry.handleSetCardinalityCommand,  // Nombre de membres
		"SDIFF":  commandRegistry.handleSetDifferenceCommand,   // Différence de sets
		"SINTER": commandRegistry.handleSetIntersectionCommand, // Intersection de sets
		"SUNION": commandRegistry.handleSetUnionCommand,        // Union de sets

		// Commandes Hash
		"HSET":    commandRegistry.handleHashSetCommand,
		"HGET":    commandRegistry.handleHashGetCommand,
		"HGETALL": commandRegistry.handleHashGetAllCommand,

		// Nouvelles commandes Hash avancées
		"HEXISTS":      commandRegistry.handleHashExistsCommand,           // Field existe?
		"HDEL":         commandRegistry.handleHashDeleteCommand,           // Delete fields
		"HLEN":         commandRegistry.handleHashLengthCommand,           // Nombre de fields
		"HKEYS":        commandRegistry.handleHashKeysCommand,             // Tous les field names
		"HVALS":        commandRegistry.handleHashValuesCommand,           // Toutes les values
		"HINCRBY":      commandRegistry.handleHashIncrementByCommand,      // Incrément entier
		"HINCRBYFLOAT": commandRegistry.handleHashIncrementByFloatCommand, // Incrément float

		// Commandes utilitaires
		"PING":     commandRegistry.handlePingCommand,
		"ECHO":     commandRegistry.handleEchoCommand,
		"DBSIZE":   commandRegistry.handleDatabaseSizeCommand,
		"FLUSHALL": commandRegistry.handleFlushAllCommand,
		"ALAIDE":   commandRegistry.handleHelpCommand,
	}

	for commandName, handler := range commands {
		commandRegistry.registeredCommands[commandName] = handler
	}
}

// ExecuteCommand exécute une commande donnée
func (commandRegistry *RedisCommandRegistry) ExecuteCommand(commandName string, commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	upperCommandName := strings.ToUpper(commandName)
	commandHandler, commandExists := commandRegistry.registeredCommands[upperCommandName]

	if !commandExists {
		suggestion := commandRegistry.findSimilarCommand(upperCommandName)
		if suggestion != "" {
			return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : commande inconnue '%s'. Vouliez-vous dire '%s' ?", commandName, suggestion))
		}
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : commande inconnue '%s'", commandName))
	}

	return commandHandler(commandArguments, redisStorage, protocolEncoder)
}

// findSimilarCommand trouve la commande la plus similaire en utilisant la distance de Levenshtein
func (commandRegistry *RedisCommandRegistry) findSimilarCommand(input string) string {
	minDistance := 3 // Seuil de similarité
	bestMatch := ""

	for commandName := range commandRegistry.registeredCommands {
		distance := levenshteinDistance(input, commandName)
		if distance < minDistance {
			minDistance = distance
			bestMatch = commandName
		}
	}

	return bestMatch
}

// levenshteinDistance calcule la distance de Levenshtein entre deux chaînes
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = minimum(
				matrix[i-1][j]+1,      // suppression
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func minimum(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
