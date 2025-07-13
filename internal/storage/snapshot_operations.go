package storage

import "time"

// StorageSnapshot représente un snapshot complet du stockage
type StorageSnapshot struct {
	Data      map[string]*RedisStorageValue `json:"data"`
	Timestamp time.Time                     `json:"timestamp"`
	Version   string                        `json:"version"`
}

// CreateSnapshot crée un snapshot complet du stockage
func (redisStorage *RedisInMemoryStorage) CreateSnapshot() StorageSnapshot {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	snapshot := StorageSnapshot{
		Data:      make(map[string]*RedisStorageValue),
		Timestamp: time.Now(),
		Version:   "1.0",
	}

	// Copier toutes les données valides (non expirées)
	currentTime := time.Now()
	for key, value := range redisStorage.storageData {
		if value.ExpirationTime == nil || currentTime.Before(*value.ExpirationTime) {
			// Copie profonde de la valeur
			snapshot.Data[key] = &RedisStorageValue{
				StoredData:     copyStoredData(value.StoredData, value.DataType),
				DataType:       value.DataType,
				ExpirationTime: copyTime(value.ExpirationTime),
			}
		}
	}

	// Reset le compteur de changements après snapshot
	redisStorage.changesSinceLastSave = 0

	return snapshot
}

// RestoreFromSnapshot restaure les données depuis un snapshot
func (redisStorage *RedisInMemoryStorage) RestoreFromSnapshot(snapshot StorageSnapshot) {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	// Vider le stockage actuel
	redisStorage.storageData = make(map[string]*RedisStorageValue)

	// Restaurer les données
	currentTime := time.Now()

	for key, value := range snapshot.Data {
		// Vérifier si la clé n'a pas expiré depuis la sauvegarde
		if value.ExpirationTime == nil || currentTime.Before(*value.ExpirationTime) {
			redisStorage.storageData[key] = &RedisStorageValue{
				StoredData:     copyStoredData(value.StoredData, value.DataType),
				DataType:       value.DataType,
				ExpirationTime: copyTime(value.ExpirationTime),
			}
		}
	}

	// Reset le compteur de changements après restauration
	redisStorage.changesSinceLastSave = 0
}

// copyStoredData effectue une copie profonde des données selon leur type
func copyStoredData(data interface{}, dataType RedisDataType) interface{} {
	switch dataType {
	case RedisStringType:
		return data.(string)

	case RedisListType:
		original := data.(*RedisListStructure)
		copy := &RedisListStructure{
			ListElements: make([]string, len(original.ListElements)),
		}
		for i, element := range original.ListElements {
			copy.ListElements[i] = element
		}
		return copy

	case RedisSetType:
		original := data.(*RedisSetStructure)
		copy := &RedisSetStructure{
			SetElements: make(map[string]bool),
		}
		for member := range original.SetElements {
			copy.SetElements[member] = true
		}
		return copy

	case RedisHashType:
		original := data.(*RedisHashStructure)
		copy := &RedisHashStructure{
			HashFields: make(map[string]string),
		}
		for field, value := range original.HashFields {
			copy.HashFields[field] = value
		}
		return copy

	default:
		// Pour les types non supportés, retourner tel quel
		return data
	}
}

// copyTime effectue une copie d'un pointeur time.Time
func copyTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	copied := *t
	return &copied
}

// GetChangesSinceLastSave retourne le nombre de changements depuis la dernière sauvegarde
func (redisStorage *RedisInMemoryStorage) GetChangesSinceLastSave() int64 {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()
	return redisStorage.changesSinceLastSave
}

// incrementChanges incrémente le compteur de changements
func (redisStorage *RedisInMemoryStorage) incrementChanges() {
	redisStorage.changesSinceLastSave++
}
