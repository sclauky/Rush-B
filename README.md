# README

## Prérequis

Avant de lancer le projet, il faut au minimum :

- **Git** pour cloner le dépôt et changer de branche si besoin.
- **Docker Desktop** installé et démarré, avec `docker compose` disponible.
- Un terminal ouvert à la racine du projet.

Vous pouvez vérifier rapidement que tout est bien installé avec :

```bash
git --version
docker --version
docker compose version
```

## Comment lancer le projet

Démarrez l'application sur votre ordinateur Docker Desktop que vous avez téléchargé au préalable (voir la partie "prérequis")

Placez-vous à la racine du projet, puis lancez :

```bash
docker compose down
docker compose up --build -d
```

En suite, il faut penser à lancer npm pour faire fonctionner les messages privés. Pour ce faire:

```
cd .\backend\overall-chat\
```

Puis lancer NPM

```
npm run dev
```

## Ouvrir le projet

### En HTTPS

Ouvre cette URL dans votre navigateur :

```text
https://rushb.localhost
```

### En HTTP

```text
http://rushb.localhost
```

## Si ça ne marche pas

```bash
docker compose logs app --tail=100
docker compose logs caddy --tail=100
```

La commande `docker compose logs` permet d'afficher les logs d'un service, et `--tail` limite le nombre de lignes affichées.

## Arborescence du projet

```text
C:.
├───.git/
├───backend
│   ├───database
│   │   └───schema.sql
│   └───forum
│       ├───auth
│       │   ├───db
│       │   │   └───db.go
│       │   ├───handlers
│       │   │   ├───auth.go
│       │   │   └───twofactor.go
│       │   ├───middleware
│       │   │   └───auth.go
│       │   ├───models
│       │   │   └───user.go
│       │   └───utils
│       │       ├───email.go
│       │       ├───jwt.go
│       │       ├───password.go
│       │       └───session.go
│       ├───comment
│       │   ├───db
│       │   │   └───comment_db.go
│       │   ├───handlers
│       │   │   └───comment.go
│       │   └───models
│       │       └───comment.go
│       ├───like
│       │   ├───db
│       │   │   └───like_db.go
│       │   └───handlers
│       │       └───like.go
│       ├───overall-chat
│       │   ├───package.json
│       │   └───src
│       │       ├───controllers
│       │       │   └───auth.controller.js
│       │       ├───index.js
│       │       ├───lib
│       │       │   ├───db.js
│       │       │   └───utils.js
│       │       ├───models
│       │       │   └───user.model.js
│       │       └───routes
│       │           └───auth.route.js
│       ├───post
│       │   ├───db
│       │   │   └───post_db.go
│       │   ├───handlers
│       │   │   └───post.go
│       │   ├───models
│       │   │   └───post.go
│       │   └───post.go
│       ├───search
│       │   └───handler.go
│       └───swipe
│           └───handler.go
├───frontend
│   ├───overall-chat
│   │   ├───public
│   │   ├───src
│   │   │   ├───App.css
│   │   │   ├───App.jsx
│   │   │   ├───index.css
│   │   │   ├───main.jsx
│   │   │   └───assets
│   │   ├───eslint.config.js
│   │   ├───index.html
│   │   ├───package.json
│   │   ├───README.md
│   │   ├───tmp.html
│   │   └───vite.config.js
│   └───web
│       ├───static
│       │   ├───css
│       │   ├───images
│       │   └───js
│       ├───templates
│       │   ├───404.html
│       │   ├───500.html
│       │   ├───base.html
│       │   ├───create-post.html
│       │   ├───edit-profile.html
│       │   ├───f2a-mail.html
│       │   ├───home.html
│       │   ├───loadingscreen.html
│       │   ├───login.html
│       │   ├───post-detail.html
│       │   ├───posts.html
│       │   ├───profile.html
│       │   ├───register.html
│       │   ├───search.html
│       │   ├───swipe.html
│       │   ├───user-profile.html
│       │   └───verify-2fa.html
│       └───uploads
│           └───profile
├───package
│   ├───LICENSE
│   ├───package.json
│   ├───rollup.config.js
│   └───src
│       └───...
├───postman
│   ├───collections
│   ├───documents
│   ├───environments
│   ├───flows
│   ├───globals
│   │   └───workspace.globals.yaml
│   ├───mocks
│   └───specs
├───Caddyfile
├───Dockerfile
├───README.md
├───compose.yaml
├───go.mod
├───main.go
├───ogl-0.0.36.tgz
├───settings.json
├───root.crt
├───socialauth.exe
├───.dockerignore
├───.env
├───.gitignore
└───.postman
```