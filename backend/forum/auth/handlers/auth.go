package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"socialauth/backend/forum/auth/models"
)

type RegisterRequest struct {
	Username  string `form:"username"  json:"username"  binding:"required"`
	FirstName string `form:"Firstname" json:"first_name" binding:"required"`
	LastName  string `form:"Lastname"  json:"last_name"  binding:"required"`
	Email     string `form:"email"     json:"email"      binding:"required,email"`
	Password  string `form:"password"  json:"password"   binding:"required,min=6"`
}

func RegisterHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBind(&req); err != nil {
			c.Redirect(http.StatusSeeOther, "/register?error=Champs+invalides")
			return
		}
		var exists int
		db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ? OR username = ?`, req.Email, req.Username).Scan(&exists)
		if exists > 0 {
			c.Redirect(http.StatusSeeOther, "/register?error=Email+ou+pseudo+deja+utilise")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/register?error=Erreur+serveur")
			return
		}
		_, err = db.Exec(`INSERT INTO users (first_name, last_name, email, username, password_hash, role) VALUES (?, ?, ?, ?, ?, 'user')`,
			req.FirstName, req.LastName, req.Email, req.Username, string(hash))
			
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/register?error=Erreur+creation+compte")
			return
		}
		c.Redirect(http.StatusSeeOther, "/login")
	}
}

type LoginRequest struct {
	Email    string `form:"email"    json:"email"    binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required"`
}

func LoginHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBind(&req); err != nil {
			c.Redirect(http.StatusSeeOther, "/login?error=Champs+invalides")
			return
		}
		var user models.User
		err := db.QueryRow(`SELECT id, first_name, last_name, email, username, password_hash, two_fa_enabled FROM users WHERE email = ?`, req.Email).
			Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.PasswordHash, &user.TwoFAEnabled)
		if err == sql.ErrNoRows {
			c.Redirect(http.StatusSeeOther, "/login?error=Email+ou+mot+de+passe+incorrect")
			return
		}
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/login?error=Erreur+serveur")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.Redirect(http.StatusSeeOther, "/login?error=Email+ou+mot+de+passe+incorrect")
			return
		}
		if !user.TwoFAEnabled {
			if err := createSession(c, db, user.ID); err != nil {
				c.Redirect(http.StatusSeeOther, "/login?error=Erreur+session")
			} else {
				c.Redirect(http.StatusSeeOther, "/")
			}
			return
		}
		if err := Send2FACodeForUser(db, user.ID, user.Email, user.FirstName); err != nil {
			c.Redirect(http.StatusSeeOther, "/login?error=Impossible+d%27envoyer+le+code")
			return
		}
		c.SetCookie("pending_2fa_user", strconv.Itoa(user.ID), 600, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/verify-2fa")
	}
}

func LogoutHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, err := c.Cookie("session_token"); err == nil {
			db.Exec(`DELETE FROM sessions WHERE session_token = ?`, token)
		}
		c.SetCookie("session_token", "", -1, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/login")
	}
}

func createSession(c *gin.Context, db *sql.DB, userID int) error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err := db.Exec(`
		INSERT INTO sessions (user_id, session_token, expires_at)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET session_token = excluded.session_token, expires_at = excluded.expires_at
	`, userID, token, expiresAt)
	if err != nil {
		return err
	}
	c.SetCookie("session_token", token, int((7 * 24 * time.Hour).Seconds()), "/", "", false, true)
	return nil
}

func isAuthenticated(c *gin.Context, db *sql.DB) bool {
	token, err := c.Cookie("session_token")
	if err != nil || token == "" {
		return false
	}

	var userID int
	var expiresAt time.Time

	err = db.QueryRow(`
		SELECT user_id, expires_at
		FROM sessions
		WHERE session_token = ?
	`, token).Scan(&userID, &expiresAt)

	if err != nil {
		return false
	}

	if time.Now().After(expiresAt) {
		return false
	}

	return true
}
