package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"socialauth/backend/forum/auth/utils"
	"strconv"
	"strings"
	"time"

	postDB "socialauth/backend/forum/post/db"
	postHandlers "socialauth/backend/forum/post/handlers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type User struct {
	ID          int      `json:"id"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Username    string   `json:"username"`
	Biography   string   `json:"biography"`
	GameTypes   []string `json:"game_types"`
	GamesPlayed []string `json:"games_played"`
	SkillLevel  string   `json:"skill_level"`
	PlayStyle   string   `json:"play_style"`
	Photos      []string `json:"photos"`
}

type SwipeRequest struct {
	TargetID  int    `json:"target_id"`
	Direction string `json:"direction"`
}

type MatchResponse struct {
	Matched bool `json:"matched"`
	MatchID int  `json:"match_id,omitempty"`
}

type MatchEntry struct {
	User      User      `json:"user"`
	MatchID   int       `json:"match_id"`
	MatchedAt time.Time `json:"matched_at"`
}

type ProfilePost struct {
	ID        int
	Title     string
	Content   string
	ImagePath string
	CreatedAt string
	Likes     int
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			cookie, err := c.Cookie("session_token")
			if err != nil {
				// Requête API → JSON, navigation browser → redirect
				if c.GetHeader("Accept") == "application/json" {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
				} else {
					c.Redirect(http.StatusFound, "/login")
					c.Abort()
				}
				return
			}
			tokenStr = cookie
		}

		var userID int
		var expiresAt time.Time
		err := db.QueryRow(
			`SELECT user_id, expires_at FROM sessions WHERE session_token = ?`, tokenStr,
		).Scan(&userID, &expiresAt)
		if err != nil || time.Now().After(expiresAt) {
			if c.GetHeader("Accept") == "application/json" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
			} else {
				c.Redirect(http.StatusFound, "/login")
				c.Abort()
			}
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func authMiddlewareJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := os.Getenv("JWT_SECRET")
		tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if tokenStr == "" {
			cookie, err := c.Cookie("session_token")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
				return
			}
			tokenStr = cookie
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}
		userID := int(claims["user_id"].(float64))
		c.Set("userID", userID)
		c.Next()
	}
}

func currentUserID(c *gin.Context) int {
	v, exists := c.Get("userID")
	if !exists {
		return 0
	}

	userID, ok := v.(int)
	if !ok {
		return 0
	}

	return userID
}

func validateSteamOpenID(values url.Values) error {
	values.Set("openid.mode", "check_authentication")

	resp, err := http.PostForm("https://steamcommunity.com/openid/login", values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(string(body), "is_valid:true") {
		return fmt.Errorf("openid not valid")
	}

	return nil
}

func extractSteamID(claimedID string) (string, error) {
	re := regexp.MustCompile(`^https?://steamcommunity\.com/openid/id/(\d+)$`)
	matches := re.FindStringSubmatch(claimedID)
	if len(matches) != 2 {
		return "", fmt.Errorf("invalid claimed id")
	}
	return matches[1], nil
}

func steamLinkHandler(c *gin.Context) {
	userID := currentUserID(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	returnTo := baseURL + "/auth/steam/callback"
	realm := baseURL

	params := url.Values{}
	params.Set("openid.ns", "http://specs.openid.net/auth/2.0")
	params.Set("openid.mode", "checkid_setup")
	params.Set("openid.return_to", returnTo)
	params.Set("openid.realm", realm)
	params.Set("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
	params.Set("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")

	redirectURL := "https://steamcommunity.com/openid/login?" + params.Encode()
	c.Redirect(http.StatusFound, redirectURL)
}

func steamCallbackHandler(c *gin.Context) {
	userID := currentUserID(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	values := url.Values{}
	for key, vals := range c.Request.URL.Query() {
		if strings.HasPrefix(key, "openid.") && len(vals) > 0 {
			values.Set(key, vals[0])
		}
	}

	if err := validateSteamOpenID(values); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid steam response"})
		return
	}

	claimedID := values.Get("openid.claimed_id")
	steamID, err := extractSteamID(claimedID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid steam id"})
		return
	}

	var existingUserID int
	err = db.QueryRow(`SELECT id FROM users WHERE steam_id = ?`, steamID).Scan(&existingUserID)
	if err == nil && existingUserID != 0 && existingUserID != userID {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "steam account already linked"})
		return
	}

	_, err = db.Exec(`UPDATE users SET steam_id = ? WHERE id = ?`, steamID, userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	c.Redirect(http.StatusFound, "/profile")
}

func getProfiles(c *gin.Context) {
	userID := currentUserID(c)

	rows, err := db.Query(`
        SELECT u.id, u.first_name, u.last_name, u.username, COALESCE(u.biography, ''),
               COALESCE(up.skill_level, ''), COALESCE(up.play_style, '')
        FROM users u
        LEFT JOIN user_preferences up ON up.user_id = u.id
        WHERE u.id != ?
          AND u.id NOT IN (SELECT target_id FROM swipes WHERE swiper_id = ?)
        ORDER BY RANDOM()
        LIMIT 20
    `, userID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()

	profiles := []User{}
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Username, &u.Biography, &u.SkillLevel, &u.PlayStyle)
		u.GameTypes = queryStrings(db, `SELECT game_type FROM user_game_types WHERE user_id = ?`, u.ID)
		u.GamesPlayed = queryStrings(db, `SELECT game_name FROM user_games WHERE user_id = ?`, u.ID)
		u.Photos = queryStrings(db, `SELECT photo_path FROM user_photos WHERE user_id = ? ORDER BY position ASC LIMIT 6`, u.ID)
		profiles = append(profiles, u)
	}
	c.JSON(http.StatusOK, profiles)
}

func handleSwipe(c *gin.Context) {
	userID := currentUserID(c)

	var req SwipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	_, err := db.Exec(
		`INSERT OR IGNORE INTO swipes (swiper_id, target_id, direction, created_at) VALUES (?, ?, ?, ?)`,
		userID, req.TargetID, req.Direction, time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	resp := MatchResponse{}
	if req.Direction == "like" {
		var reverseID int
		err := db.QueryRow(
			`SELECT id FROM swipes WHERE swiper_id = ? AND target_id = ? AND direction = 'like'`,
			req.TargetID, userID,
		).Scan(&reverseID)
		if err == nil {
			var matchID int
			err2 := db.QueryRow(
				`SELECT id FROM matches WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)`,
				userID, req.TargetID, req.TargetID, userID,
			).Scan(&matchID)
			if err2 == sql.ErrNoRows {
				res, _ := db.Exec(
					`INSERT INTO matches (user1_id, user2_id, created_at) VALUES (?, ?, ?)`,
					userID, req.TargetID, time.Now(),
				)
				id64, _ := res.LastInsertId()
				matchID = int(id64)
			}
			resp.Matched = true
			resp.MatchID = matchID
		}
	}
	c.JSON(http.StatusOK, resp)
}

func getMatches(c *gin.Context) {
	userID := currentUserID(c)

	rows, err := db.Query(`
        SELECT u.id, u.first_name, u.last_name, u.username, COALESCE(u.biography, ''),
               COALESCE(up.skill_level, ''), COALESCE(up.play_style, ''), m.id, m.created_at
        FROM matches m
        JOIN users u ON (u.id = CASE WHEN m.user1_id = ? THEN m.user2_id ELSE m.user1_id END)
        LEFT JOIN user_preferences up ON up.user_id = u.id
        WHERE m.user1_id = ? OR m.user2_id = ?
        ORDER BY m.created_at DESC
    `, userID, userID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()

	matches := []MatchEntry{}
	for rows.Next() {
		var u User
		var me MatchEntry
		rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Username, &u.Biography,
			&u.SkillLevel, &u.PlayStyle, &me.MatchID, &me.MatchedAt)
		u.GameTypes = queryStrings(db, `SELECT game_type FROM user_game_types WHERE user_id = ?`, u.ID)
		u.Photos = queryStrings(db, `SELECT photo_path FROM user_photos WHERE user_id = ? ORDER BY position ASC LIMIT 1`, u.ID)
		me.User = u
		matches = append(matches, me)
	}
	c.JSON(http.StatusOK, matches)
}

func deleteMatch(c *gin.Context) {
	userID := currentUserID(c)
	matchID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var u1, u2 int
	err = db.QueryRow(`SELECT user1_id, user2_id FROM matches WHERE id = ?`, matchID).Scan(&u1, &u2)
	if err != nil || (u1 != userID && u2 != userID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	db.Exec(`DELETE FROM matches WHERE id = ?`, matchID)
	c.JSON(http.StatusOK, gin.H{"ok": "unmatched"})
}

func uploadPhoto(c *gin.Context) {
	userID := currentUserID(c)

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM user_photos WHERE user_id = ?`, userID).Scan(&count)
	if count >= 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max 6 photos"})
		return
	}

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file"})
		return
	}
	defer file.Close()

	uploadDir := "frontend/web/uploads/swipe"
	os.MkdirAll(uploadDir, 0755)

	filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().Unix(), header.Filename)
	dst, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot save file"})
		return
	}
	defer dst.Close()

	buf := make([]byte, 10<<20)
	n, _ := file.Read(buf)
	dst.Write(buf[:n])

	photoPath := "/uploads/swipe/" + filename
	db.Exec(`INSERT INTO user_photos (user_id, photo_path, position) VALUES (?, ?, ?)`,
		userID, photoPath, count)

	c.JSON(http.StatusOK, gin.H{"path": photoPath})
}

func deletePhoto(c *gin.Context) {
	userID := currentUserID(c)
	photoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	db.Exec(`DELETE FROM user_photos WHERE id = ? AND user_id = ?`, photoID, userID)
	c.JSON(http.StatusOK, gin.H{"ok": "deleted"})
}

func queryStrings(db *sql.DB, query string, args ...interface{}) []string {
	rows, err := db.Query(query, args...)
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		out = append(out, s)
	}
	if out == nil {
		return []string{}
	}
	return out
}

func runMigrationsFromFile(db *sql.DB, filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(content))
	return err
}

func registerHandler(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	firstName := strings.TrimSpace(c.PostForm("first_name"))
	lastName := strings.TrimSpace(c.PostForm("last_name"))
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	if username == "" || firstName == "" || lastName == "" || email == "" || password == "" || confirmPassword == "" {
		c.HTML(http.StatusBadRequest, "home.html", gin.H{
			"RegisterError": "Tous les champs sont obligatoires",
		})
		return
	}

	if password != confirmPassword {
		c.HTML(http.StatusBadRequest, "home.html", gin.H{
			"ConfirmPasswordError": "Les mots de passe ne correspondent pas",
		})
		return
	}

	if !utils.IsValidPassword(password) {
		c.HTML(http.StatusBadRequest, "home.html", gin.H{
			"PasswordError": "Le mot de passe doit contenir au moins 7 caractères, une majuscule, une minuscule et un chiffre.",
		})
		return
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
		return
	}

	_, err = db.Exec(`
        INSERT INTO users (first_name, last_name, email, username, password_hash)
        VALUES (?, ?, ?, ?, ?)
    `, firstName, lastName, email, username, hashedPassword)
	if err != nil {
		c.HTML(http.StatusBadRequest, "home.html", gin.H{
			"RegisterError": "Impossible de créer le compte. Email ou pseudo peut-être déjà utilisé.",
		})
		return
	}

	c.Redirect(http.StatusFound, "/login")
}

func loginHandler(c *gin.Context) {
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")

	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Error": "Email et mot de passe obligatoires",
		})
		return
	}

	var userID int
	var passwordHash string

	err := db.QueryRow(`SELECT id, password_hash FROM users WHERE email = ?`, email).Scan(&userID, &passwordHash)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Error": "Identifiants invalides",
		})
		return
	}

	if !utils.CheckPassword(password, passwordHash) {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Error": "Identifiants invalides",
		})
		return
	}

	sessionToken := fmt.Sprintf("%d_%d", userID, time.Now().UnixNano())
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = db.Exec(
		`INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)`,
		userID, sessionToken, expiresAt,
	)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
		return
	}

	c.SetCookie("session_token", sessionToken, 86400, "/", "", false, true)
	c.Redirect(http.StatusFound, "/profile")
}

func logoutHandler(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err == nil && token != "" {
		db.Exec(`DELETE FROM sessions WHERE session_token = ?`, token)
	}

	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func followUserHandler(c *gin.Context) {
	currentID := currentUserID(c)

	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "404.html", gin.H{})
		return
	}

	if targetID == currentID {
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	_, err = db.Exec(`
        INSERT OR IGNORE INTO follows (follower_id, followed_id, created_at)
        VALUES (?, ?, ?)
    `, currentID, targetID, time.Now())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
		return
	}

	c.Redirect(http.StatusFound, "/users/"+strconv.Itoa(targetID))
}

func unfollowUserHandler(c *gin.Context) {
	currentID := currentUserID(c)

	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "404.html", gin.H{})
		return
	}

	_, err = db.Exec(`
        DELETE FROM follows
        WHERE follower_id = ? AND followed_id = ?
    `, currentID, targetID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
		return
	}

	c.Redirect(http.StatusFound, "/users/"+strconv.Itoa(targetID))
}

func getOtherUserProfile(c *gin.Context) {
	currentID := currentUserID(c)

	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "404.html", gin.H{})
		return
	}

	if targetID == currentID {
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	var u User
	err = db.QueryRow(`
        SELECT u.id, u.first_name, u.last_name, u.username, COALESCE(u.biography, ''),
               COALESCE(up.skill_level, ''), COALESCE(up.play_style, '')
        FROM users u
        LEFT JOIN user_preferences up ON up.user_id = u.id
        WHERE u.id = ?
    `, targetID).Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Username,
		&u.Biography,
		&u.SkillLevel,
		&u.PlayStyle,
	)

	if err == sql.ErrNoRows {
		c.HTML(http.StatusNotFound, "404.html", gin.H{})
		return
	}
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
		return
	}

	u.GameTypes = queryStrings(db, `SELECT game_type FROM user_game_types WHERE user_id = ?`, u.ID)
	u.GamesPlayed = queryStrings(db, `SELECT game_name FROM user_games WHERE user_id = ?`, u.ID)
	u.Photos = queryStrings(db, `SELECT photo_path FROM user_photos WHERE user_id = ? ORDER BY position ASC LIMIT 6`, u.ID)

	var followCount int
	err = db.QueryRow(`
        SELECT COUNT(*)
        FROM follows
        WHERE follower_id = ? AND followed_id = ?
    `, currentID, targetID).Scan(&followCount)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
		return
	}

	isFollowing := followCount > 0

	c.HTML(http.StatusOK, "user-profile.html", gin.H{
		"User":        u,
		"IsFollowing": isFollowing,
	})
}

func main() {
	_ = godotenv.Load()

	var err error
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "backend/database/data.db"
	}

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Erreur ouverture DB:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Erreur ping DB:", err)
	}

	if err := runMigrationsFromFile(db, "backend/database/schema.sql"); err != nil {
		log.Fatal("Erreur migration DB:", err)
	}

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.LoadHTMLGlob("frontend/web/templates/*.html")
	r.Static("/static", "frontend/web/static")
	r.Static("/uploads", "frontend/web/uploads")

	r.GET("/swipe", func(c *gin.Context) {
		c.HTML(http.StatusOK, "swipe.html", gin.H{})
	})

	r.GET("/", func(c *gin.Context) {
		isLoggedIn := false

		token, err := c.Cookie("session_token")
		if err == nil {
			var userID int
			var expiresAt time.Time

			err = db.QueryRow(
				`SELECT user_id, expires_at FROM sessions WHERE session_token = ?`,
				token,
			).Scan(&userID, &expiresAt)

			if err == nil && time.Now().Before(expiresAt) {
				isLoggedIn = true
			}
		}

		c.HTML(http.StatusOK, "home.html", gin.H{
			"IsLoggedIn": isLoggedIn,
		})
	})

	r.GET("/logout", logoutHandler)

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})
	r.POST("/login", loginHandler)
	r.POST("/register", registerHandler)

	r.GET("/posts", authMiddleware(), func(c *gin.Context) {
		pdb := &postDB.PostDB{DB: db}
		posts, err := pdb.GetPosts(postDB.PostFilter{}, 0)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
			return
		}
		c.HTML(http.StatusOK, "posts.html", gin.H{
			"Posts": posts,
		})
	})

	r.GET("/create-post", authMiddleware(), postHandlers.ShowCreatePost(db))
	r.POST("/create-post", authMiddleware(), postHandlers.SubmitCreatePost(db))

	r.GET("/edit-profile", authMiddleware(), func(c *gin.Context) {
		userID := currentUserID(c)

		var username, email, biography string
		err := db.QueryRow(`
			SELECT username, email, COALESCE(biography, '')
			FROM users
			WHERE id = ?
		`, userID).Scan(&username, &email, &biography)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
			return
		}

		var avatarPath string
		err = db.QueryRow(`
			SELECT photo_path
			FROM user_photos
			WHERE user_id = ?
			ORDER BY position ASC, id ASC
			LIMIT 1
		`, userID).Scan(&avatarPath)
		if err != nil || avatarPath == "" {
			avatarPath = "/static/images/default-avatar.png"
		}

		c.HTML(http.StatusOK, "edit-profile.html", gin.H{
			"Username":   username,
			"Email":      email,
			"Biography":  biography,
			"AvatarPath": avatarPath,
		})
	})

	r.POST("/edit-profile", authMiddleware(), func(c *gin.Context) {
        userID := currentUserID(c)

        username := strings.TrimSpace(c.PostForm("username"))
        email := strings.TrimSpace(c.PostForm("email"))
        biography := strings.TrimSpace(c.PostForm("bio"))

        var avatarPath string
        err := db.QueryRow(`
            SELECT photo_path
            FROM user_photos
            WHERE user_id = ?
            ORDER BY position ASC, id ASC
            LIMIT 1
        `, userID).Scan(&avatarPath)
        if err != nil || avatarPath == "" {
            avatarPath = "/static/images/default-avatar.png"
        }

        if username == "" || email == "" {
            c.HTML(http.StatusBadRequest, "edit-profile.html", gin.H{
                "Username":     username,
                "Email":        email,
                "Biography":    biography,
                "AvatarPath":   avatarPath,
                "ErrorMessage": "Le pseudo et l'email sont obligatoires.",
            })
            return
        }

        if len([]rune(biography)) > 201 { // Pour que ça marche avec 200 caractères tout pile.
            c.HTML(http.StatusBadRequest, "edit-profile.html", gin.H{
                "Username":     username,
                "Email":        email,
                "Biography":    biography,
                "AvatarPath":   avatarPath,
                "ErrorMessage": "La biographie ne doit pas dépasser 200 caractères.",
            })
            return
        }

        _, err = db.Exec(`
            UPDATE users
            SET username = ?, email = ?, biography = ?
            WHERE id = ?
        `, username, email, biography, userID)
        if err != nil {
            c.HTML(http.StatusBadRequest, "edit-profile.html", gin.H{
                "Username":     username,
                "Email":        email,
                "Biography":    biography,
                "AvatarPath":   avatarPath,
                "ErrorMessage": "Impossible de mettre à jour le profil. Email ou pseudo peut-être déjà utilisé.",
            })
            return
        }

        file, err := c.FormFile("avatar")
        if err == nil && file != nil {
            uploadDir := "frontend/web/uploads/profile"
            os.MkdirAll(uploadDir, 0755)

            filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().Unix(), filepath.Base(file.Filename))
            dst := filepath.Join(uploadDir, filename)

            if err := c.SaveUploadedFile(file, dst); err == nil {
                photoPath := "/uploads/profile/" + filename

                var existingPhotoID int
                err = db.QueryRow(`
                    SELECT id
                    FROM user_photos
                    WHERE user_id = ?
                    ORDER BY position ASC, id ASC
                    LIMIT 1
                `, userID).Scan(&existingPhotoID)

                if err == nil {
                    db.Exec(`
                        UPDATE user_photos
                        SET photo_path = ?
                        WHERE id = ?
                    `, photoPath, existingPhotoID)
                } else {
                    db.Exec(`
                        INSERT INTO user_photos (user_id, photo_path, position)
                        VALUES (?, ?, 0)
                    `, userID, photoPath)
                }
            }
        }

        c.Redirect(http.StatusFound, "/profile")
    })

	r.GET("/profile", authMiddleware(), func(c *gin.Context) {
		userID := currentUserID(c)

		var username, biography, createdAt string
		err := db.QueryRow(`
            SELECT username, COALESCE(biography, ''), created_at
            FROM users
            WHERE id = ?
        `, userID).Scan(&username, &biography, &createdAt)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
			return
		}

		var postCount int
		err = db.QueryRow(`
            SELECT COUNT(*)
            FROM posts
            WHERE user_id = ?
        `, userID).Scan(&postCount)
		if err != nil {
			postCount = 0
		}

		var followersCount int
		err = db.QueryRow(`
            SELECT COUNT(*)
            FROM matches
            WHERE user1_id = ? OR user2_id = ?
        `, userID, userID).Scan(&followersCount)
		if err != nil {
			followersCount = 0
		}

		var avatarPath string
		err = db.QueryRow(`
            SELECT photo_path
            FROM user_photos
            WHERE user_id = ?
            ORDER BY position ASC, id ASC
            LIMIT 1
        `, userID).Scan(&avatarPath)
		if err != nil || avatarPath == "" {
			avatarPath = "/static/images/default-avatar.png"
		}

		rows, err := db.Query(`
            SELECT
                p.id,
                p.title,
                p.content,
                COALESCE(p.image_path, ''),
                strftime('%d/%m/%Y', p.created_at) as created_at,
                COALESCE(SUM(CASE WHEN pr.reaction_type = 'like' THEN 1 ELSE 0 END), 0) as likes
            FROM posts p
            LEFT JOIN post_reactions pr ON pr.post_id = p.id
            WHERE p.user_id = ?
            GROUP BY p.id
            ORDER BY p.created_at DESC
        `, userID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{})
			return
		}
		defer rows.Close()

		var posts []ProfilePost
		for rows.Next() {
			var p ProfilePost
			if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Likes); err == nil {
				posts = append(posts, p)
			}
		}

		joinedDate := createdAt
		layouts := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04:05Z",
		}

		for _, layout := range layouts {
			if t, err := time.Parse(layout, createdAt); err == nil {
				joinedDate = t.Format("02/01/2006")
				break
			}
		}

		c.HTML(http.StatusOK, "profile.html", gin.H{
			"Username":       username,
			"Biography":      biography,
			"PostCount":      postCount,
			"FollowersCount": followersCount,
			"JoinedDate":     joinedDate,
			"AvatarPath":     avatarPath,
			"Posts":          posts,
		})
	})

	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{})
	})

	steamGroup := r.Group("/", authMiddleware())
	{
		steamGroup.GET("/account/steam/link", steamLinkHandler)
		steamGroup.GET("/auth/steam/callback", steamCallbackHandler)
	}

	api := r.Group("/api/swipe", authMiddleware())
	{
		api.GET("/profiles", getProfiles)
		api.POST("/action", handleSwipe)
		api.GET("/matches", getMatches)
		api.DELETE("/matches/:id", deleteMatch)
		api.POST("/photos", uploadPhoto)
		api.DELETE("/photos/:id", deletePhoto)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Serveur lancé sur http://localhost:%s", port)
	log.Fatal(r.Run(":" + port))
}
