package storage

// PushElementsToList ajoute des éléments à une liste (gauche ou droite)
func (redisStorage *RedisInMemoryStorage) PushElementsToList(listKey string, newElements []string, pushToLeft bool) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	var redisListStructure *RedisListStructure

	if !keyExists {
		// Créer une nouvelle liste
		redisListStructure = &RedisListStructure{ListElements: make([]string, 0)}
		redisStorage.storageData[listKey] = &RedisStorageValue{
			StoredData: redisListStructure,
			DataType:   RedisListType,
		}
		redisStorage.incrementChanges()
	} else {
		// Vérifier que c'est bien une liste
		if storageValue.DataType != RedisListType {
			return -1 // Erreur de type
		}
		redisListStructure = storageValue.StoredData.(*RedisListStructure)
	}

	// Ajouter les éléments
	if pushToLeft {
		// LPUSH - ajouter à gauche (début)
		updatedElements := make([]string, len(newElements)+len(redisListStructure.ListElements))
		copy(updatedElements, newElements)
		copy(updatedElements[len(newElements):], redisListStructure.ListElements)
		redisListStructure.ListElements = updatedElements
	} else {
		// RPUSH - ajouter à droite (fin)
		redisListStructure.ListElements = append(redisListStructure.ListElements, newElements...)
	}

	redisStorage.incrementChanges()
	return len(redisListStructure.ListElements)
}

// PopElementFromList supprime et retourne un élément de la liste
func (redisStorage *RedisInMemoryStorage) PopElementFromList(listKey string, popFromLeft bool) (string, bool) {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return "", false
	}

	if storageValue.DataType != RedisListType {
		return "", false
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	if len(redisListStructure.ListElements) == 0 {
		return "", false
	}

	var poppedElement string
	if popFromLeft {
		// LPOP - supprimer à gauche
		poppedElement = redisListStructure.ListElements[0]
		redisListStructure.ListElements = redisListStructure.ListElements[1:]
	} else {
		// RPOP - supprimer à droite
		poppedElement = redisListStructure.ListElements[len(redisListStructure.ListElements)-1]
		redisListStructure.ListElements = redisListStructure.ListElements[:len(redisListStructure.ListElements)-1]
	}

	// Supprimer la clé si la liste est vide
	if len(redisListStructure.ListElements) == 0 {
		delete(redisStorage.storageData, listKey)
	}

	redisStorage.incrementChanges()
	return poppedElement, true
}

// GetListLength retourne la longueur d'une liste
func (redisStorage *RedisInMemoryStorage) GetListLength(listKey string) int {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return 0
	}

	if storageValue.DataType != RedisListType {
		return -1 // Erreur de type
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	return len(redisListStructure.ListElements)
}

// GetListElementsInRange retourne une partie de la liste
func (redisStorage *RedisInMemoryStorage) GetListElementsInRange(listKey string, startIndex, stopIndex int) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return []string{}
	}

	if storageValue.DataType != RedisListType {
		return nil // Erreur de type
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	listLength := len(redisListStructure.ListElements)

	if listLength == 0 {
		return []string{}
	}

	// Gérer les indices négatifs (comme Redis)
	if startIndex < 0 {
		startIndex = listLength + startIndex
	}
	if stopIndex < 0 {
		stopIndex = listLength + stopIndex
	}

	// Limiter aux bornes
	if startIndex < 0 {
		startIndex = 0
	}
	if stopIndex >= listLength {
		stopIndex = listLength - 1
	}
	if startIndex > stopIndex {
		return []string{}
	}

	return redisListStructure.ListElements[startIndex : stopIndex+1]
}

// SetListElement définit un élément à un index spécifique (LSET)
func (redisStorage *RedisInMemoryStorage) SetListElement(listKey string, index int, newElement string) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return -1 // Liste n'existe pas
	}

	if storageValue.DataType != RedisListType {
		return -1 // Erreur de type
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	listLength := len(redisListStructure.ListElements)

	// Gérer les indices négatifs
	if index < 0 {
		index = listLength + index
	}

	// Vérifier les limites
	if index < 0 || index >= listLength {
		return 0 // Index hors limites
	}

	redisListStructure.ListElements[index] = newElement
	redisStorage.incrementChanges()
	return 1 // Succès
}

// RemoveListElements supprime des éléments selon un critère (LREM)
func (redisStorage *RedisInMemoryStorage) RemoveListElements(listKey string, count int, elementToRemove string) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return 0 // Liste n'existe pas
	}

	if storageValue.DataType != RedisListType {
		return -1 // Erreur de type
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	originalElements := redisListStructure.ListElements
	newElements := make([]string, 0, len(originalElements))
	removedCount := 0

	if count == 0 {
		// Supprimer toutes les occurrences
		for _, element := range originalElements {
			if element != elementToRemove {
				newElements = append(newElements, element)
			} else {
				removedCount++
			}
		}
	} else if count > 0 {
		// Supprimer les premières occurrences
		for _, element := range originalElements {
			if element == elementToRemove && removedCount < count {
				removedCount++
			} else {
				newElements = append(newElements, element)
			}
		}
	} else {
		// count < 0 : supprimer les dernières occurrences
		absCount := -count
		// Parcourir de la fin vers le début
		for i := len(originalElements) - 1; i >= 0; i-- {
			element := originalElements[i]
			if element == elementToRemove && removedCount < absCount {
				removedCount++
			} else {
				newElements = append([]string{element}, newElements...)
			}
		}
	}

	redisListStructure.ListElements = newElements

	// Supprimer la clé si la liste devient vide
	if len(newElements) == 0 {
		delete(redisStorage.storageData, listKey)
	}

	if removedCount > 0 {
		redisStorage.incrementChanges()
	}

	return removedCount
}

// InsertIntoList insère un élément avant ou après un pivot (LINSERT)
func (redisStorage *RedisInMemoryStorage) InsertIntoList(listKey string, insertBefore bool, pivotElement string, newElement string) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return -1 // Liste n'existe pas
	}

	if storageValue.DataType != RedisListType {
		return -1 // Erreur de type
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	originalElements := redisListStructure.ListElements

	// Chercher le pivot
	pivotIndex := -1
	for i, element := range originalElements {
		if element == pivotElement {
			pivotIndex = i
			break
		}
	}

	if pivotIndex == -1 {
		return -2 // Pivot non trouvé
	}

	// Insérer l'élément
	newElements := make([]string, 0, len(originalElements)+1)

	if insertBefore {
		// Insérer avant le pivot
		newElements = append(newElements, originalElements[:pivotIndex]...)
		newElements = append(newElements, newElement)
		newElements = append(newElements, originalElements[pivotIndex:]...)
	} else {
		// Insérer après le pivot
		newElements = append(newElements, originalElements[:pivotIndex+1]...)
		newElements = append(newElements, newElement)
		newElements = append(newElements, originalElements[pivotIndex+1:]...)
	}

	redisListStructure.ListElements = newElements
	redisStorage.incrementChanges()
	return len(newElements)
}

// TrimList garde seulement les éléments dans la plage spécifiée (LTRIM)
func (redisStorage *RedisInMemoryStorage) TrimList(listKey string, startIndex, stopIndex int) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return 0 // Liste n'existe pas, rien à faire
	}

	if storageValue.DataType != RedisListType {
		return -1 // Erreur de type
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	listLength := len(redisListStructure.ListElements)

	if listLength == 0 {
		return 1 // Liste vide, rien à faire
	}

	// Gérer les indices négatifs
	if startIndex < 0 {
		startIndex = listLength + startIndex
	}
	if stopIndex < 0 {
		stopIndex = listLength + stopIndex
	}

	// Limiter aux bornes
	if startIndex < 0 {
		startIndex = 0
	}
	if stopIndex >= listLength {
		stopIndex = listLength - 1
	}

	if startIndex > stopIndex {
		// Plage invalide, vider la liste
		delete(redisStorage.storageData, listKey)
	} else {
		// Garder seulement la plage spécifiée
		redisListStructure.ListElements = redisListStructure.ListElements[startIndex : stopIndex+1]

		// Si la liste devient vide après trim, supprimer la clé
		if len(redisListStructure.ListElements) == 0 {
			delete(redisStorage.storageData, listKey)
		}
	}

	redisStorage.incrementChanges()
	return 1 // Succès
}
