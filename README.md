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

go run main.go 