package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			cookie, err := c.Cookie("session_token")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
				return
			}
			tokenStr = cookie
		}

		var userID int
		var expiresAt time.Time
		err := db.QueryRow(
			`SELECT user_id, expires_at FROM sessions WHERE session_token = ?`, tokenStr,
		).Scan(&userID, &expiresAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}
		if time.Now().After(expiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
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
	id, _ := c.Get("userID")
	return id.(int)
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

    // Templates + statiques
    r.LoadHTMLGlob("frontend/web/templates/*.html")
    r.Static("/static", "frontend/web/static")
    r.Static("/uploads", "frontend/web/uploads")

    r.StaticFile("/swipe", "/frontend/web/templates/swipe.html")
    r.StaticFile("/edit-profile.html", "/frontend/web/profile/edit-profile.html")

    // Pages HTML
    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "home.html", gin.H{})
    })

    r.GET("/login", func(c *gin.Context) {
        c.HTML(http.StatusOK, "login.html", gin.H{})
    })

    r.GET("/posts", func(c *gin.Context) {
        c.HTML(http.StatusOK, "posts.html", gin.H{})
    })

	r.GET("/frontend/web/swipe/swipe.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "swipe.html", gin.H{})
	})

    r.GET("/create-post", func(c *gin.Context) {
        c.HTML(http.StatusOK, "create-post.html", gin.H{})
    })

    r.GET("/profile", func(c *gin.Context) {
        c.HTML(http.StatusOK, "profile.html", gin.H{})
    })

    // API swipe
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