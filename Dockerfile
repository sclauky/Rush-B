# On installe l'image officielle de Go pour commencer le projet.
FROM golang:1.25

# Installe le gcc pour compiler le go avec CGO_ENABLED=1 et sqlite3 pour ouvrir la base de données.
RUN apt-get update && apt-get install -y gcc libc6-dev sqlite3

# ça definit le repertoire de travail dans l'image pour que toutes les commandes soient faites ici.
WORKDIR /app 

# On copie le go mod et go sum pour pouvoir telecharger les bonnes dépendances avec la prochaine commande.
COPY go.mod go.sum ./
# Télécharge les dépendances du projet. 
RUN go mod download 

# Copie la racine du projet dans le conteneur.
COPY . . 

# Active CGO pour activer la compilation Go.
ENV CGO_ENABLED=1 

# Compile le projet et créer le binaire nommé 'forum-app' à la racine. Etape faite au moment de la création de l'image.
RUN go build -o forum-app main.go 

# Definition du port.
EXPOSE 8080 

# Lance le binaire 'forum-app' quand le conteneur démarre. Quand tu fais 'docker run forum-app', Docker exécute cette commande par défaut.
CMD ["./forum-app"] 