# Utilisation de l'image officielle Golang
FROM golang:1.24-alpine

# Installation des dépendances nécessaires
RUN apk add --no-cache gcc musl-dev

# Définition du répertoire de travail
WORKDIR /app

# Copie des fichiers du projet
COPY . .

# Compilation de l'application
RUN go mod tidy && go build -o server ./server/main.go

# Exposition du port HTTPS
EXPOSE 8080

# Commande pour lancer l'application
CMD ["/app/server/main"]
