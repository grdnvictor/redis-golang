package persistence

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"redis-go/internal/storage"
)

// RDBPersistence gère la sauvegarde/restauration RDB
type RDBPersistence struct {
	filePath       string
	saveInterval   time.Duration
	storage        *storage.RedisInMemoryStorage
	stopChannel    chan struct{}
	saveInProgress bool
	saveMutex      sync.Mutex
	lastSaveTime   time.Time
	totalSaves     int64
	lastSaveStatus string
	isShuttingDown bool
}

// NewRDBPersistence crée une nouvelle instance de persistence RDB
func NewRDBPersistence(filePath string, saveInterval time.Duration, storage *storage.RedisInMemoryStorage) *RDBPersistence {
	return &RDBPersistence{
		filePath:       filePath,
		saveInterval:   saveInterval,
		storage:        storage,
		stopChannel:    make(chan struct{}),
		lastSaveStatus: "ok",
	}
}

// StartAutomaticSave démarre la sauvegarde automatique en arrière-plan
func (rdb *RDBPersistence) StartAutomaticSave() {
	go func() {
		ticker := time.NewTicker(rdb.saveInterval)
		defer ticker.Stop()

		log.Printf("💾 RDB: Sauvegarde automatique démarrée (intervalle: %v)", rdb.saveInterval)

		for {
			select {
			case <-rdb.stopChannel:
				log.Printf("💾 RDB: Arrêt de la sauvegarde automatique")
				return
			case <-ticker.C:
				if !rdb.isShuttingDown {
					if err := rdb.BackgroundSave(); err != nil {
						log.Printf("❌ RDB: Erreur sauvegarde automatique: %v", err)
					}
				}
			}
		}
	}()
}

// BackgroundSave effectue une sauvegarde non-bloquante (BGSAVE)
func (rdb *RDBPersistence) BackgroundSave() error {
	rdb.saveMutex.Lock()
	if rdb.saveInProgress {
		rdb.saveMutex.Unlock()
		return fmt.Errorf("sauvegarde déjà en cours")
	}
	rdb.saveInProgress = true
	rdb.saveMutex.Unlock()

	// Lancer la sauvegarde dans une goroutine séparée
	go func() {
		defer func() {
			rdb.saveMutex.Lock()
			rdb.saveInProgress = false
			rdb.saveMutex.Unlock()
		}()

		if err := rdb.performSave(); err != nil {
			log.Printf("❌ RDB: Erreur BGSAVE: %v", err)
			rdb.lastSaveStatus = "error"
		} else {
			rdb.lastSaveStatus = "ok"
		}
	}()

	return nil
}

// Save effectue une sauvegarde synchrone bloquante (SAVE)
func (rdb *RDBPersistence) Save() error {
	rdb.saveMutex.Lock()
	defer rdb.saveMutex.Unlock()

	if rdb.saveInProgress {
		return fmt.Errorf("sauvegarde en cours, utilisez BGSAVE")
	}

	rdb.saveInProgress = true
	defer func() {
		rdb.saveInProgress = false
	}()

	return rdb.performSave()
}

// performSave effectue la sauvegarde réelle
func (rdb *RDBPersistence) performSave() error {
	startTime := time.Now()
	log.Printf("💾 RDB: Début sauvegarde...")

	// Créer le dossier si nécessaire
	if err := os.MkdirAll(filepath.Dir(rdb.filePath), 0755); err != nil {
		return fmt.Errorf("création dossier: %v", err)
	}

	// Fichier temporaire pour sauvegarde atomique
	tempFile := rdb.filePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("création fichier temp: %v", err)
	}
	defer file.Close()

	// Créer le snapshot des données
	snapshot := rdb.storage.CreateSnapshot()

	// Encoder les données
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(snapshot); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("encodage données: %v", err)
	}

	// Forcer l'écriture sur disque
	if err := file.Sync(); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("sync fichier: %v", err)
	}

	file.Close()

	// Remplacer le fichier principal atomiquement
	if err := os.Rename(tempFile, rdb.filePath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("remplacement fichier: %v", err)
	}

	// Mettre à jour les statistiques
	rdb.lastSaveTime = time.Now()
	rdb.totalSaves++

	duration := time.Since(startTime)
	log.Printf("✅ RDB: Sauvegarde terminée (%s, %d clés, %v)",
		rdb.filePath, len(snapshot.Data), duration)

	return nil
}

// LoadSnapshot restaure les données depuis le fichier RDB
func (rdb *RDBPersistence) LoadSnapshot() error {
	if _, err := os.Stat(rdb.filePath); os.IsNotExist(err) {
		log.Printf("📂 RDB: Aucun fichier trouvé (%s), démarrage à vide", rdb.filePath)
		return nil
	}

	log.Printf("📥 RDB: Chargement depuis %s...", rdb.filePath)

	file, err := os.Open(rdb.filePath)
	if err != nil {
		return fmt.Errorf("ouverture fichier RDB: %v", err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var snapshot storage.StorageSnapshot

	if err := decoder.Decode(&snapshot); err != nil {
		return fmt.Errorf("décodage RDB: %v", err)
	}

	// Restaurer les données dans le storage
	rdb.storage.RestoreFromSnapshot(snapshot)

	log.Printf("✅ RDB: Données restaurées (%d clés, snapshot du %v)",
		len(snapshot.Data), snapshot.Timestamp.Format("2006-01-02 15:04:05"))

	return nil
}

// Stop arrête la persistence et effectue une sauvegarde finale
func (rdb *RDBPersistence) Stop() {
	rdb.isShuttingDown = true
	close(rdb.stopChannel)

	log.Printf("💾 RDB: Sauvegarde finale avant arrêt...")
	if err := rdb.Save(); err != nil {
		log.Printf("❌ RDB: Erreur sauvegarde finale: %v", err)
	} else {
		log.Printf("✅ RDB: Sauvegarde finale terminée")
	}
}

// GetStats retourne les statistiques RDB
func (rdb *RDBPersistence) GetStats() map[string]interface{} {
	rdb.saveMutex.Lock()
	defer rdb.saveMutex.Unlock()

	return map[string]interface{}{
		"rdb_changes_since_last_save": rdb.storage.GetChangesSinceLastSave(),
		"rdb_bgsave_in_progress":      rdb.saveInProgress,
		"rdb_last_save_time":          rdb.lastSaveTime.Unix(),
		"rdb_last_bgsave_status":      rdb.lastSaveStatus,
		"rdb_total_saves":             rdb.totalSaves,
		"rdb_file_path":               rdb.filePath,
	}
}

// IsSaveInProgress retourne true si une sauvegarde est en cours
func (rdb *RDBPersistence) IsSaveInProgress() bool {
	rdb.saveMutex.Lock()
	defer rdb.saveMutex.Unlock()
	return rdb.saveInProgress
}

// GetLastSaveTime retourne le timestamp de la dernière sauvegarde
func (rdb *RDBPersistence) GetLastSaveTime() int64 {
	return rdb.lastSaveTime.Unix()
}
