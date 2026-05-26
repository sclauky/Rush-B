# README

## Ce qu'il faut avoir

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

Place-toi à la racine du projet, puis lance :

```bash
git checkout integration/rebuild-clean
docker compose down
docker compose up --build -d
```

`docker compose up --build -d` permet de reconstruire puis démarrer les services en arrière-plan.

Ensuite, vérifie que tout tourne :

```bash
docker compose ps
```

## Ouvrir le projet

### En HTTPS

Ouvre cette URL dans ton navigateur :

```text
https://rushb.localhost
```

Caddy peut servir automatiquement du HTTPS local pour `localhost` et les domaines en `.localhost`, selon sa configuration.

### En HTTP

Si tu veux le lancer en HTTP, ouvre :

```text
http://rushb.localhost
```

## Si ça ne marche pas

Regarde les logs des services :

```bash
docker compose logs app --tail=100
docker compose logs caddy --tail=100
```

La commande `docker compose logs` permet d'afficher les logs d'un service, et `--tail` limite le nombre de lignes affichées.

## Arborescence du projet

Colle l'arborescence ici :

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