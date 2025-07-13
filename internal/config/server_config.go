package config

import (
	"os"
	"strconv"
	"time"
)

// ServerConfiguration contient toute la configuration du serveur Redis
type ServerConfiguration struct {
	NetworkConfiguration     NetworkConfiguration
	PerformanceConfiguration PerformanceConfiguration
	MaintenanceConfiguration MaintenanceConfiguration
	PersistenceConfiguration PersistenceConfiguration // Nouveau
}

// NetworkConfiguration gère les paramètres réseau
type NetworkConfiguration struct {
	HostAddress string
	PortNumber  int
}

// PerformanceConfiguration gère les paramètres de performance
type PerformanceConfiguration struct {
	MaximumConnections int
}

// MaintenanceConfiguration gère les paramètres de maintenance
type MaintenanceConfiguration struct {
	ExpirationCheckInterval time.Duration
}

// PersistenceConfiguration gère les paramètres de persistence RDB
type PersistenceConfiguration struct {
	RDBEnabled      bool          // Activer/désactiver RDB
	RDBFilePath     string        // Chemin du fichier RDB
	RDBSaveInterval time.Duration // Intervalle de sauvegarde auto
	RDBSaveOnExit   bool          // Sauvegarder à l'arrêt
}

// LoadServerConfiguration charge la configuration depuis les variables d'environnement
// avec des valeurs par défaut raisonnables
func LoadServerConfiguration() *ServerConfiguration {
	configuration := &ServerConfiguration{
		NetworkConfiguration: NetworkConfiguration{
			HostAddress: getEnvironmentString("REDIS_HOST", "localhost"),
			PortNumber:  getEnvironmentInteger("REDIS_PORT", 6379),
		},
		PerformanceConfiguration: PerformanceConfiguration{
			MaximumConnections: getEnvironmentInteger("REDIS_MAX_CONNECTIONS", 1000),
		},
		MaintenanceConfiguration: MaintenanceConfiguration{
			ExpirationCheckInterval: time.Duration(getEnvironmentInteger("REDIS_EXPIRATION_CHECK_INTERVAL", 1)) * time.Second,
		},
		PersistenceConfiguration: PersistenceConfiguration{
			RDBEnabled:      getEnvironmentBool("REDIS_RDB_ENABLED", true),
			RDBFilePath:     getEnvironmentString("REDIS_RDB_FILE", "./data/dump.rdb"),
			RDBSaveInterval: time.Duration(getEnvironmentInteger("REDIS_RDB_SAVE_INTERVAL", 300)) * time.Second, // 5 minutes par défaut
			RDBSaveOnExit:   getEnvironmentBool("REDIS_RDB_SAVE_ON_EXIT", true),
		},
	}

	return configuration
}

// getEnvironmentString récupère une variable d'environnement string avec valeur par défaut
func getEnvironmentString(environmentKey, defaultValue string) string {
	if environmentValue := os.Getenv(environmentKey); environmentValue != "" {
		return environmentValue
	}
	return defaultValue
}

// getEnvironmentInteger récupère une variable d'environnement int avec valeur par défaut
func getEnvironmentInteger(environmentKey string, defaultValue int) int {
	if environmentValue := os.Getenv(environmentKey); environmentValue != "" {
		if integerValue, parseError := strconv.Atoi(environmentValue); parseError == nil {
			return integerValue
		}
	}
	return defaultValue
}

// getEnvironmentBool récupère une variable d'environnement bool avec valeur par défaut
func getEnvironmentBool(environmentKey string, defaultValue bool) bool {
	if environmentValue := os.Getenv(environmentKey); environmentValue != "" {
		if boolValue, parseError := strconv.ParseBool(environmentValue); parseError == nil {
			return boolValue
		}
	}
	return defaultValue
}
