package commands

import (
	"fmt"
	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// RDBPersistenceInterface définit l'interface pour les commandes RDB
type RDBPersistenceInterface interface {
	Save() error
	BackgroundSave() error
	IsSaveInProgress() bool
	GetLastSaveTime() int64
	GetStats() map[string]interface{}
}

// rdbPersistence stocke la référence vers le système RDB
var rdbPersistence RDBPersistenceInterface

// SetRDBPersistence configure la persistence RDB pour les commandes
func (commandRegistry *RedisCommandRegistry) SetRDBPersistence(rdb RDBPersistenceInterface) {
	rdbPersistence = rdb

	// Ajouter les commandes RDB au registre existant
	commandRegistry.registeredCommands["SAVE"] = commandRegistry.handleSaveCommand
	commandRegistry.registeredCommands["BGSAVE"] = commandRegistry.handleBackgroundSaveCommand
	commandRegistry.registeredCommands["LASTSAVE"] = commandRegistry.handleLastSaveCommand
	commandRegistry.registeredCommands["INFO"] = commandRegistry.handleInfoCommand
}

// handleSaveCommand implémente SAVE (sauvegarde synchrone bloquante)
func (commandRegistry *RedisCommandRegistry) handleSaveCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : SAVE ne prend aucun argument")
	}

	if rdbPersistence == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : persistence RDB non configurée")
	}

	if err := rdbPersistence.Save(); err != nil {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : sauvegarde échouée: %v", err))
	}

	return protocolEncoder.WriteSimpleStringResponse("OK")
}

// handleBackgroundSaveCommand implémente BGSAVE (sauvegarde en arrière-plan)
func (commandRegistry *RedisCommandRegistry) handleBackgroundSaveCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : BGSAVE ne prend aucun argument")
	}

	if rdbPersistence == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : persistence RDB non configurée")
	}

	// Vérifier si une sauvegarde est déjà en cours
	if rdbPersistence.IsSaveInProgress() {
		return protocolEncoder.WriteErrorResponse("ERREUR : sauvegarde en arrière-plan déjà en cours")
	}

	if err := rdbPersistence.BackgroundSave(); err != nil {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : impossible de démarrer BGSAVE: %v", err))
	}

	return protocolEncoder.WriteSimpleStringResponse("Background saving started")
}

// handleLastSaveCommand implémente LASTSAVE
func (commandRegistry *RedisCommandRegistry) handleLastSaveCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : LASTSAVE ne prend aucun argument")
	}

	if rdbPersistence == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : persistence RDB non configurée")
	}

	lastSaveTime := rdbPersistence.GetLastSaveTime()
	return protocolEncoder.WriteIntegerResponse(lastSaveTime)
}

// handleInfoCommand implémente INFO [section]
func (commandRegistry *RedisCommandRegistry) handleInfoCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) > 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'INFO' (attendu: INFO [section])")
	}

	section := "all"
	if len(commandArguments) == 1 {
		section = commandArguments[0]
	}

	var infoResponse string

	switch section {
	case "all", "server":
		infoResponse += "# Server\r\n"
		infoResponse += "redis_version:Redis-Go-1.0\r\n"
		infoResponse += "redis_mode:standalone\r\n"
		infoResponse += "uptime_in_seconds:unknown\r\n"
		infoResponse += "\r\n"
		fallthrough

	case "persistence":
		if rdbPersistence != nil {
			infoResponse += "# Persistence\r\n"
			stats := rdbPersistence.GetStats()

			for key, value := range stats {
				switch v := value.(type) {
				case int64:
					infoResponse += fmt.Sprintf("%s:%d\r\n", key, v)
				case bool:
					boolValue := 0
					if v {
						boolValue = 1
					}
					infoResponse += fmt.Sprintf("%s:%d\r\n", key, boolValue)
				case string:
					infoResponse += fmt.Sprintf("%s:%s\r\n", key, v)
				}
			}
			infoResponse += "\r\n"
		}
		fallthrough

	case "memory":
		if section == "memory" || section == "all" {
			infoResponse += "# Memory\r\n"
			infoResponse += fmt.Sprintf("used_memory_keys:%d\r\n", redisStorage.GetStorageSize())
			infoResponse += "\r\n"
		}

	default:
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : section INFO inconnue '%s'", section))
	}

	return protocolEncoder.WriteBulkStringResponse(infoResponse)
}
