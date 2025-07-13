package storage

import "time"

// GetKeyTTL retourne le TTL d'une clé en secondes ou millisecondes
// Retourne -2 si la clé n'existe pas, -1 si pas de TTL, sinon le temps restant
func (redisStorage *RedisInMemoryStorage) GetKeyTTL(storageKey string, inMilliseconds bool) int64 {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[storageKey]
	if !keyExists {
		return -2 // Clé n'existe pas
	}

	// Vérifier si la clé a déjà expiré
	currentTime := time.Now()
	if storageValue.ExpirationTime != nil && currentTime.After(*storageValue.ExpirationTime) {
		// Clé expirée - suppression lazy
		delete(redisStorage.storageData, storageKey)
		return -2
	}

	// Pas de TTL défini
	if storageValue.ExpirationTime == nil {
		return -1
	}

	// Calculer le temps restant
	timeRemaining := storageValue.ExpirationTime.Sub(currentTime)
	if inMilliseconds {
		return int64(timeRemaining.Nanoseconds() / 1000000) // Convertir en millisecondes
	}
	return int64(timeRemaining.Seconds()) // Convertir en secondes
}

// SetKeyExpiration définit un TTL sur une clé existante
// Retourne true si la clé existe et que le TTL a été défini, false sinon
func (redisStorage *RedisInMemoryStorage) SetKeyExpiration(storageKey string, timeToLive time.Duration) bool {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[storageKey]
	if !keyExists {
		return false
	}

	// Vérifier si la clé a déjà expiré
	currentTime := time.Now()
	if storageValue.ExpirationTime != nil && currentTime.After(*storageValue.ExpirationTime) {
		// Clé expirée - suppression
		delete(redisStorage.storageData, storageKey)
		return false
	}

	// Définir la nouvelle expiration
	newExpirationTime := currentTime.Add(timeToLive)
	storageValue.ExpirationTime = &newExpirationTime

	return true
}

// RemoveKeyExpiration supprime le TTL d'une clé (la rend persistante)
// Retourne true si la clé existe et avait un TTL, false sinon
func (redisStorage *RedisInMemoryStorage) RemoveKeyExpiration(storageKey string) bool {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[storageKey]
	if !keyExists {
		return false
	}

	// Vérifier si la clé a déjà expiré
	currentTime := time.Now()
	if storageValue.ExpirationTime != nil && currentTime.After(*storageValue.ExpirationTime) {
		// Clé expirée - suppression
		delete(redisStorage.storageData, storageKey)
		return false
	}

	// Vérifier si la clé avait un TTL
	hadTTL := storageValue.ExpirationTime != nil

	// Supprimer le TTL
	storageValue.ExpirationTime = nil

	return hadTTL
}
