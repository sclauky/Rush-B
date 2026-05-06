package handlers

import (
	"net/http"
	"strconv"

	likedb "socialauth/backend/forum/like/db"

	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
	DB *likedb.LikeDB
}

func getUserID(c *gin.Context) (int, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	}
	return 0, false
}

// -------------------------------------------------------
// POST /posts/:id/like
// POST /posts/:id/dislike
// -------------------------------------------------------

func (h *LikeHandler) ReactToPost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non connecté"})
		return
	}

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id post invalide"})
		return
	}

	// On déduit le type de réaction depuis l'URL
	// /posts/:id/like → "like" | /posts/:id/dislike → "dislike"
	reactionType := c.Param("reaction") // "like" ou "dislike"
	if reactionType != "like" && reactionType != "dislike" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "réaction invalide"})
		return
	}

	err = h.DB.TogglePostReaction(postID, userID, reactionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}

	// On retourne les nouveaux compteurs pour que le frontend
	// puisse mettre à jour l'affichage sans recharger la page
	counts, err := h.DB.GetPostReactions(postID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur comptage"})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// -------------------------------------------------------
// POST /posts/:id/comments/:comment_id/like
// POST /posts/:id/comments/:comment_id/dislike
// -------------------------------------------------------

func (h *LikeHandler) ReactToComment(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non connecté"})
		return
	}

	commentID, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id commentaire invalide"})
		return
	}

	reactionType := c.Param("reaction")
	if reactionType != "like" && reactionType != "dislike" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "réaction invalide"})
		return
	}

	err = h.DB.ToggleCommentReaction(commentID, userID, reactionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}

	counts, err := h.DB.GetCommentReactions(commentID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur comptage"})
		return
	}

	c.JSON(http.StatusOK, counts)
}
