package server

import (
	"net"
	"sync"

	"redis-go/internal/commands"
	"redis-go/internal/config"
	"redis-go/internal/persistence"
	"redis-go/internal/storage"
)

// RedisServerInstance représente le serveur Redis
type RedisServerInstance struct {
	serverConfiguration *config.ServerConfiguration
	redisStorage        *storage.RedisInMemoryStorage
	commandRegistry     *commands.RedisCommandRegistry
	rdbPersistence      *persistence.RDBPersistence // Nouveau
	networkListener     net.Listener
	connectedClients    map[net.Conn]bool
	clientsMutex        sync.RWMutex
	shutdownSignal      chan struct{}
	activeGoroutines    sync.WaitGroup
}

// NewRedisServerInstance crée une nouvelle instance de serveur
func NewRedisServerInstance(serverConfiguration *config.ServerConfiguration) *RedisServerInstance {
	redisStorage := storage.NewRedisInMemoryStorage()
	commandRegistry := commands.NewRedisCommandRegistry()

	redisServerInstance := &RedisServerInstance{
		serverConfiguration: serverConfiguration,
		redisStorage:        redisStorage,
		commandRegistry:     commandRegistry,
		connectedClients:    make(map[net.Conn]bool),
		shutdownSignal:      make(chan struct{}),
	}

	// Initialiser la persistence RDB si activée
	if serverConfiguration.PersistenceConfiguration.RDBEnabled {
		redisServerInstance.rdbPersistence = persistence.NewRDBPersistence(
			serverConfiguration.PersistenceConfiguration.RDBFilePath,
			serverConfiguration.PersistenceConfiguration.RDBSaveInterval,
			redisStorage,
		)

		// Configurer les commandes RDB
		commandRegistry.SetRDBPersistence(redisServerInstance.rdbPersistence)
	}

	// Démarrage du garbage collector pour les clés expirées
	redisServerInstance.startExpirationGarbageCollector()

	return redisServerInstance
}
