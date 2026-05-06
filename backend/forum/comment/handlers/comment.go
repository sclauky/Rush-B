package handlers

import (
	"net/http"
	"strconv"
	"strings"

	commentdb "socialauth/backend/forum/comment/db"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	DB *commentdb.CommentDB
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
// GET /posts/:id/comments — liste les commentaires d'un post
// -------------------------------------------------------

func (h *CommentHandler) ListComments(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id post invalide"})
		return
	}

	viewerID, _ := getUserID(c)

	comments, err := h.DB.GetCommentsByPost(postID, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// -------------------------------------------------------
// POST /posts/:id/comments — créer un commentaire
// -------------------------------------------------------

func (h *CommentHandler) CreateComment(c *gin.Context) {
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

	content := strings.TrimSpace(c.PostForm("content"))
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contenu requis"})
		return
	}

	commentID, err := h.DB.CreateComment(postID, userID, content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur création commentaire"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": commentID})
}

// -------------------------------------------------------
// PUT /posts/:id/comments/:comment_id — éditer un commentaire
// -------------------------------------------------------

func (h *CommentHandler) UpdateComment(c *gin.Context) {
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

	commentID, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id commentaire invalide"})
		return
	}

	// Vérification de sécurité : le commentaire appartient bien à ce post
	ok, err = h.DB.CommentBelongsToPost(commentID, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "commentaire introuvable"})
		return
	}

	content := strings.TrimSpace(c.PostForm("content"))
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contenu requis"})
		return
	}

	err = h.DB.UpdateComment(commentID, userID, content)
	if err != nil {
		if strings.Contains(err.Error(), "not found or forbidden") {
			c.JSON(http.StatusForbidden, gin.H{"error": "commentaire introuvable ou accès refusé"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur mise à jour"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "commentaire mis à jour"})
}

// -------------------------------------------------------
// DELETE /posts/:id/comments/:comment_id — supprimer
// -------------------------------------------------------

func (h *CommentHandler) DeleteComment(c *gin.Context) {
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

	commentID, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id commentaire invalide"})
		return
	}

	ok, err = h.DB.CommentBelongsToPost(commentID, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "commentaire introuvable"})
		return
	}

	err = h.DB.DeleteComment(commentID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found or forbidden") {
			c.JSON(http.StatusForbidden, gin.H{"error": "accès refusé"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur suppression"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "commentaire supprimé"})
}
