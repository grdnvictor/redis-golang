package server

import (
	"log"
	"time"
)

// startExpirationGarbageCollector démarre le garbage collector pour les clés expirées
func (redisServerInstance *RedisServerInstance) startExpirationGarbageCollector() {
	redisServerInstance.activeGoroutines.Add(1)
	go func() {
		defer redisServerInstance.activeGoroutines.Done()

		garbageCollectionTicker := time.NewTicker(redisServerInstance.serverConfiguration.MaintenanceConfiguration.ExpirationCheckInterval)
		defer garbageCollectionTicker.Stop()

		log.Printf("🧹 Garbage collector démarré (intervalle: %v)", redisServerInstance.serverConfiguration.MaintenanceConfiguration.ExpirationCheckInterval)

		for {
			select {
			case <-redisServerInstance.shutdownSignal:
				log.Printf("🧹 Arrêt du garbage collector")
				return
			case <-garbageCollectionTicker.C:
				// Nettoyage des clés expirées
				cleanedKeyCount := redisServerInstance.redisStorage.CleanupExpiredKeys()
				if cleanedKeyCount > 0 {
					log.Printf("🧹 Nettoyage: %d clés expirées supprimées", cleanedKeyCount)
				}
			}
		}
	}()
}
