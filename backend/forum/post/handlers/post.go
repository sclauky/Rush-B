package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	postdb "socialauth/backend/forum/post/db"
	"socialauth/backend/forum/post/models"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	DB *postdb.PostDB
}

// -------------------------------------------------------
// UTILITAIRE — récupérer l'user ID depuis le contexte Gin
// -------------------------------------------------------

// À adapter selon ce que ton collègue stocke dans le JWT.
// Demande-lui : c'est "sub", "user_id" ou "id" ?
func getUserID(c *gin.Context) (int, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return int(v), true // JWT décode les nombres en float64
	case int:
		return v, true
	}
	return 0, false
}

// -------------------------------------------------------
// GET /posts — liste avec filtres optionnels
// -------------------------------------------------------
// Query params possibles :
//   category_id=1
//   game_id=2
//   filter=mine        (mes posts — connecté requis)
//   filter=liked       (mes likes — connecté requis)

func (h *PostHandler) ListPosts(c *gin.Context) {
	viewerID, _ := getUserID(c) // 0 si non connecté, c'est géré

	filter := postdb.PostFilter{}

	if catStr := c.Query("category_id"); catStr != "" {
		id, err := strconv.Atoi(catStr)
		if err == nil {
			filter.CategoryID = &id
		}
	}
	if gameStr := c.Query("game_id"); gameStr != "" {
		id, err := strconv.Atoi(gameStr)
		if err == nil {
			filter.GameID = &id
		}
	}

	switch c.Query("filter") {
	case "mine":
		if viewerID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "connecté requis"})
			return
		}
		filter.UserID = &viewerID
	case "liked":
		if viewerID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "connecté requis"})
			return
		}
		filter.LikedBy = &viewerID
	}

	posts, err := h.DB.GetPosts(filter, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// -------------------------------------------------------
// GET /posts/:id — détail d'un post
// -------------------------------------------------------

func (h *PostHandler) GetPost(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalide"})
		return
	}

	viewerID, _ := getUserID(c)

	post, err := h.DB.GetPostByID(postID, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur serveur"})
		return
	}
	if post == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post introuvable"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// -------------------------------------------------------
// POST /posts — créer un post (connecté requis)
// -------------------------------------------------------
// Attend un multipart/form-data avec :
//   title, content, game_id, mode, category_ids (multiple), image (optionnel)

func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non connecté"})
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	content := strings.TrimSpace(c.PostForm("content"))
	mode := c.PostForm("mode")
	gameIDStr := c.PostForm("game_id")

	if title == "" || content == "" || gameIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title, content et game_id requis"})
		return
	}
	if mode != "Chill" && mode != "Ranked" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode doit être Chill ou Ranked"})
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id invalide"})
		return
	}

	// Catégories (plusieurs IDs possibles)
	catStrs := c.PostFormArray("category_ids")
	var categoryIDs []int
	for _, s := range catStrs {
		id, err := strconv.Atoi(s)
		if err == nil {
			categoryIDs = append(categoryIDs, id)
		}
	}

	// Upload image (optionnel)
	var imagePath *string
	file, header, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()
		path, err := saveImage(file, header.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur upload image"})
			return
		}
		imagePath = &path
	}

	input := postdb.CreatePostInput{
		UserID:      userID,
		GameID:      gameID,
		Mode:        mode,
		Title:       title,
		Content:     content,
		ImagePath:   imagePath,
		CategoryIDs: categoryIDs,
	}

	postID, err := h.DB.CreatePost(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur création post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": postID})
}

// -------------------------------------------------------
// PUT /posts/:id — éditer un post (proprio uniquement)
// -------------------------------------------------------

func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non connecté"})
		return
	}

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalide"})
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	content := strings.TrimSpace(c.PostForm("content"))
	mode := c.PostForm("mode")

	if title == "" || content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title et content requis"})
		return
	}
	if mode != "Chill" && mode != "Ranked" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode doit être Chill ou Ranked"})
		return
	}

	catStrs := c.PostFormArray("category_ids")
	var categoryIDs []int
	for _, s := range catStrs {
		id, err := strconv.Atoi(s)
		if err == nil {
			categoryIDs = append(categoryIDs, id)
		}
	}

	var imagePath *string
	file, header, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()
		path, err := saveImage(file, header.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur upload image"})
			return
		}
		imagePath = &path
	}

	input := postdb.UpdatePostInput{
		PostID:      postID,
		UserID:      userID,
		Mode:        mode,
		Title:       title,
		Content:     content,
		ImagePath:   imagePath,
		CategoryIDs: categoryIDs,
	}

	err = h.DB.UpdatePost(input)
	if err != nil {
		if strings.Contains(err.Error(), "not found or forbidden") {
			c.JSON(http.StatusForbidden, gin.H{"error": "post introuvable ou accès refusé"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur mise à jour"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post mis à jour"})
}

// -------------------------------------------------------
// DELETE /posts/:id — supprimer un post (proprio uniquement)
// -------------------------------------------------------

func (h *PostHandler) DeletePost(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "non connecté"})
		return
	}

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalide"})
		return
	}

	err = h.DB.DeletePost(postID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found or forbidden") {
			c.JSON(http.StatusForbidden, gin.H{"error": "post introuvable ou accès refusé"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur suppression"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post supprimé"})
}

// -------------------------------------------------------
// GET /posts/form-data — jeux + catégories pour le formulaire
// -------------------------------------------------------

func (h *PostHandler) GetFormData(c *gin.Context) {
	games, err := h.DB.GetAllGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur chargement jeux"})
		return
	}
	cats, err := h.DB.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur chargement catégories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"games":      games,
		"categories": cats,
	})
}

// -------------------------------------------------------
// UTILITAIRE — sauvegarde d'image sur disque
// -------------------------------------------------------

func saveImage(src io.Reader, originalName string) (string, error) {
	ext := filepath.Ext(originalName)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[strings.ToLower(ext)] {
		return "", fmt.Errorf("extension non autorisée")
	}

	// Nom unique basé sur le timestamp
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dir := "./frontend/web/static/uploads"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// On retourne le chemin relatif utilisable dans les templates HTML
	return "/static/uploads/" + filename, nil
}

// -------------------------------------------------------
// GET /games — liste tous les jeux (utile pour le frontend)
// -------------------------------------------------------

func (h *PostHandler) ListGames(c *gin.Context) {
	games, err := h.DB.GetAllGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur"})
		return
	}
	c.JSON(http.StatusOK, games)
}

// -------------------------------------------------------
// GET /categories — liste toutes les catégories
// -------------------------------------------------------

func (h *PostHandler) ListCategories(c *gin.Context) {
	cats, err := h.DB.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur"})
		return
	}

	// On retourne aussi les modes disponibles pour le formulaire
	c.JSON(http.StatusOK, gin.H{
		"categories": cats,
		"modes":      []string{"Chill", "Ranked"},
	})
}

// ServeCreateForm sert la page HTML de création de post
func (h *PostHandler) ServeCreateForm(c *gin.Context) {
	games, err := h.DB.GetAllGames()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	cats, err := h.DB.GetAllCategories()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusOK, "create-post.html", gin.H{
		"Games":      games,
		"Categories": cats,
		"Modes":      []models.Mode{"Chill", "Ranked"},
	})
}
