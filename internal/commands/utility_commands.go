package commands

import (
	"strings"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handlePingCommand implémente PING [message]
func (commandRegistry *RedisCommandRegistry) handlePingCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		return protocolEncoder.WriteSimpleStringResponse("PONG")
	}

	return protocolEncoder.WriteBulkStringResponse(commandArguments[0])
}

// handleEchoCommand implémente ECHO message
func (commandRegistry *RedisCommandRegistry) handleEchoCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'ECHO' (attendu: ECHO message)")
	}

	return protocolEncoder.WriteBulkStringResponse(commandArguments[0])
}

// handleDatabaseSizeCommand implémente DBSIZE
func (commandRegistry *RedisCommandRegistry) handleDatabaseSizeCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : DBSIZE ne prend aucun argument")
	}

	return protocolEncoder.WriteIntegerResponse(int64(redisStorage.GetStorageSize()))
}

// handleFlushAllCommand implémente FLUSHALL
func (commandRegistry *RedisCommandRegistry) handleFlushAllCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : FLUSHALL ne prend aucun argument")
	}

	redisStorage.FlushAllKeys()
	return protocolEncoder.WriteSimpleStringResponse("OK")
}

// handleHelpCommand implémente ALAIDE [commande] - Version mise à jour avec toutes les nouvelles commandes
func (commandRegistry *RedisCommandRegistry) handleHelpCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		// Liste toutes les commandes séparées par des virgules
		return protocolEncoder.WriteSimpleStringResponse("ALAIDE Redis-Go: SET, GET, DEL, EXISTS, TYPE, INCR, DECR, INCRBY, DECRBY, APPEND, STRLEN, GETRANGE, SETRANGE, MSET, MGET, GETSET, MSETNX, GETDEL, TTL, PTTL, EXPIRE, PEXPIRE, PERSIST, LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE, LSET, LREM, LINSERT, LTRIM, SADD, SMEMBERS, SISMEMBER, SREM, SCARD, SDIFF, SINTER, SUNION, HSET, HGET, HGETALL, HEXISTS, HDEL, HLEN, HKEYS, HVALS, HINCRBY, HINCRBYFLOAT, SAVE, BGSAVE, LASTSAVE, INFO, PING, ECHO, KEYS, DBSIZE, FLUSHALL - Tapez ALAIDE <commande> pour details")
	}

	// Aide détaillée pour une commande spécifique
	requestedCommand := strings.ToUpper(commandArguments[0])

	switch requestedCommand {
	case "SET":
		return protocolEncoder.WriteSimpleStringResponse("SET key value [EX seconds] - Stocke une valeur avec TTL optionnel en secondes")
	case "GET":
		return protocolEncoder.WriteSimpleStringResponse("GET key - Recupere une valeur. Retourne (nil) si la cle n'existe pas")
	case "SETNX":
		return protocolEncoder.WriteSimpleStringResponse("SETNX key value - SET si clé n'existe pas. Retourne 1 si créée, 0 si existe déjà")
	case "SETEX":
		return protocolEncoder.WriteSimpleStringResponse("SETEX key seconds value - SET avec expiration automatique en secondes")
	case "DEL":
		return protocolEncoder.WriteSimpleStringResponse("DEL key [key ...] - Supprime une ou plusieurs cles")
	case "EXISTS":
		return protocolEncoder.WriteSimpleStringResponse("EXISTS key [key ...] - Verifie l'existence de cles")
	case "TYPE":
		return protocolEncoder.WriteSimpleStringResponse("TYPE key - Retourne le type de donnees (string, list, set, hash, none)")
	case "INCR":
		return protocolEncoder.WriteSimpleStringResponse("INCR key - Incremente un compteur de 1")
	case "DECR":
		return protocolEncoder.WriteSimpleStringResponse("DECR key - Decremente un compteur de 1")
	case "INCRBY":
		return protocolEncoder.WriteSimpleStringResponse("INCRBY key increment - Incremente un compteur par la valeur donnee")
	case "DECRBY":
		return protocolEncoder.WriteSimpleStringResponse("DECRBY key decrement - Decremente un compteur par la valeur donnee")
	case "APPEND":
		return protocolEncoder.WriteSimpleStringResponse("APPEND key value - Concatene une valeur a la fin d'une string existante")
	case "STRLEN":
		return protocolEncoder.WriteSimpleStringResponse("STRLEN key - Retourne la longueur d'une string")
	case "GETRANGE", "SUBSTR":
		return protocolEncoder.WriteSimpleStringResponse("GETRANGE key start end - Extrait une sous-chaine (indices negatifs supportes)")
	case "SETRANGE":
		return protocolEncoder.WriteSimpleStringResponse("SETRANGE key offset value - Remplace une partie d'une string a partir de l'offset")
	case "MSET":
		return protocolEncoder.WriteSimpleStringResponse("MSET key value [key value ...] - Definit plusieurs cles en une seule operation")
	case "MGET":
		return protocolEncoder.WriteSimpleStringResponse("MGET key [key ...] - Recupere plusieurs valeurs en une seule operation")
	case "GETSET":
		return protocolEncoder.WriteSimpleStringResponse("GETSET key value - Atomique: recupere l'ancienne valeur et definit la nouvelle")
	case "MSETNX":
		return protocolEncoder.WriteSimpleStringResponse("MSETNX key value [key value ...] - Multi-set si AUCUNE des cles n'existe")
	case "GETDEL":
		return protocolEncoder.WriteSimpleStringResponse("GETDEL key - Atomique: recupere la valeur puis supprime la cle")
	case "TTL":
		return protocolEncoder.WriteSimpleStringResponse("TTL key - Retourne le TTL en secondes (-2=inexistante, -1=pas de TTL)")
	case "PTTL":
		return protocolEncoder.WriteSimpleStringResponse("PTTL key - Retourne le TTL en millisecondes (-2=inexistante, -1=pas de TTL)")
	case "EXPIRE":
		return protocolEncoder.WriteSimpleStringResponse("EXPIRE key seconds - Definit un TTL en secondes sur une cle existante")
	case "PEXPIRE":
		return protocolEncoder.WriteSimpleStringResponse("PEXPIRE key milliseconds - Definit un TTL en millisecondes sur une cle existante")
	case "PERSIST":
		return protocolEncoder.WriteSimpleStringResponse("PERSIST key - Supprime le TTL d'une cle (la rend permanente)")
	case "LPUSH":
		return protocolEncoder.WriteSimpleStringResponse("LPUSH key element [element ...] - Ajoute des elements au debut de la liste")
	case "RPUSH":
		return protocolEncoder.WriteSimpleStringResponse("RPUSH key element [element ...] - Ajoute des elements a la fin de la liste")
	case "LPOP":
		return protocolEncoder.WriteSimpleStringResponse("LPOP key - Retire et retourne le premier element de la liste")
	case "RPOP":
		return protocolEncoder.WriteSimpleStringResponse("RPOP key - Retire et retourne le dernier element de la liste")
	case "LLEN":
		return protocolEncoder.WriteSimpleStringResponse("LLEN key - Retourne la longueur de la liste")
	case "LRANGE":
		return protocolEncoder.WriteSimpleStringResponse("LRANGE key start stop - Retourne une partie de la liste (indices, -1 = dernier)")
	case "LSET":
		return protocolEncoder.WriteSimpleStringResponse("LSET key index element - Definit un element a un index specifique")
	case "LREM":
		return protocolEncoder.WriteSimpleStringResponse("LREM key count element - Supprime count occurrences d'un element (0=toutes)")
	case "LINSERT":
		return protocolEncoder.WriteSimpleStringResponse("LINSERT key BEFORE|AFTER pivot element - Insere un element avant/apres un pivot")
	case "LTRIM":
		return protocolEncoder.WriteSimpleStringResponse("LTRIM key start stop - Garde seulement les elements dans la plage donnee")
	case "SADD":
		return protocolEncoder.WriteSimpleStringResponse("SADD key member [member ...] - Ajoute des membres uniques a un set")
	case "SMEMBERS":
		return protocolEncoder.WriteSimpleStringResponse("SMEMBERS key - Retourne tous les membres d'un set")
	case "SISMEMBER":
		return protocolEncoder.WriteSimpleStringResponse("SISMEMBER key member - Teste si un membre appartient au set (retourne 1 ou 0)")
	case "SREM":
		return protocolEncoder.WriteSimpleStringResponse("SREM key member [member ...] - Supprime des membres d'un set")
	case "SCARD":
		return protocolEncoder.WriteSimpleStringResponse("SCARD key - Retourne le nombre de membres dans un set")
	case "SDIFF":
		return protocolEncoder.WriteSimpleStringResponse("SDIFF key [key ...] - Calcule la difference entre sets (premier - autres)")
	case "SINTER":
		return protocolEncoder.WriteSimpleStringResponse("SINTER key [key ...] - Calcule l'intersection de plusieurs sets")
	case "SUNION":
		return protocolEncoder.WriteSimpleStringResponse("SUNION key [key ...] - Calcule l'union de plusieurs sets")
	case "HSET":
		return protocolEncoder.WriteSimpleStringResponse("HSET key field value [field value ...] - Definit des champs dans un hash")
	case "HGET":
		return protocolEncoder.WriteSimpleStringResponse("HGET key field - Recupere la valeur d'un champ dans un hash")
	case "HGETALL":
		return protocolEncoder.WriteSimpleStringResponse("HGETALL key - Retourne tous les champs et valeurs d'un hash")
	case "HEXISTS":
		return protocolEncoder.WriteSimpleStringResponse("HEXISTS key field - Teste si un champ existe dans un hash")
	case "HDEL":
		return protocolEncoder.WriteSimpleStringResponse("HDEL key field [field ...] - Supprime des champs d'un hash")
	case "HLEN":
		return protocolEncoder.WriteSimpleStringResponse("HLEN key - Retourne le nombre de champs dans un hash")
	case "HKEYS":
		return protocolEncoder.WriteSimpleStringResponse("HKEYS key - Retourne tous les noms de champs d'un hash")
	case "HVALS":
		return protocolEncoder.WriteSimpleStringResponse("HVALS key - Retourne toutes les valeurs d'un hash")
	case "HINCRBY":
		return protocolEncoder.WriteSimpleStringResponse("HINCRBY key field increment - Incremente un champ entier dans un hash")
	case "HINCRBYFLOAT":
		return protocolEncoder.WriteSimpleStringResponse("HINCRBYFLOAT key field increment - Incremente un champ flottant dans un hash")
	case "SAVE":
		return protocolEncoder.WriteSimpleStringResponse("SAVE - Sauvegarde synchrone (bloquante) des donnees sur disque")
	case "BGSAVE":
		return protocolEncoder.WriteSimpleStringResponse("BGSAVE - Sauvegarde asynchrone (arriere-plan) des donnees sur disque")
	case "LASTSAVE":
		return protocolEncoder.WriteSimpleStringResponse("LASTSAVE - Retourne le timestamp Unix de la derniere sauvegarde")
	case "INFO":
		return protocolEncoder.WriteSimpleStringResponse("INFO [section] - Informations sur le serveur (sections: server, persistence, memory)")
	case "PING":
		return protocolEncoder.WriteSimpleStringResponse("PING [message] - Test de connexion. Retourne PONG ou le message")
	case "ECHO":
		return protocolEncoder.WriteSimpleStringResponse("ECHO message - Retourne le message tel quel")
	case "KEYS":
		return protocolEncoder.WriteSimpleStringResponse("KEYS pattern - Recherche des cles par motif (* = tout, ? = 1 char, [abc] = choix)")
	case "DBSIZE":
		return protocolEncoder.WriteSimpleStringResponse("DBSIZE - Retourne le nombre total de cles dans la base")
	case "FLUSHALL":
		return protocolEncoder.WriteSimpleStringResponse("FLUSHALL - Vide completement la base de donnees")
	default:
		return protocolEncoder.WriteSimpleStringResponse("Commande inconnue. Tapez ALAIDE pour voir toutes les commandes disponibles")
	}
}
