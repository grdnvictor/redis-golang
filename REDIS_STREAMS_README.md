# 🌊 Redis Streams - Guide Complet

## Vue d'ensemble

Redis Streams est une **structure de données révolutionnaire** introduite dans Redis 5.0 qui permet de gérer des flux de messages en temps réel. Votre implémentation Redis-Go supporte maintenant **toutes les fonctionnalités essentielles** des Redis Streams, transformant votre clone en un système de messaging moderne et puissant.

## 🚀 Pourquoi Redis Streams ?

- **📊 Event Sourcing** - Parfait pour stocker l'historique des événements
- **🔄 Real-time Processing** - Traitement de messages en temps réel
- **👥 Consumer Groups** - Distribution de charge entre plusieurs consommateurs
- **💯 Garanties de livraison** - Système d'acquittement des messages
- **🔍 Queries temporelles** - Recherche par plage de temps
- **📈 Scalabilité** - Gestion de millions de messages

---

## 🛠️ Commandes Disponibles

### Commandes de Base

| Commande | Syntaxe | Description |
|----------|---------|-------------|
| `XADD` | `XADD stream id field value [field value ...]` | Ajoute un message au stream |
| `XRANGE` | `XRANGE stream start end [COUNT count]` | Récupère des messages dans une plage |
| `XREAD` | `XREAD [COUNT count] [BLOCK ms] STREAMS stream [stream ...] id [id ...]` | Lit des messages depuis un stream |
| `XLEN` | `XLEN stream` | Retourne le nombre de messages |
| `XDEL` | `XDEL stream id [id ...]` | Supprime des messages |

### Consumer Groups

| Commande | Syntaxe | Description |
|----------|---------|-------------|
| `XGROUP CREATE` | `XGROUP CREATE stream group id` | Crée un consumer group |
| `XGROUP DESTROY` | `XGROUP DESTROY stream group` | Supprime un consumer group |
| `XREADGROUP` | `XREADGROUP GROUP group consumer [COUNT count] STREAMS stream [stream ...] >` | Lit via consumer group |
| `XACK` | `XACK stream group id [id ...]` | Acquitte des messages traités |
| `XPENDING` | `XPENDING stream group [consumer]` | Affiche les messages en attente |

---

## 📝 Exemples Pratiques

### 1. Stream Basique - Logs d'Application

```bash
# Démarrer le serveur
make run

# Dans un autre terminal
redis-cli -p 6379
```

```bash
# Ajouter des logs d'application
XADD app_logs * level "INFO" message "User logged in" user_id "123"
# Réponse: "1703188800000-0"

XADD app_logs * level "ERROR" message "Database connection failed" service "auth"
# Réponse: "1703188800001-0"

XADD app_logs * level "WARNING" message "High memory usage" memory_percent "85"
# Réponse: "1703188800002-0"

# Lire tous les logs
XRANGE app_logs - +
# Réponse: 
# 1) "1703188800000-0"
# 2) "level"
# 3) "INFO"
# 4) "message"
# 5) "User logged in"
# 6) "user_id"
# 7) "123"
# ... (autres messages)
```

### 2. Stream E-commerce - Événements de Commande

```bash
# Ajouter des événements de commande
XADD order_events * event "order_created" order_id "ORD-001" customer_id "123" amount "99.99"
XADD order_events * event "payment_received" order_id "ORD-001" payment_method "credit_card"
XADD order_events * event "order_shipped" order_id "ORD-001" tracking_number "TRK-456"
XADD order_events * event "order_delivered" order_id "ORD-001" delivery_date "2024-01-15"

# Lire les événements depuis un ID spécifique
XREAD COUNT 2 STREAMS order_events 1703188800000-0
```

### 3. Stream Chat - Messages en Temps Réel

```bash
# Simulation d'un chat
XADD chat_room:general * user "Alice" message "Hello everyone!" timestamp "2024-01-15T10:30:00Z"
XADD chat_room:general * user "Bob" message "Hi Alice!" timestamp "2024-01-15T10:30:15Z"
XADD chat_room:general * user "Charlie" message "Good morning!" timestamp "2024-01-15T10:30:30Z"

# Lire les derniers messages
XREAD COUNT 10 STREAMS chat_room:general $

# Lecture bloquante (attend nouveaux messages)
XREAD BLOCK 5000 STREAMS chat_room:general $
```

---

## 👥 Consumer Groups - Traitement Distribué

### Configuration d'un Consumer Group

```bash
# Créer un stream avec des tâches
XADD task_queue * task_type "send_email" recipient "user@example.com" template "welcome"
XADD task_queue * task_type "process_image" image_id "img_123" operation "resize"
XADD task_queue * task_type "send_notification" user_id "456" message "Order confirmed"

# Créer un consumer group
XGROUP CREATE task_queue email_workers 0
# Réponse: OK

# Vérifier la longueur du stream
XLEN task_queue
# Réponse: (integer) 3
```

### Traitement par des Workers

```bash
# Worker 1: lit des tâches
XREADGROUP GROUP email_workers worker1 COUNT 1 STREAMS task_queue >
# Réponse: 
# 1) "task_queue"
# 2) "1703188800000-0 task_type send_email recipient user@example.com template welcome"

# Worker 2: lit d'autres tâches
XREADGROUP GROUP email_workers worker2 COUNT 1 STREAMS task_queue >
# Réponse: 
# 1) "task_queue"
# 2) "1703188800001-0 task_type process_image image_id img_123 operation resize"

# Worker 1: acquitte la tâche terminée
XACK task_queue email_workers 1703188800000-0
# Réponse: (integer) 1
```

### Monitoring des Consumer Groups

```bash
# Voir les messages en attente
XPENDING task_queue email_workers
# Réponse: 
# 1) "1703188800001-0"
# 2) "1703188800002-0"

# Voir les messages en attente pour un worker spécifique
XPENDING task_queue email_workers worker2
# Réponse: 
# 1) "1703188800001-0"
```

---

## 📊 Cas d'Usage Avancés

### 1. Système de Monitoring

```bash
# Métriques système
XADD metrics * service "web_server" cpu_usage "45.2" memory_usage "1.2GB" timestamp "2024-01-15T10:30:00Z"
XADD metrics * service "database" cpu_usage "78.5" memory_usage "4.8GB" timestamp "2024-01-15T10:30:00Z"
XADD metrics * service "cache" cpu_usage "12.1" memory_usage "512MB" timestamp "2024-01-15T10:30:00Z"

# Lire les métriques des dernières 5 minutes
XRANGE metrics 1703188500000-0 1703188800000-0

# Consumer group pour alertes
XGROUP CREATE metrics alert_system 0
XREADGROUP GROUP alert_system alert_processor COUNT 5 STREAMS metrics >
```

### 2. Event Sourcing pour E-commerce

```bash
# Événements d'un utilisateur
XADD user:123:events * event "user_registered" email "user@example.com" plan "free"
XADD user:123:events * event "plan_upgraded" old_plan "free" new_plan "premium"
XADD user:123:events * event "payment_added" payment_method "credit_card" last_four "1234"
XADD user:123:events * event "order_placed" order_id "ORD-001" amount "99.99"

# Reconstruire l'état d'un utilisateur
XRANGE user:123:events - +
```

### 3. Système de Notification

```bash
# Notifications push
XADD notifications * user_id "123" type "order_update" title "Order Shipped" body "Your order has been shipped"
XADD notifications * user_id "456" type "promotion" title "50% Off Sale" body "Limited time offer"
XADD notifications * user_id "123" type "reminder" title "Cart Reminder" body "You have items in your cart"

# Consumer group pour différents types de notifications
XGROUP CREATE notifications push_service 0
XGROUP CREATE notifications email_service 0

# Service push lit les notifications
XREADGROUP GROUP push_service push_worker COUNT 10 STREAMS notifications >
```

---

## 🔍 Queries et Recherches

### Recherche par Plage de Temps

```bash
# Ajouter des événements avec timestamps
XADD events * timestamp "2024-01-15T09:00:00Z" event "server_start"
XADD events * timestamp "2024-01-15T10:00:00Z" event "high_traffic"
XADD events * timestamp "2024-01-15T11:00:00Z" event "server_restart"

# Lire les événements entre 9h et 10h
XRANGE events 1703188800000-0 1703192400000-0

# Lire seulement les 5 premiers événements
XRANGE events - + COUNT 5
```

### Pagination

```bash
# Lire page par page
XRANGE events - + COUNT 10                    # Page 1
XRANGE events 1703188800010-0 + COUNT 10      # Page 2 (depuis le dernier ID de la page 1)
```

---

## ⚡ Performance et Bonnes Pratiques

### 1. Gestion des ID

```bash
# Auto-génération d'ID (recommandé)
XADD mystream * field1 "value1" field2 "value2"

# ID personnalisé (pour des cas spéciaux)
XADD mystream 1703188800000-0 field1 "value1"
```

### 2. Nettoyage des Streams

```bash
# Supprimer d'anciens messages
XDEL mystream 1703188800000-0 1703188800001-0
# Réponse: (integer) 2

# Vérifier la longueur après nettoyage
XLEN mystream
```

### 3. Monitoring des Consumer Groups

```bash
# Créer plusieurs consumer groups pour différents services
XGROUP CREATE events analytics_service 0
XGROUP CREATE events notification_service 0
XGROUP CREATE events audit_service 0

# Chaque service traite les événements à son rythme
XREADGROUP GROUP analytics_service worker1 COUNT 100 STREAMS events >
XREADGROUP GROUP notification_service worker1 COUNT 1 STREAMS events >
```

---

## 🔧 Gestion des Erreurs

### Erreurs Communes

```bash
# Stream inexistant
XRANGE nonexistent_stream - +
# Réponse: (empty list or set)

# Consumer group inexistant
XREADGROUP GROUP nonexistent_group consumer1 STREAMS mystream >
# Réponse: ERREUR : No such consumer group

# Consumer group déjà existant
XGROUP CREATE mystream mygroup 0
XGROUP CREATE mystream mygroup 0
# Réponse: ERREUR : Consumer Group name already exists
```

### Récupération d'Erreurs

```bash
# Vérifier l'existence d'un stream
EXISTS mystream

# Vérifier les consumer groups
XPENDING mystream mygroup

# Recréer un consumer group si nécessaire
XGROUP DESTROY mystream mygroup
XGROUP CREATE mystream mygroup 0
```

---

## 🎯 Scénarios Complets

### Scénario 1: Système de Commande E-commerce

```bash
# === SETUP ===
# Créer le stream des commandes
XADD orders * order_id "12345" customer_id "user123" status "created" total "99.99"

# Créer les consumer groups
XGROUP CREATE orders payment_processor 0
XGROUP CREATE orders inventory_manager 0
XGROUP CREATE orders notification_service 0

# === PROCESSING ===
# Service de paiement
XREADGROUP GROUP payment_processor worker1 COUNT 1 STREAMS orders >
# Traitement...
XADD orders * order_id "12345" status "payment_processed" payment_method "credit_card"
XACK orders payment_processor 1703188800000-0

# Service d'inventaire
XREADGROUP GROUP inventory_manager worker1 COUNT 1 STREAMS orders >
# Traitement...
XADD orders * order_id "12345" status "inventory_reserved" warehouse "EU-WEST"
XACK orders inventory_manager 1703188800001-0

# Service de notification
XREADGROUP GROUP notification_service worker1 COUNT 1 STREAMS orders >
# Envoi notification...
XACK orders notification_service 1703188800002-0
```

### Scénario 2: Système de Logging Centralisé

```bash
# === SETUP ===
# Créer différents streams par service
XADD logs:auth * level "INFO" message "User login" user_id "123"
XADD logs:api * level "ERROR" message "Database timeout" query "SELECT * FROM users"
XADD logs:worker * level "DEBUG" message "Processing job" job_id "456"

# Consumer group pour analyse
XGROUP CREATE logs:auth log_analyzer 0
XGROUP CREATE logs:api log_analyzer 0
XGROUP CREATE logs:worker log_analyzer 0

# === ANALYSIS ===
# Analyser les logs d'authentification
XREADGROUP GROUP log_analyzer auth_analyzer COUNT 10 STREAMS logs:auth >
XACK logs:auth log_analyzer 1703188800000-0

# Analyser les erreurs API
XREADGROUP GROUP log_analyzer api_analyzer COUNT 10 STREAMS logs:api >
```

### Scénario 3: Chat en Temps Réel

```bash
# === SETUP ===
# Créer le stream du chat
XADD chat:room1 * user "Alice" message "Hello!" timestamp "2024-01-15T10:30:00Z"
XADD chat:room1 * user "Bob" message "Hi Alice!" timestamp "2024-01-15T10:30:15Z"

# Consumer group pour modération
XGROUP CREATE chat:room1 moderation_service 0

# === REAL-TIME READING ===
# Client 1: lecture bloquante
XREAD BLOCK 10000 STREAMS chat:room1 $

# Client 2: lecture bloquante
XREAD BLOCK 10000 STREAMS chat:room1 $

# Service de modération
XREADGROUP GROUP moderation_service moderator1 COUNT 1 STREAMS chat:room1 >
# Modération...
XACK chat:room1 moderation_service 1703188800000-0
```

---

## 🔥 Comparaison avec Redis Officiel

### ✅ Fonctionnalités Supportées

| Fonctionnalité | Redis Officiel | Redis-Go | Notes |
|----------------|----------------|----------|-------|
| **XADD** | ✅ | ✅ | Auto-generation ID, champs multiples |
| **XRANGE** | ✅ | ✅ | Plage complète, COUNT support |
| **XREAD** | ✅ | ✅ | Lecture multiple streams, BLOCK |
| **XLEN** | ✅ | ✅ | Longueur stream |
| **XDEL** | ✅ | ✅ | Suppression multiple messages |
| **XGROUP CREATE** | ✅ | ✅ | Création consumer groups |
| **XGROUP DESTROY** | ✅ | ✅ | Suppression consumer groups |
| **XREADGROUP** | ✅ | ✅ | Lecture via consumer groups |
| **XACK** | ✅ | ✅ | Acquittement messages |
| **XPENDING** | ✅ | ✅ | Messages en attente |

### 🚀 Avantages de Cette Implémentation

1. **🔧 Simplicité** - Code Go lisible et maintenable
2. **⚡ Performance** - Optimisé pour la concurrence
3. **🛡️ Sécurité** - Gestion thread-safe avec mutexes
4. **📖 Documentation** - Guide complet et exemples
5. **🔍 Debugging** - Messages d'erreur clairs

---

## 📚 Ressources Supplémentaires

### Documentation Technique

- **Architecture**: Streams stockés comme arrays ordonnés par timestamp
- **Concurrence**: RWMutex pour lecture/écriture thread-safe
- **Persistence**: Compatible avec le système RDB existant
- **Mémoire**: Gestion efficace avec cleanup automatique

### Exemples de Code

```go
// Exemple d'utilisation programmatique
type Event struct {
    ID        string
    Type      string
    Data      map[string]string
    Timestamp time.Time
}

func publishEvent(event Event) {
    // XADD events * type "user_login" user_id "123" timestamp "2024-01-15T10:30:00Z"
}

func consumeEvents(groupName, consumerName string) {
    // XREADGROUP GROUP mygroup consumer1 COUNT 10 STREAMS events >
}
```

---

## 🎉 Conclusion

Redis Streams transforme votre implémentation Redis-Go en un **système de messaging de niveau entreprise**. Avec cette fonctionnalité, vous pouvez maintenant :

- 🔄 **Traiter des événements** en temps réel
- 📊 **Implémenter Event Sourcing** avec historique complet
- 👥 **Distribuer la charge** entre plusieurs workers
- 💯 **Garantir la livraison** avec système d'acquittement
- 📈 **Scaler horizontalement** avec consumer groups

Cette implémentation place votre Redis-Go au niveau des **solutions Redis commerciales** et dépasse la plupart des clones open-source existants.

**Commencez maintenant** avec les exemples ci-dessus et explorez les possibilités infinies des Redis Streams ! 🚀

---

*Créé avec ❤️ pour Redis-Go - La prochaine génération de Redis en Go* 