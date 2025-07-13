package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"redis-go/internal/config"
	"redis-go/internal/server"
)

func main() {
	// Chargement de la configuration depuis les variables d'environnement ou valeurs par défaut
	serverConfiguration := config.LoadServerConfiguration()

	// Création du serveur Redis
	redisServerInstance := server.NewRedisServerInstance(serverConfiguration)

	// Canal pour intercepter les signaux système (CTRL+C, SIGTERM)
	systemInterruptSignal := make(chan os.Signal, 1)
	signal.Notify(systemInterruptSignal, os.Interrupt, syscall.SIGTERM)

	// Démarrage du serveur dans une goroutine séparée
	go func() {
		log.Printf("🎯 Démarrage du serveur Redis-Go sur %s:%d",
			serverConfiguration.NetworkConfiguration.HostAddress,
			serverConfiguration.NetworkConfiguration.PortNumber)
		if startupError := redisServerInstance.StartRedisServer(); startupError != nil {
			log.Fatalf("❌ Impossible de démarrer le serveur: %v", startupError)
		}
	}()

	// Attente du signal d'arrêt
	<-systemInterruptSignal
	fmt.Println("\n🛑 Arrêt du serveur en cours...")

	// Arrêt propre du serveur
	if shutdownError := redisServerInstance.StopRedisServer(); shutdownError != nil {
		log.Printf("⚠️  Erreur lors de l'arrêt: %v", shutdownError)
	}

	log.Println("✅ Serveur arrêté proprement")
}
