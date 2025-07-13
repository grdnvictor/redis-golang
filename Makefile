# Redis-Go Makefile - Version Docker uniquement
.PHONY: build run down restart \
        test test-auto \
        logs cli status \
        clean fmt deps \
        help
# Variables
COMPOSE_COMMAND = docker compose
REDIS_SERVICE = redis-go
REDIS_CLI_SERVICE = redis-cli
REDIS_TEST_SERVICE = redis-test

# Commandes principales
build:
	@echo "🔨 Construction de l'image Docker..."
	@$(COMPOSE_COMMAND) build $(REDIS_SERVICE) > /dev/null 2>&1
	@echo "✅ Image construite"

run: build
	@echo "🚀 Démarrage du serveur Redis-Go..."
	@$(COMPOSE_COMMAND) up -d $(REDIS_SERVICE) > /dev/null 2>&1
	@echo "✅ Serveur démarré en arrière-plan"
	@echo "💡 Utilisez 'make logs' pour voir les logs"
	@echo "💡 Utilisez 'make cli' pour vous connecter"


test:
	@echo "🧪 Exécution des tests..."
	@$(COMPOSE_COMMAND) run --rm $(REDIS_SERVICE) go test -v ./... > /dev/null 2>&1
	@echo "✅ Tests terminés"

down:
	@if $(COMPOSE_COMMAND) ps -q | grep -q .; then \
		echo "🧹 Nettoyage des conteneurs..."; \
		$(COMPOSE_COMMAND) down -v > /dev/null 2>&1; \
		echo "✅ Nettoyage terminé"; \
	else \
		echo "ℹ️ Aucun conteneur à nettoyer"; \
	fi

logs:
	@echo "📋 Logs du serveur Redis-Go..."
	@$(COMPOSE_COMMAND) logs -f $(REDIS_SERVICE)

# Session interactive redis-cli
cli:
	@echo "🐳 Préparation de redis-cli..."
	@$(COMPOSE_COMMAND) up -d $(REDIS_CLI_SERVICE) > /dev/null 2>&1
	@echo "✅ Container redis-cli prêt"
	@echo "🆘 Tapez 'ALAIDE' pour voir toutes les commandes disponibles"
	@echo "💡 Tapez 'exit' ou Ctrl+C pour quitter"
	@echo "🔗 Connexion à Redis-Go..."
	@sleep 1
	@$(COMPOSE_COMMAND) exec -it $(REDIS_CLI_SERVICE) redis-cli -h $(REDIS_SERVICE) -p 6379 --no-auth-warning

# Tests automatisés complets
test-auto:
	@echo "🧪 Lancement des tests automatisés..."
	@$(COMPOSE_COMMAND) up --build $(REDIS_TEST_SERVICE) > /dev/null 2>&1
	@echo "✅ Tests automatisés terminés"

# Formatage du code
fmt:
	@echo "📝 Formatage du code..."
	@$(COMPOSE_COMMAND) run --rm $(REDIS_SERVICE) go fmt ./... > /dev/null 2>&1
	@echo "✅ Code formaté"

# Redémarrage complet
restart: down run

# Statut des services
status:
	@echo "📊 Statut des services:"
	@$(COMPOSE_COMMAND) ps --format "table {{.Name}}\t{{.Image}}\t{{.Service}}\t{{.Status}}\t{{.Ports}}"

# Mise à jour des dépendances Go
deps:
	@echo "📦 Mise à jour des dépendances..."
	@$(COMPOSE_COMMAND) run --rm $(REDIS_SERVICE) go mod tidy > /dev/null 2>&1
	@echo "✅ Dépendances mises à jour"

# Aide
help:
	@echo "🎯 Redis-Go - Commandes Docker disponibles :"
	@echo ""
	@echo "📦 Gestion du serveur :"
	@echo "  run       - Démarre le serveur Redis-Go (en arrière-plan)"
	@echo "  down      - Arrête tous les services"
	@echo "  restart   - Redémarre le serveur"
	@echo "  logs      - Affiche les logs du serveur"
	@echo "  status    - Statut des services"
	@echo ""
	@echo "🔧 Développement :"
	@echo "  build     - Build l'image Docker"
	@echo "  test      - Lance les tests unitaires"
	@echo "  fmt       - Formate le code"
	@echo "  deps      - Met à jour les dépendances"
	@echo ""
	@echo "🎮 Utilisation :"
	@echo "  cli       - Redis-cli interactif"
	@echo "  test-auto - Tests automatisés complets"
	@echo ""
	@echo "🚀 Démarrage rapide :"
	@echo "  1. make run     # Démarre le serveur"
	@echo "  2. make cli     # Se connecte avec redis-cli"
	@echo "  3. ALAIDE       # Voir toutes les commandes"
	@echo "  4. ALAIDE SET   # Aide détaillée pour SET"