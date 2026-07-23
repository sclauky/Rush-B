# README

## Prérequis

Avant de lancer le projet, il faut au minimum :

- **Git** pour cloner le dépôt et changer de branche si besoin.
- **Docker Desktop** installé et démarré, avec `docker compose` disponible.
- Un terminal ouvert à la racine du projet.

Tu peux vérifier rapidement que tout est bien installé avec :

```bash
git --version
docker --version
docker compose version
```

## Comment lancer le projet

Démarrez Docker Desktop.

Placez-vous à la racine du projet, puis lancez :

```bash
docker compose down
docker compose up --build -d
```

Ensuite, vérifiez que tout tourne :

```bash
docker compose ps
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
├───backend
│   ├───database
│   └───forum
│       ├───auth
│       │   ├───db
│       │   ├───handlers
│       │   ├───middleware
│       │   ├───models
│       │   └───utils
│       ├───comment
│       │   ├───db
│       │   ├───handlers
│       │   └───models
│       ├───like
│       │   ├───db
│       │   └───handlers
│       ├───post
│       │   ├───db
│       │   ├───handlers
│       │   └───models
│       └───search
└───frontend
    └───web
        ├───static
        │   ├───css
        │   ├───images
        │   └───js
        ├───templates
        └───uploads
            └───profile
```