# Rush-B

## Intro

Notre Projet ***Rush B*** est un forum. Il a pour but de mettre en lien des joueurs afin de trouver des gens ayant les mêmes centres d'intérêt, des équipiers de jeux ou, tout simplement, échanger avec d'autres utilisateurs.

## Arborescence

```bash
Rush-B/
├── go.mod
├── main.go
├── README.md
├── backend/
│   ├── main.go
│   ├── database/
│   │   └── data.db
│   └── forum/
│       ├── auth/
│       │   ├── db/
│       │   │   └── db.go
│       │   ├── handlers/
│       │   │   └── auth.go
│       │   ├── middleware/
│       │   │   └── auth.go
│       │   ├── models/
│       │   │   └── user.go
│       │   └── utils/
│       │       ├── jwt.go
│       │       └── password.go
│       └── post/
│           └── post.go
└── frontend/
    └── web/
        ├── static/
        │   ├── css/
        │   │   ├── components.css
        │   │   ├── responsive.css
        │   │   └── style.css
        │   └── js/
        │       └── main.js
        └── templates/
            ├── 404.html
            ├── 500.html
            ├── base.html
            ├── create-post.html
            ├── home.html
            ├── login.html
            ├── post.html
            ├── posts.html
            ├── profile.html
            └── register.html
```

---

## Lancement du projet avec Docker

### 1. Prérequis

- Docker Desktop installé et en cours d’exécution (Windows / macOS / Linux) 

### 2. Construction et lancement

Depuis la **racine du projet** (là où se trouvent `Dockerfile` et `compose.yaml`), exécuter :

```bash
docker compose up --build
```

- `--build` force la reconstruction de l’image à partir du `Dockerfile`, ce qui est utile après une modification du code ou de la configuration. 
- Cette commande construit l’image, crée le conteneur et lance l’application avec la configuration définie dans `compose.yaml`. 

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