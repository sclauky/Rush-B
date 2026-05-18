package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"socialauth/backend/forum/post"

	"github.com/gin-gonic/gin"
)

func ShowCreatePost(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		games, err := loadGames(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "create-post.html", gin.H{
				"Error": "Impossible de charger les jeux.",
			})
			return
		}

		c.HTML(http.StatusOK, "create-post.html", gin.H{
			"Games": games,
		})
	}
}

func SubmitCreatePost(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		title := c.PostForm("title")
		content := c.PostForm("content")
		mode := c.PostForm("mode")
		gameIDStr := c.PostForm("game_id")

		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil || gameID <= 0 {
			games, _ := loadGames(db)
			c.HTML(http.StatusBadRequest, "create-post.html", gin.H{
				"Games": games,
				"Error": "Le jeu sélectionné est invalide.",
			})
			return
		}

		if title == "" || content == "" || mode == "" {
			games, _ := loadGames(db)
			c.HTML(http.StatusBadRequest, "create-post.html", gin.H{
				"Games": games,
				"Error": "Tous les champs sont obligatoires.",
			})
			return
		}

		if mode != "Chill" && mode != "Ranked" {
			games, _ := loadGames(db)
			c.HTML(http.StatusBadRequest, "create-post.html", gin.H{
				"Games": games,
				"Error": "Le mode doit être Chill ou Ranked.",
			})
			return
		}

		games, err := loadGames(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "create-post.html", gin.H{
				"Error": "Impossible de recharger les jeux.",
			})
			return
		}

		c.HTML(http.StatusOK, "create-post.html", gin.H{
			"Games":   games,
			"Success": "Post prêt à être enregistré. L'utilisateur n'est pas encore branché.",
		})
	}
}

func loadGames(db *sql.DB) ([]post.Game, error) {
	rows, err := db.Query(`SELECT id, name FROM games ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []post.Game
	for rows.Next() {
		var g post.Game
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, rows.Err()
}
