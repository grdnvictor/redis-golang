package server

import (
	"log"
	"net"
	"time"

	"redis-go/internal/protocol"
)

// handleClientConnection gère une connexion client
func (redisServerInstance *RedisServerInstance) handleClientConnection(clientConnection net.Conn) {
	defer redisServerInstance.activeGoroutines.Done()
	defer func() {
		log.Printf("🔌 Connexion fermée depuis %s", clientConnection.RemoteAddr())
		clientConnection.Close()
		redisServerInstance.clientsMutex.Lock()
		delete(redisServerInstance.connectedClients, clientConnection)
		redisServerInstance.clientsMutex.Unlock()
	}()

	protocolParser := protocol.NewRedisSerializationProtocolParser(clientConnection)
	protocolEncoder := protocol.NewRedisSerializationProtocolEncoder(clientConnection)

	// Boucle de traitement des commandes
	for {
		select {
		case <-redisServerInstance.shutdownSignal:
			return
		default:
			// Définir un timeout pour éviter les blocages
			clientConnection.SetReadDeadline(time.Now().Add(30 * time.Second))

			// Parsing de la commande
			parsedCommandArguments, parseError := protocolParser.ParseIncomingCommand()
			if parseError != nil {
				// Log différencié selon le type d'erreur
				if networkError, isNetworkError := parseError.(net.Error); isNetworkError && networkError.Timeout() {
					log.Printf("⏰ Timeout de connexion pour %s", clientConnection.RemoteAddr())
				} else {
					log.Printf("⚠️  Erreur de parsing depuis %s: %v", clientConnection.RemoteAddr(), parseError)
				}
				return
			}

			if len(parsedCommandArguments) == 0 {
				continue
			}

			// Extraction de la commande et des arguments
			receivedCommandName := parsedCommandArguments[0]
			receivedCommandArguments := parsedCommandArguments[1:]

			// Log des commandes (optionnel, peut être verbeux)
			// log.Printf("📝 Commande reçue de %s: %s %v", clientConnection.RemoteAddr(), receivedCommandName, receivedCommandArguments)

			// Exécution de la commande
			if executionError := redisServerInstance.commandRegistry.ExecuteCommand(receivedCommandName, receivedCommandArguments, redisServerInstance.redisStorage, protocolEncoder); executionError != nil {
				log.Printf("❌ Erreur d'exécution de commande pour %s: %v", clientConnection.RemoteAddr(), executionError)
				protocolEncoder.WriteErrorResponse("ERREUR : erreur interne du serveur")
			}
		}
	}
}
