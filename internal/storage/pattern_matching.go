package storage

import "time"

// FindKeysByPattern retourne toutes les clés correspondant au pattern (style Redis glob)
func (redisStorage *RedisInMemoryStorage) FindKeysByPattern(searchPattern string) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	var matchingKeys []string
	currentTime := time.Now()

	for storageKey, storageValue := range redisStorage.storageData {
		// Ignorer les clés expirées
		if storageValue.ExpirationTime != nil && currentTime.After(*storageValue.ExpirationTime) {
			continue
		}

		// Vérifier si la clé correspond au pattern
		if matchesGlobPattern(searchPattern, storageKey) {
			matchingKeys = append(matchingKeys, storageKey)
		}
	}

	return matchingKeys
}

// matchesGlobPattern implémente le pattern matching style Redis avec *, ?, et [...]
func matchesGlobPattern(searchPattern, targetString string) bool {
	return matchGlobRecursive(searchPattern, targetString, 0, 0)
}

// matchGlobRecursive fonction récursive pour le pattern matching
func matchGlobRecursive(searchPattern, targetString string, patternIndex, stringIndex int) bool {
	// Fin du pattern et de la string = match
	if patternIndex == len(searchPattern) && stringIndex == len(targetString) {
		return true
	}

	// Fin du pattern mais pas de la string = pas de match
	if patternIndex == len(searchPattern) {
		return false
	}

	// Fin de la string mais pas du pattern = seulement OK si le reste du pattern est que des *
	if stringIndex == len(targetString) {
		for remainingIndex := patternIndex; remainingIndex < len(searchPattern); remainingIndex++ {
			if searchPattern[remainingIndex] != '*' {
				return false
			}
		}
		return true
	}

	// Caractère actuel du pattern
	currentPatternChar := searchPattern[patternIndex]

	switch currentPatternChar {
	case '*':
		// * peut matcher 0 ou plusieurs caractères
		// Essayer de matcher 0 caractère (avancer dans le pattern seulement)
		if matchGlobRecursive(searchPattern, targetString, patternIndex+1, stringIndex) {
			return true
		}
		// Essayer de matcher 1 caractère à la fois
		if stringIndex < len(targetString) {
			return matchGlobRecursive(searchPattern, targetString, patternIndex, stringIndex+1)
		}
		return false

	case '?':
		// ? doit matcher exactement 1 caractère
		if stringIndex < len(targetString) {
			return matchGlobRecursive(searchPattern, targetString, patternIndex+1, stringIndex+1)
		}
		return false

	case '[':
		// Classe de caractères [abc] ou [a-z] ou [^abc]
		if stringIndex >= len(targetString) {
			return false
		}

		// Trouver la fin de la classe
		characterClassEnd := patternIndex + 1
		for characterClassEnd < len(searchPattern) && searchPattern[characterClassEnd] != ']' {
			characterClassEnd++
		}

		if characterClassEnd >= len(searchPattern) {
			// Pas de ] fermant, traiter [ comme caractère littéral
			if searchPattern[patternIndex] == targetString[stringIndex] {
				return matchGlobRecursive(searchPattern, targetString, patternIndex+1, stringIndex+1)
			}
			return false
		}

		// Extraire la classe sans [ et ]
		characterClass := searchPattern[patternIndex+1 : characterClassEnd]
		characterMatched := matchesCharacterClass(characterClass, targetString[stringIndex])

		if characterMatched {
			return matchGlobRecursive(searchPattern, targetString, characterClassEnd+1, stringIndex+1)
		}
		return false

	case '\\':
		// Caractère d'échappement
		if patternIndex+1 < len(searchPattern) && stringIndex < len(targetString) {
			if searchPattern[patternIndex+1] == targetString[stringIndex] {
				return matchGlobRecursive(searchPattern, targetString, patternIndex+2, stringIndex+1)
			}
		}
		return false

	default:
		// Caractère littéral
		if stringIndex < len(targetString) && currentPatternChar == targetString[stringIndex] {
			return matchGlobRecursive(searchPattern, targetString, patternIndex+1, stringIndex+1)
		}
		return false
	}
}

// matchesCharacterClass vérifie si un caractère correspond à une classe [abc], [a-z], [^abc]
func matchesCharacterClass(characterClass string, targetCharacter byte) bool {
	if len(characterClass) == 0 {
		return false
	}

	isNegatedClass := false
	classIndex := 0

	// Vérifier si c'est une classe négative [^...]
	if characterClass[0] == '^' {
		isNegatedClass = true
		classIndex = 1
	}

	characterMatched := false

	for classIndex < len(characterClass) {
		if classIndex+2 < len(characterClass) && characterClass[classIndex+1] == '-' {
			// Intervalle de caractères [a-z]
			if targetCharacter >= characterClass[classIndex] && targetCharacter <= characterClass[classIndex+2] {
				characterMatched = true
				break
			}
			classIndex += 3
		} else {
			// Caractère simple
			if characterClass[classIndex] == targetCharacter {
				characterMatched = true
				break
			}
			classIndex++
		}
	}

	// Appliquer la négation si nécessaire
	if isNegatedClass {
		return !characterMatched
	}
	return characterMatched
}
