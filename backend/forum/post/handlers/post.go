package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	postDB "socialauth/backend/forum/post/db"

	"github.com/gin-gonic/gin"
)

// récupère l'userID injecté par authMiddleware
func currentUserID(c *gin.Context) (int, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := val.(int)
	return id, ok
}

func ShowCreatePost(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pdb := &postDB.PostDB{DB: db}
		games, err := pdb.GetAllGames()
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
		pdb := &postDB.PostDB{DB: db}

		// 1. Récupérer l'utilisateur connecté (injecté par authMiddleware)
		userID, ok := currentUserID(c)
		if !ok {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		// 2. Lire les champs du formulaire
		title := c.PostForm("title")
		content := c.PostForm("content")
		mode := c.PostForm("mode")
		gameIDStr := c.PostForm("game_id")

		// helper pour renvoyer le formulaire avec une erreur
		renderError := func(msg string) {
			games, _ := pdb.GetAllGames()
			c.HTML(http.StatusBadRequest, "create-post.html", gin.H{
				"Games": games,
				"Error": msg,
			})
		}

		// 3. Validation
		if title == "" || content == "" || mode == "" || gameIDStr == "" {
			renderError("Tous les champs sont obligatoires.")
			return
		}
		if mode != "Chill" && mode != "Ranked" {
			renderError("Le mode doit être Chill ou Ranked.")
			return
		}
		gameID, err := strconv.Atoi(gameIDStr)
		if err != nil || gameID <= 0 {
			renderError("Le jeu sélectionné est invalide.")
			return
		}

		// 4. Insertion en base
		_, err = pdb.CreatePost(postDB.CreatePostInput{
			UserID:  userID,
			GameID:  gameID,
			Mode:    mode,
			Title:   title,
			Content: content,
		})
		if err != nil {
			renderError("Erreur lors de la création du post. Réessaie.")
			return
		}

		// 5. Redirection vers le feed
		c.Redirect(http.StatusFound, "/posts")
	}
}
