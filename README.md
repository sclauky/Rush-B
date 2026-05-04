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

## lancement du site

1) Télécharger msys2 depuis ce lien : [https://www.msys2.org/](msys2) et l'installé
2) Cliquer sur Démarrer et chercher variables d'environnement, cliquer sur "modifier les variables d'environnement système" 
3) Cliquez sur Variables d'environement
4) Puis sur PATH dans variable système, puis cliquez sur modifier*
5) Puis sur nouvelle et ajoutez le chemin d'acces de où msys2 est installé, en général c'est : C:\msys64\ucrt64\bin

go run main.go
