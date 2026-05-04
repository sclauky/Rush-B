package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Bio       string    `json:"bio"`
	Avatar    string    `json:"avatar"`
	PostCount int       `json:"post_count"`
	Followers int       `json:"followers"`
	CreatedAt time.Time `json:"created_at"`
}

type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url"`
	Likes     int       `json:"likes"`
	CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	initDB()

	// Créer le dossier uploads s'il n'existe pas
	os.MkdirAll("./uploads", 0755)

	// Routes
	http.HandleFunc("/api/user/", getUserProfile)
	http.HandleFunc("/api/user/update", updateUserProfile)
	http.HandleFunc("/api/user/posts/", getUserPosts)
	http.HandleFunc("/api/upload", uploadImage)

	// Servir les fichiers statiques
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			bio TEXT,
			avatar TEXT,
			post_count INTEGER DEFAULT 0,
			followers INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			content TEXT NOT NULL,
			image_url TEXT,
			likes INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Insérer un utilisateur de test s'il n'existe pas
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		_, err := db.Exec(`INSERT INTO users (username, email, bio, avatar, post_count, followers) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			"johndoe", "john@example.com", "Passionné de technologie et de design moderne 🚀", 
			"/uploads/default-avatar.jpg", 12, 348)
		if err != nil {
			log.Println("Error inserting test user:", err)
		}

		// Insérer quelques posts de test
		posts := []string{
			"Découverte incroyable aujourd'hui ! Le nouveau framework que j'utilise est vraiment puissant 💪",
			"Qui d'autre est fan de l'architecture moderne ? Ce design est absolument magnifique !",
			"Journée productive : 3 features complétées et 2 bugs corrigés ✅",
		}
		for _, content := range posts {
			db.Exec("INSERT INTO posts (user_id, content, likes) VALUES (?, ?, ?)", 
				1, content, 0)
		}
		db.Exec("UPDATE users SET post_count = ? WHERE id = ?", len(posts), 1)
	}
}

func getUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.URL.Path[len("/api/user/"):]
	if userID == "" {
		userID = "1" // Par défaut
	}

	var user User
	err := db.QueryRow(`SELECT id, username, email, bio, COALESCE(avatar, ''), post_count, followers, created_at 
		FROM users WHERE id = ?`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Bio, &user.Avatar, 
		&user.PostCount, &user.Followers, &user.CreatedAt)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func getUserPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.URL.Path[len("/api/user/posts/"):]
	if userID == "" {
		userID = "1"
	}

	rows, err := db.Query(`SELECT id, user_id, content, COALESCE(image_url, ''), likes, created_at 
		FROM posts WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.ImageURL, &post.Likes, &post.CreatedAt)
		if err != nil {
			continue
		}
		posts = append(posts, post)
	}

	if posts == nil {
		posts = []Post{}
	}

	json.NewEncoder(w).Encode(posts)
}

func updateUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	username := r.FormValue("username")
	email := r.FormValue("email")
	bio := r.FormValue("bio")

	avatarPath := ""
	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		// Générer un nom de fichier unique
		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("avatar_%s_%d%s", userID, time.Now().Unix(), ext)
		avatarPath = "/uploads/" + filename

		dst, err := os.Create("./uploads/" + filename)
		if err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
	}

	// Mettre à jour la base de données
	query := "UPDATE users SET username = ?, email = ?, bio = ?"
	args := []interface{}{username, email, bio}

	if avatarPath != "" {
		query += ", avatar = ?"
		args = append(args, avatarPath)
	}

	query += " WHERE id = ?"
	args = append(args, userID)

	_, err = db.Exec(query, args...)
	if err != nil {
		http.Error(w, "Unable to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Profile updated successfully",
	})
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Unable to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("upload_%d%s", time.Now().Unix(), ext)
	filepath := "/uploads/" + filename

	dst, err := os.Create("./uploads/" + filename)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": filepath,
	})
}
