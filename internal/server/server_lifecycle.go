package server

import (
	"fmt"
	"log"
	"net"
)

// StartRedisServer démarre le serveur TCP
func (redisServerInstance *RedisServerInstance) StartRedisServer() error {
	// Charger les données depuis RDB si disponible
	if redisServerInstance.rdbPersistence != nil {
		if err := redisServerInstance.rdbPersistence.LoadSnapshot(); err != nil {
			log.Printf("⚠️  Erreur chargement RDB: %v", err)
		}

		// Démarrer la sauvegarde automatique
		redisServerInstance.rdbPersistence.StartAutomaticSave()
	}

	serverAddress := fmt.Sprintf("%s:%d",
		redisServerInstance.serverConfiguration.NetworkConfiguration.HostAddress,
		redisServerInstance.serverConfiguration.NetworkConfiguration.PortNumber)

	networkListener, listenError := net.Listen("tcp", serverAddress)
	if listenError != nil {
		return fmt.Errorf("impossible d'écouter sur %s: %v", serverAddress, listenError)
	}

	redisServerInstance.networkListener = networkListener
	log.Printf("🚀 Serveur Redis-Go en écoute sur %s", serverAddress)

	// Boucle d'acceptation des connexions
	for {
		clientConnection, acceptError := networkListener.Accept()
		if acceptError != nil {
			select {
			case <-redisServerInstance.shutdownSignal:
				// Arrêt normal du serveur
				return nil
			default:
				log.Printf("❌ Erreur lors de l'acceptation de connexion: %v", acceptError)
				continue
			}
		}

		log.Printf("🔗 Nouvelle connexion depuis %s", clientConnection.RemoteAddr())

		// Vérification du nombre maximum de connexions
		redisServerInstance.clientsMutex.Lock()
		if len(redisServerInstance.connectedClients) >= redisServerInstance.serverConfiguration.PerformanceConfiguration.MaximumConnections {
			redisServerInstance.clientsMutex.Unlock()
			clientConnection.Close()
			log.Printf("🚫 Connexion refusée: limite atteinte (%d connexions max)",
				redisServerInstance.serverConfiguration.PerformanceConfiguration.MaximumConnections)
			continue
		}

		redisServerInstance.connectedClients[clientConnection] = true
		redisServerInstance.clientsMutex.Unlock()

		// Gestion du client dans une goroutine séparée
		redisServerInstance.activeGoroutines.Add(1)
		go redisServerInstance.handleClientConnection(clientConnection)
	}
}

// StopRedisServer arrête le serveur proprement
func (redisServerInstance *RedisServerInstance) StopRedisServer() error {
	log.Printf("⏹️  Arrêt du serveur en cours...")
	close(redisServerInstance.shutdownSignal)

	if redisServerInstance.networkListener != nil {
		redisServerInstance.networkListener.Close()
	}

	// Fermeture de toutes les connexions clients
	redisServerInstance.clientsMutex.Lock()
	connectedClientCount := len(redisServerInstance.connectedClients)
	for clientConnection := range redisServerInstance.connectedClients {
		clientConnection.Close()
	}
	redisServerInstance.clientsMutex.Unlock()

	if connectedClientCount > 0 {
		log.Printf("🔌 Fermeture de %d connexions clients...", connectedClientCount)
	}

	// Sauvegarde finale RDB si configurée
	if redisServerInstance.rdbPersistence != nil && redisServerInstance.serverConfiguration.PersistenceConfiguration.RDBSaveOnExit {
		redisServerInstance.rdbPersistence.Stop()
	}

	// Attente de la fin de toutes les goroutines
	redisServerInstance.activeGoroutines.Wait()

	return nil
}
