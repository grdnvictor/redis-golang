package storage

import "strconv"

// SetHashField définit un field dans un hash
func (redisStorage *RedisInMemoryStorage) SetHashField(hashKey string, fieldName string, fieldValue string) bool {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	var redisHashStructure *RedisHashStructure

	if !keyExists {
		redisHashStructure = &RedisHashStructure{HashFields: make(map[string]string)}
		redisStorage.storageData[hashKey] = &RedisStorageValue{
			StoredData: redisHashStructure,
			DataType:   RedisHashType,
		}
		redisStorage.incrementChanges()
	} else {
		if storageValue.DataType != RedisHashType {
			return false
		}
		redisHashStructure = storageValue.StoredData.(*RedisHashStructure)
	}

	_, fieldAlreadyExists := redisHashStructure.HashFields[fieldName]
	redisHashStructure.HashFields[fieldName] = fieldValue
	if !fieldAlreadyExists {
		redisStorage.incrementChanges()
	}
	return !fieldAlreadyExists // true si nouveau field
}

// GetHashField récupère un field d'un hash
func (redisStorage *RedisInMemoryStorage) GetHashField(hashKey string, fieldName string) (string, bool) {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	if !keyExists {
		return "", false
	}

	if storageValue.DataType != RedisHashType {
		return "", false
	}

	redisHashStructure := storageValue.StoredData.(*RedisHashStructure)
	fieldValue, fieldExists := redisHashStructure.HashFields[fieldName]
	return fieldValue, fieldExists
}

// GetAllHashFields retourne tous les fields et valeurs d'un hash
func (redisStorage *RedisInMemoryStorage) GetAllHashFields(hashKey string) map[string]string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	if !keyExists {
		return map[string]string{}
	}

	if storageValue.DataType != RedisHashType {
		return nil
	}

	redisHashStructure := storageValue.StoredData.(*RedisHashStructure)
	hashFieldsCopy := make(map[string]string)
	for fieldName, fieldValue := range redisHashStructure.HashFields {
		hashFieldsCopy[fieldName] = fieldValue
	}
	return hashFieldsCopy
}

// DeleteHashFields supprime des fields d'un hash
func (redisStorage *RedisInMemoryStorage) DeleteHashFields(hashKey string, fieldsToDelete []string) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	if !keyExists {
		return 0 // Hash n'existe pas, 0 fields supprimés
	}

	if storageValue.DataType != RedisHashType {
		return -1 // Erreur de type
	}

	redisHashStructure := storageValue.StoredData.(*RedisHashStructure)
	deletedCount := 0

	for _, fieldName := range fieldsToDelete {
		if _, fieldExists := redisHashStructure.HashFields[fieldName]; fieldExists {
			delete(redisHashStructure.HashFields, fieldName)
			deletedCount++
		}
	}

	// Si le hash devient vide, supprimer la clé
	if len(redisHashStructure.HashFields) == 0 {
		delete(redisStorage.storageData, hashKey)
	}

	if deletedCount > 0 {
		redisStorage.incrementChanges()
	}

	return deletedCount
}

// GetHashLength retourne le nombre de fields dans un hash
func (redisStorage *RedisInMemoryStorage) GetHashLength(hashKey string) int {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	if !keyExists {
		return 0
	}

	if storageValue.DataType != RedisHashType {
		return -1 // Erreur de type
	}

	redisHashStructure := storageValue.StoredData.(*RedisHashStructure)
	return len(redisHashStructure.HashFields)
}

// GetHashKeys retourne tous les field names d'un hash
func (redisStorage *RedisInMemoryStorage) GetHashKeys(hashKey string) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	if !keyExists {
		return []string{}
	}

	if storageValue.DataType != RedisHashType {
		return nil // Erreur de type
	}

	redisHashStructure := storageValue.StoredData.(*RedisHashStructure)
	hashKeys := make([]string, 0, len(redisHashStructure.HashFields))
	for fieldName := range redisHashStructure.HashFields {
		hashKeys = append(hashKeys, fieldName)
	}

	return hashKeys
}

// GetHashValues retourne toutes les valeurs d'un hash
func (redisStorage *RedisInMemoryStorage) GetHashValues(hashKey string) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	if !keyExists {
		return []string{}
	}

	if storageValue.DataType != RedisHashType {
		return nil // Erreur de type
	}

	redisHashStructure := storageValue.StoredData.(*RedisHashStructure)
	hashValues := make([]string, 0, len(redisHashStructure.HashFields))
	for _, fieldValue := range redisHashStructure.HashFields {
		hashValues = append(hashValues, fieldValue)
	}

	return hashValues
}

// IncrementHashField incrémente un field entier dans un hash
func (redisStorage *RedisInMemoryStorage) IncrementHashField(hashKey string, fieldName string, increment int64) *int64 {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	var redisHashStructure *RedisHashStructure

	if !keyExists {
		redisHashStructure = &RedisHashStructure{HashFields: make(map[string]string)}
		redisStorage.storageData[hashKey] = &RedisStorageValue{
			StoredData: redisHashStructure,
			DataType:   RedisHashType,
		}
		redisStorage.incrementChanges()
	} else {
		if storageValue.DataType != RedisHashType {
			return nil // Erreur de type
		}
		redisHashStructure = storageValue.StoredData.(*RedisHashStructure)
	}

	// Récupérer la valeur actuelle
	currentValue := int64(0)
	if existingValue, fieldExists := redisHashStructure.HashFields[fieldName]; fieldExists {
		if parsedValue, parseError := strconv.ParseInt(existingValue, 10, 64); parseError == nil {
			currentValue = parsedValue
		} else {
			return nil // Field existe mais n'est pas un nombre
		}
	}

	// Incrémenter et stocker
	newValue := currentValue + increment
	redisHashStructure.HashFields[fieldName] = strconv.FormatInt(newValue, 10)
	redisStorage.incrementChanges()

	return &newValue
}

// IncrementHashFieldFloat incrémente un field float dans un hash
func (redisStorage *RedisInMemoryStorage) IncrementHashFieldFloat(hashKey string, fieldName string, increment float64) *float64 {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	var redisHashStructure *RedisHashStructure

	if !keyExists {
		redisHashStructure = &RedisHashStructure{HashFields: make(map[string]string)}
		redisStorage.storageData[hashKey] = &RedisStorageValue{
			StoredData: redisHashStructure,
			DataType:   RedisHashType,
		}
		redisStorage.incrementChanges()
	} else {
		if storageValue.DataType != RedisHashType {
			return nil // Erreur de type
		}
		redisHashStructure = storageValue.StoredData.(*RedisHashStructure)
	}

	// Récupérer la valeur actuelle
	currentValue := float64(0)
	if existingValue, fieldExists := redisHashStructure.HashFields[fieldName]; fieldExists {
		if parsedValue, parseError := strconv.ParseFloat(existingValue, 64); parseError == nil {
			currentValue = parsedValue
		} else {
			return nil // Field existe mais n'est pas un nombre
		}
	}

	// Incrémenter et stocker
	newValue := currentValue + increment
	redisHashStructure.HashFields[fieldName] = strconv.FormatFloat(newValue, 'f', -1, 64)
	redisStorage.incrementChanges()

	return &newValue
}
