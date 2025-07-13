# Dockerfile pour Redis-Go

# Build stage
FROM golang:1.24-alpine AS builder

# Installation des dépendances de build
RUN apk add --no-cache git

# Répertoire de travail
WORKDIR /app

# Copie des fichiers de dépendances
COPY go.mod ./

# Téléchargement des dépendances
RUN go mod download

# Copie du code source
COPY . .

# Compilation du binaire
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o redis-go main.go

# Runtime stage
FROM alpine:latest

# Installation de ca-certificates pour HTTPS si besoin
RUN apk --no-cache add ca-certificates

# Création d'un utilisateur non-root pour la sécurité
RUN adduser -D -s /bin/sh redis

# Répertoire de travail
WORKDIR /app

# Copie du binaire depuis le stage de build
COPY --from=builder /app/redis-go .

# Création des dossiers nécessaires
RUN mkdir -p logs data && chown -R redis:redis logs data

# Changement vers l'utilisateur non-root
USER redis

# Port exposé
EXPOSE 6379

# Variables d'environnement par défaut
ENV REDIS_HOST=0.0.0.0
ENV REDIS_PORT=6379
ENV REDIS_MAX_CONNECTIONS=1000
ENV REDIS_RDB_ENABLED=true
ENV REDIS_RDB_FILE=./data/dump.rdb
ENV REDIS_RDB_SAVE_INTERVAL=300
ENV REDIS_RDB_SAVE_ON_EXIT=true

# Commande par défaut
CMD ["./redis-go"]