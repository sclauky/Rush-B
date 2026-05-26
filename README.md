# Rush-B

## Intro

Notre Projet **_Rush B_** est un forum. Il a pour but de mettre en lien des joueurs afin de trouver des gens ayant les mêmes centres d'intérêt, des équipiers de jeux ou, tout simplement, échanger avec d'autres utilisateurs.

## Arborescence

```bash
Rush-B/
│   .dockerignore
│   .gitignore
│   compose.yaml
│   Dockerfile
│   go.mod
│   go.sum
│   main.go
│   README.md
│   socialauth.exe
│
├───backend
│   ├───database
│   │       data.db
│   │       schema.sql
│   │
│   └───forum
│       ├───auth
│       │   ├───db
│       │   │       db.go
│       │   │
│       │   ├───handlers
│       │   │       auth.go
│       │   │       twofactor.go
│       │   │
│       │   ├───middleware
│       │   │       auth.go
│       │   │
│       │   ├───models
│       │   │       user.go
│       │   │
│       │   └───utils
│       │           email.go
│       │           jwt.go
│       │           password.go
│       │           session.go
│       │
│       ├───comment
│       │   ├───db
│       │   │       comment_db.go
│       │   │
│       │   ├───handlers
│       │   │       comment.go
│       │   │
│       │   └───models
│       │           comment.go
│       │
│       ├───like
│       │   ├───db
│       │   │       like_db.go
│       │   │
│       │   └───handlers
│       │           like.go
│       │
│       ├───post
│       │   │   post.go
│       │   │
│       │   ├───db
│       │   │       post_db.go
│       │   │
│       │   ├───handlers
│       │   │       post.go
│       │   │
│       │   └───models
│       │           post.go
│       │
│       └───search
│               handler.go
│
└───frontend
    │   go.mod
    │   go.sum
    │
    └───web
        ├───static
        │   ├───css
        │   │       chat-swipe.css
        │   │       components.css
        │   │       create-post.css
        │   │       error.css
        │   │       f2a-mail.css
        │   │       post-detail.css
        │   │       profile.css
        │   │       responsive.css
        │   │       search.css
        │   │       style.css
        │   │       swipe.css
        │   │       verify-2fa.css
        │   │
        │   ├───images
        │   │       default-avatar.png
        │   │
        │   └───js
        │           chat-swipe.js
        │           edit-profile.js
        │           main.js
        │           post-detail.js
        │           posts.js
        │           profile.js
        │           search.js
        │           swipe.js
        │           verify-2fa.js
        │
        ├───templates
        │       404.html
        │       500.html
        │       base.html
        │       chat-swipe.html
        │       create-post.html
        │       edit-profile.html
        │       f2a-mail.html
        │       home.html
        │       login.html
        │       post-detail.html
        │       posts.html
        │       profile.html
        │       register.html
        │       search.html
        │       swipe.html
        │       user-profile.html
        │       verify-2fa.html
        │
        └───uploads
            └───profile
                    1_1779119276_clement kirk.png
```

---

## Lancement du projet avec Docker

### 1. Prérequis

- Docker Desktop installé et apres avoir lancé l'application (Windows / macOS / Linux)

### 2. Construction et lancement

Depuis la **racine du projet** (là où se trouvent `Dockerfile` et `compose.yaml`), exécuter :

```
docker compose up --build
```

- `--build` force la reconstruction de l’image à partir du `Dockerfile`, ce qui est utile après une modification du code ou de la configuration.
  -> Cette commande construit l’image, crée le conteneur et lance l’application avec la configuration définie dans `compose.yaml`.

### 3. Accès au site

Une fois le conteneur démarré, le site est accessible depuis la machine hôte à l’adresse suivante :

```text
http://localhost:8080
```

Si le port `8080` est publié dans `compose.yaml`, le site peut aussi être ouvert depuis une autre machine du même réseau local avec :

```text
http://ADRESSE_IP_DE_LA_MACHINE_HOTE:8080
```

La publication des ports permet de rendre un service du conteneur accessible depuis l’extérieur du conteneur. [8][9]

### 4. Lancement en arrière-plan

Pour démarrer le projet sans bloquer le terminal :

```bash
docker compose up -d --build
```

L’option `-d` lance les conteneurs en arrière-plan. [5]
