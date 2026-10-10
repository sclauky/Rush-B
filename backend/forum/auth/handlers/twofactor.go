package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"socialauth/backend/forum/auth/utils"
)

func generateCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	n := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", n%1000000), nil
}

func Send2FACodeForUser(db *sql.DB, userID int, email, firstName string) error {
	code, err := generateCode()
	if err != nil {
		return err
	}

	db.Exec(`UPDATE two_factor_codes SET used = 1 WHERE user_id = ? AND used = 0`, userID)

	expiresAt := time.Now().Add(10 * time.Minute)
	_, err = db.Exec(
		`INSERT INTO two_factor_codes (user_id, code, expires_at) VALUES (?, ?, ?)`,
		userID, code, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("db insert 2fa: %w", err)
	}

	return utils.Send2FACode(email, firstName, code)
}

type Verify2FARequest struct {
	Code string `json:"code" form:"code" binding:"required"`
}

func Verify2FAHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pendingUserID, err := c.Cookie("pending_2fa_user")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session 2FA expirée"})
			return
		}

		var req Verify2FARequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "code requis"})
			return
		}

		var codeID int
		var storedCode string
		var expiresAt time.Time
		var used int

		err = db.QueryRow(`
			SELECT id, code, expires_at, used
			FROM two_factor_codes
			WHERE user_id = ?
			ORDER BY created_at DESC
			LIMIT 1
		`, pendingUserID).Scan(&codeID, &storedCode, &expiresAt, &used)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "aucun code actif"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur DB"})
			return
		}
		if used == 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "code déjà utilisé"})
			return
		}
		if time.Now().After(expiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "code expiré"})
			return
		}
		if req.Code != storedCode {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "code incorrect"})
			return
		}

		db.Exec(`UPDATE two_factor_codes SET used = 1 WHERE id = ?`, codeID)

		sessionToken, err := utils.GenerateSessionToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "génération session"})
			return
		}

		expiresSession := time.Now().Add(7 * 24 * time.Hour)
		_, err = db.Exec(`
			INSERT INTO sessions (user_id, session_token, expires_at)
			VALUES (?, ?, ?)
			ON CONFLICT(user_id) DO UPDATE SET
				session_token = excluded.session_token,
				expires_at    = excluded.expires_at
		`, pendingUserID, sessionToken, expiresSession)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "création session"})
			return
		}

		c.SetCookie("pending_2fa_user", "", -1, "/", "", false, true)

		c.SetCookie("session_token", sessionToken, int(7*24*time.Hour.Seconds()), "/", "", false, true)

		if c.GetHeader("Accept") == "application/json" {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		} else {
			c.Redirect(http.StatusSeeOther, "/")
		}
	}
}

func Resend2FAHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pendingUserID, err := c.Cookie("pending_2fa_user")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session 2FA expirée"})
			return
		}

		var lastCreated time.Time
		err = db.QueryRow(`
			SELECT created_at FROM two_factor_codes
			WHERE user_id = ? AND used = 0
			ORDER BY created_at DESC LIMIT 1
		`, pendingUserID).Scan(&lastCreated)

		if err == nil && time.Since(lastCreated) < time.Minute {
    remaining := time.Minute - time.Since(lastCreated)
    c.JSON(http.StatusTooManyRequests, gin.H{
        "error":    "trop de demandes",
        "retry_in": int(remaining.Seconds()),
    })
    return
}

		var email, firstName string
		var userID int
		err = db.QueryRow(
			`SELECT id, email, first_name FROM users WHERE id = ?`, pendingUserID,
		).Scan(&userID, &email, &firstName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "utilisateur introuvable"})
			return
		}

		if err := Send2FACodeForUser(db, userID, email, firstName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "envoi impossible"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}