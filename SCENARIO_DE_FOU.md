# Scénarios de Test - Démo Redis-Go

## 🚀 BGSAVE AU DÉBUT
```bash
BGSAVE
# Réponse: "Background saving started"
```

---

## 📝 STRINGS & COMPTEURS

### SET / GET / DEL
```bash
SET welcome "Hello Redis-Go!" EX 60
GET welcome
# Réponse: "Hello Redis-Go!"

DEL welcome
# Réponse: (integer) 1

GET welcome
# Réponse: (nil)
```

### INCR / DECR / INCRBY
```bash
SET counter "10"
INCR counter
# Réponse: (integer) 11

INCRBY counter 5
# Réponse: (integer) 16

DECR counter
# Réponse: (integer) 15
```

### GETSET / GETDEL (atomiques)
```bash
SET user:score "100"
GETSET user:score "200"
# Réponse: "100" (ancienne valeur)

GET user:score
# Réponse: "200"

GETDEL user:score
# Réponse: "200"
# La clé est supprimée automatiquement
```

### APPEND / STRLEN
```bash
SET message "Hello"
APPEND message " World!"
# Réponse: (integer) 12

STRLEN message
# Réponse: (integer) 12
```

---

## 📋 LISTES

### LPUSH / RPUSH / LPOP / RPOP
```bash
LPUSH tasks "task1" "task2" "task3"
# Réponse: (integer) 3

LRANGE tasks 0 -1
# Réponse: 1) "task3" 2) "task2" 3) "task1"

RPUSH tasks "task4"
# Réponse: (integer) 4

LPOP tasks
# Réponse: "task3"

RPOP tasks
# Réponse: "task4"
```

### LSET / LREM / LINSERT
```bash
RPUSH fruits "apple" "banana" "apple" "orange"
LSET fruits 1 "mango"
# Réponse: OK

LREM fruits 1 "apple"
# Réponse: (integer) 1 (supprime 1 occurrence)

LINSERT fruits BEFORE "orange" "kiwi"
# Réponse: (integer) 4 (nouvelle longueur)

LRANGE fruits 0 -1
# Réponse: 1) "mango" 2) "apple" 3) "kiwi" 4) "orange"
```

### LTRIM
```bash
RPUSH numbers "1" "2" "3" "4" "5"
LTRIM numbers 1 3
# Réponse: OK

LRANGE numbers 0 -1
# Réponse: 1) "2" 2) "3" 3) "4"
```

---

## 🔗 SETS

### SADD / SMEMBERS / SISMEMBER
```bash
SADD users:online "alice" "bob" "charlie"
# Réponse: (integer) 3

SMEMBERS users:online
# Réponse: 1) "alice" 2) "bob" 3) "charlie"

SISMEMBER users:online "alice"
# Réponse: (integer) 1

SISMEMBER users:online "david"
# Réponse: (integer) 0
```

### SREM / SCARD
```bash
SREM users:online "bob"
# Réponse: (integer) 1

SCARD users:online
# Réponse: (integer) 2
```

### SINTER / SUNION / SDIFF (opérations ensemblistes)
```bash
SADD users:premium "alice" "david" "eve"
SADD users:online "alice" "charlie"

SINTER users:premium users:online
# Réponse: 1) "alice" (dans les deux sets)

SUNION users:premium users:online
# Réponse: 1) "alice" 2) "david" 3) "eve" 4) "charlie"

SDIFF users:premium users:online
# Réponse: 1) "david" 2) "eve" (dans premium mais pas online)
```

---

## 🗃️ HASHES

### HSET / HGET / HGETALL
```bash
HSET user:123 name "Alice" age "25" city "Paris"
# Réponse: (integer) 3

HGET user:123 name
# Réponse: "Alice"

HGETALL user:123
# Réponse: 1) "name" 2) "Alice" 3) "age" 4) "25" 5) "city" 6) "Paris"
```

### HEXISTS / HDEL / HLEN
```bash
HEXISTS user:123 age
# Réponse: (integer) 1

HDEL user:123 city
# Réponse: (integer) 1

HLEN user:123
# Réponse: (integer) 2
```

### HINCRBY / HINCRBYFLOAT
```bash
HSET game:stats score "100" level "5.5"
HINCRBY game:stats score 50
# Réponse: (integer) 150

HINCRBYFLOAT game:stats level 0.5
# Réponse: "6"
```

### HKEYS / HVALS
```bash
HKEYS user:123
# Réponse: 1) "name" 2) "age"

HVALS user:123
# Réponse: 1) "Alice" 2) "25"
```

---

## 🔍 RECHERCHE & UTILITAIRES

### KEYS (pattern matching)
```bash
SET user:alice "data1"
SET user:bob "data2"
SET session:123 "data3"

KEYS *
# Réponse: 1) "user:alice" 2) "user:bob" 3) "session:123"

KEYS user:*
# Réponse: 1) "user:alice" 2) "user:bob"

KEYS user:?????
# Réponse: 1) "user:alice" (exactement 5 caractères après "user:")
```

### EXISTS / TYPE
```bash
EXISTS user:alice session:123 nonexistent
# Réponse: (integer) 2

TYPE user:alice
# Réponse: string

TYPE users:online
# Réponse: set
```

---

## ⏰ TTL & EXPIRATION

### EXPIRE / TTL / PERSIST
```bash
SET temp:data "important"
EXPIRE temp:data 30
# Réponse: (integer) 1

TTL temp:data
# Réponse: (integer) 27 (secondes restantes)

PERSIST temp:data
# Réponse: (integer) 1

TTL temp:data
# Réponse: (integer) -1 (pas de TTL)
```

### PEXPIRE / PTTL
```bash
PEXPIRE temp:data 5000
# Réponse: (integer) 1

PTTL temp:data
# Réponse: (integer) 4856 (millisecondes restantes)
```

---

## 💾 PERSISTENCE & INFO

### SAVE / BGSAVE / LASTSAVE
```bash
SAVE
# Réponse: OK (sauvegarde synchrone)

BGSAVE
# Réponse: "Background saving started"

LASTSAVE
# Réponse: (integer) 1704088800 (timestamp Unix)
```

### INFO / DBSIZE
```bash
INFO persistence
# Réponse: détails sur RDB (rdb_last_save_time, rdb_total_saves...)

DBSIZE
# Réponse: (integer) 15 (nombre de clés)
```

---

## 🔧 UTILITAIRES

### PING / ECHO
```bash
PING
# Réponse: PONG

PING "Hello!"
# Réponse: "Hello!"

ECHO "Test message"
# Réponse: "Test message"
```

### ALAIDE (aide)
```bash
ALAIDE
# Réponse: liste toutes les commandes disponibles

ALAIDE SET
# Réponse: "SET key value [EX seconds] - Stocke une valeur avec TTL optionnel"
```

---

## 🎯 SCÉNARIO COMPLET DE DÉMO

```bash
# 1. Sauvegarde initiale
BGSAVE

# 2. Création d'un utilisateur complet
HSET user:alice name "Alice" score "100" level "5"
SET user:alice:session "abc123" EX 60

# 3. Gestion d'une file de tâches
LPUSH tasks "send_email" "backup_db" "cleanup"
LPOP tasks

# 4. Système de groupes
SADD admins alice bob
SADD users alice charlie david
SINTER admins users

# 5. Statistiques de jeu
HINCRBY user:alice score 50
HINCRBYFLOAT user:alice level 0.5

# 6. Recherche et nettoyage
KEYS user:*
DBSIZE
INFO persistence
```