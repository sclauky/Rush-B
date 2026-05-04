package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"socialauth/backend/forum/auth/utils"
)

type AuthHandler struct {
	DB *sql.DB
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	var exists int
	err := h.DB.QueryRow(`SELECT 1 FROM users WHERE email = ?`, req.Email).Scan(&exists)
	if err == nil {
		http.Error(w, "email already used", http.StatusConflict)
		return
	}
	if err != sql.ErrNoRows {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "password hash error", http.StatusInternalServerError)
		return
	}

	res, err := h.DB.Exec(`
		INSERT INTO users (username, email, password_hash, created_at)
		VALUES (?, ?, ?, ?)
	`, req.Username, req.Email, hash, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		http.Error(w, "insert error", http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	token, err := utils.GenerateToken(id, req.Username)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "account created",
		"user_id": id,
		"token":   token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	var id int64
	var username, hash string

	err := h.DB.QueryRow(`
		SELECT id, username, password_hash
		FROM users
		WHERE email = ?
	`, req.Email).Scan(&id, &username, &hash)

	if err == sql.ErrNoRows {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if !utils.CheckPassword(req.Password, hash) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(id, username)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "login ok",
		"token":   token,
		"user_id": id,
		"username": username,
	})
}