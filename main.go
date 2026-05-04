package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	tmpl *template.Template
	db   *sql.DB
)

type HomeData struct {
	ShowOnboarding bool
}

func main() {
	var err error

	// 1) Connexion à la base SQLite
	db, err = sql.Open("sqlite3", "backend/database/data.db")
	if err != nil {
		log.Fatal("Erreur ouverture DB:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Erreur ping DB:", err)
	}

	// 1.1) Initialisation du schéma si nécessaire
	if err := initSchema(db); err != nil {
		log.Fatal("Erreur init schema:", err)
	}

	// 2) Parsing des templates
	tmpl = template.Must(template.ParseGlob("frontend/web/templates/*.html"))

	// 3) Fichiers statiques (CSS, JS, images)
	fs := http.FileServer(http.Dir("frontend/web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 4) Routes pages
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/login", serveTemplate("login.html"))
	http.HandleFunc("/register", serveTemplate("register.html"))
	http.HandleFunc("/posts", serveTemplate("posts.html"))
	http.HandleFunc("/create-post", serveTemplate("create-post.html"))
	http.HandleFunc("/profile", serveTemplate("profile.html"))

	// 5) Route pour enregistrer l’onboarding
	http.HandleFunc("/onboarding", OnboardingHandler)

	log.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Initialise le schéma si la table "users" n'existe pas encore
func initSchema(db *sql.DB) error {
	var count int

	err := db.QueryRow(`
		SELECT count(*)
		FROM sqlite_master
		WHERE type='table' AND name='users'
	`).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Schema déjà présent, pas de réinitialisation de la DB")
		return nil
	}

	log.Println("Schema absent, exécution de backend/database/schema.sql")

	schemaBytes, err := os.ReadFile("backend/database/schema.sql")
	if err != nil {
		return err
	}

	schema := string(schemaBytes)
	statements := strings.Split(schema, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err := db.Exec(stmt)
		if err != nil {
			log.Printf("Erreur Exec schema sur : %q : %v\n", stmt, err)
			return err
		}
	}

	log.Println("Schema SQL exécuté avec succès")
	return nil
}

// Handler générique pour les pages simples
func serveTemplate(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.ExecuteTemplate(w, name, nil)
		if err != nil {
			http.Error(w, "Erreur template", http.StatusInternalServerError)
			log.Println("Erreur template", name, ":", err)
		}
	}
}

// TODO: à remplacer plus tard par une vraie récupération via cookie/session
func getUserIDFromSession(r *http.Request) int {
	// Pour tester l’onboarding : on simule un user ID 1
	return 1
}

// Handler de la home : décide si on affiche la popup onboarding ou pas
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromSession(r)

	var onboardingDone int
	err := db.QueryRow(`
		SELECT onboarding_done
		FROM user_preferences
		WHERE user_id = ?
	`, userID).Scan(&onboardingDone)

	if err == sql.ErrNoRows {
		_, err = db.Exec(`
			INSERT INTO user_preferences (
				user_id,
				games_played,
				play_style,
				game_types,
				looking_for_mates,
				skill_level,
				onboarding_done
			) VALUES (?, '', 'loisir', '', 0, '', 0)
		`, userID)
		if err != nil {
			log.Println("Erreur INSERT user_preferences:", err)
		}
		onboardingDone = 0
	} else if err != nil {
		log.Println("Erreur SELECT user_preferences:", err)
	}

	data := HomeData{
		ShowOnboarding: onboardingDone == 0,
	}

	err = tmpl.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, "Erreur template", http.StatusInternalServerError)
		log.Println("Erreur template home.html:", err)
	}
}

// Handler POST /onboarding : enregistre les préférences depuis la popup
func OnboardingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromSession(r)

	gamesPlayed := r.FormValue("games_played")
	playStyle := r.FormValue("play_style")
	gameTypes := r.FormValue("game_types")
	lookingForMates := r.FormValue("looking_for_mates")
	skillLevel := r.FormValue("skill_level")

	_, err := db.Exec(`
		UPDATE user_preferences
		SET games_played = ?,
			play_style = ?,
			game_types = ?,
			looking_for_mates = ?,
			skill_level = ?,
			onboarding_done = 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
	`, gamesPlayed, playStyle, gameTypes, lookingForMates, skillLevel, userID)

	if err != nil {
		log.Println("Erreur UPDATE user_preferences:", err)
		http.Error(w, "Erreur lors de l'enregistrement des préférences", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}