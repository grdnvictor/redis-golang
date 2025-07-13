package storage

// AddMembersToSet ajoute des membres à un set
func (redisStorage *RedisInMemoryStorage) AddMembersToSet(setKey string, newMembers []string) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[setKey]
	var redisSetStructure *RedisSetStructure

	if !keyExists {
		redisSetStructure = &RedisSetStructure{SetElements: make(map[string]bool)}
		redisStorage.storageData[setKey] = &RedisStorageValue{
			StoredData: redisSetStructure,
			DataType:   RedisSetType,
		}
		redisStorage.incrementChanges()
	} else {
		if storageValue.DataType != RedisSetType {
			return -1
		}
		redisSetStructure = storageValue.StoredData.(*RedisSetStructure)
	}

	addedMemberCount := 0
	for _, newMember := range newMembers {
		if !redisSetStructure.SetElements[newMember] {
			redisSetStructure.SetElements[newMember] = true
			addedMemberCount++
		}
	}

	if addedMemberCount > 0 {
		redisStorage.incrementChanges()
	}

	return addedMemberCount
}

// GetAllSetMembers retourne tous les membres d'un set
func (redisStorage *RedisInMemoryStorage) GetAllSetMembers(setKey string) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[setKey]
	if !keyExists {
		return []string{}
	}

	if storageValue.DataType != RedisSetType {
		return nil
	}

	redisSetStructure := storageValue.StoredData.(*RedisSetStructure)
	setMembers := make([]string, 0, len(redisSetStructure.SetElements))
	for setMember := range redisSetStructure.SetElements {
		setMembers = append(setMembers, setMember)
	}

	return setMembers
}

// CheckSetMemberExists vérifie si un membre est dans un set
func (redisStorage *RedisInMemoryStorage) CheckSetMemberExists(setKey string, memberToCheck string) bool {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[setKey]
	if !keyExists {
		return false
	}

	if storageValue.DataType != RedisSetType {
		return false
	}

	redisSetStructure := storageValue.StoredData.(*RedisSetStructure)
	return redisSetStructure.SetElements[memberToCheck]
}

// RemoveMembersFromSet supprime des membres d'un set
func (redisStorage *RedisInMemoryStorage) RemoveMembersFromSet(setKey string, membersToRemove []string) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[setKey]
	if !keyExists {
		return 0 // Set n'existe pas, 0 membres supprimés
	}

	if storageValue.DataType != RedisSetType {
		return -1 // Erreur de type
	}

	redisSetStructure := storageValue.StoredData.(*RedisSetStructure)
	removedCount := 0

	for _, memberToRemove := range membersToRemove {
		if redisSetStructure.SetElements[memberToRemove] {
			delete(redisSetStructure.SetElements, memberToRemove)
			removedCount++
		}
	}

	// Si le set devient vide, supprimer la clé
	if len(redisSetStructure.SetElements) == 0 {
		delete(redisStorage.storageData, setKey)
	}

	if removedCount > 0 {
		redisStorage.incrementChanges()
	}

	return removedCount
}

// GetSetCardinality retourne le nombre de membres dans un set
func (redisStorage *RedisInMemoryStorage) GetSetCardinality(setKey string) int {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[setKey]
	if !keyExists {
		return 0
	}

	if storageValue.DataType != RedisSetType {
		return -1 // Erreur de type
	}

	redisSetStructure := storageValue.StoredData.(*RedisSetStructure)
	return len(redisSetStructure.SetElements)
}

// ComputeSetDifference calcule la différence entre sets (premier set - autres sets)
func (redisStorage *RedisInMemoryStorage) ComputeSetDifference(setKeys []string) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	if len(setKeys) == 0 {
		return []string{}
	}

	// Récupérer le premier set
	firstSetValue, keyExists := redisStorage.storageData[setKeys[0]]
	if !keyExists {
		return []string{} // Premier set n'existe pas = résultat vide
	}

	if firstSetValue.DataType != RedisSetType {
		return nil // Erreur de type
	}

	firstSet := firstSetValue.StoredData.(*RedisSetStructure)
	resultMembers := make(map[string]bool)

	// Copier tous les membres du premier set
	for member := range firstSet.SetElements {
		resultMembers[member] = true
	}

	// Soustraire les membres des autres sets
	for i := 1; i < len(setKeys); i++ {
		otherSetValue, otherKeyExists := redisStorage.storageData[setKeys[i]]
		if !otherKeyExists {
			continue // Set n'existe pas, ignorer
		}

		if otherSetValue.DataType != RedisSetType {
			return nil // Erreur de type
		}

		otherSet := otherSetValue.StoredData.(*RedisSetStructure)
		for member := range otherSet.SetElements {
			delete(resultMembers, member)
		}
	}

	// Convertir en slice
	result := make([]string, 0, len(resultMembers))
	for member := range resultMembers {
		result = append(result, member)
	}

	return result
}

// ComputeSetIntersection calcule l'intersection de tous les sets
func (redisStorage *RedisInMemoryStorage) ComputeSetIntersection(setKeys []string) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	if len(setKeys) == 0 {
		return []string{}
	}

	// Récupérer le premier set
	firstSetValue, keyExists := redisStorage.storageData[setKeys[0]]
	if !keyExists {
		return []string{} // Premier set n'existe pas = résultat vide
	}

	if firstSetValue.DataType != RedisSetType {
		return nil // Erreur de type
	}

	firstSet := firstSetValue.StoredData.(*RedisSetStructure)
	resultMembers := make(map[string]bool)

	// Commencer avec tous les membres du premier set
	for member := range firstSet.SetElements {
		resultMembers[member] = true
	}

	// Intersection avec chaque autre set
	for i := 1; i < len(setKeys); i++ {
		otherSetValue, otherKeyExists := redisStorage.storageData[setKeys[i]]
		if !otherKeyExists {
			return []string{} // Un set n'existe pas = intersection vide
		}

		if otherSetValue.DataType != RedisSetType {
			return nil // Erreur de type
		}

		otherSet := otherSetValue.StoredData.(*RedisSetStructure)

		// Garder seulement les membres qui sont aussi dans l'autre set
		for member := range resultMembers {
			if !otherSet.SetElements[member] {
				delete(resultMembers, member)
			}
		}
	}

	// Convertir en slice
	result := make([]string, 0, len(resultMembers))
	for member := range resultMembers {
		result = append(result, member)
	}

	return result
}

// ComputeSetUnion calcule l'union de tous les sets
func (redisStorage *RedisInMemoryStorage) ComputeSetUnion(setKeys []string) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	if len(setKeys) == 0 {
		return []string{}
	}

	resultMembers := make(map[string]bool)

	// Ajouter tous les membres de tous les sets
	for _, setKey := range setKeys {
		setValue, keyExists := redisStorage.storageData[setKey]
		if !keyExists {
			continue // Set n'existe pas, ignorer
		}

		if setValue.DataType != RedisSetType {
			return nil // Erreur de type
		}

		setStructure := setValue.StoredData.(*RedisSetStructure)
		for member := range setStructure.SetElements {
			resultMembers[member] = true
		}
	}

	// Convertir en slice
	result := make([]string, 0, len(resultMembers))
	for member := range resultMembers {
		result = append(result, member)
	}

	return result
}
