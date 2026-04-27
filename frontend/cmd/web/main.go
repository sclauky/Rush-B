package main

import (
	"log"
	"net/http"
)

func serveTemplate(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

func main() {
	// Fichiers statiques (CSS, JS, images)
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Pages
	http.HandleFunc("/", serveTemplate("./web/templates/home.html"))
	http.HandleFunc("/login", serveTemplate("./web/templates/login.html"))
	http.HandleFunc("/register", serveTemplate("./web/templates/register.html"))
	http.HandleFunc("/posts", serveTemplate("./web/templates/posts.html"))
	http.HandleFunc("/create-post", serveTemplate("./web/templates/create-post.html"))
	http.HandleFunc("/profile", serveTemplate("./web/templates/profile.html"))

	log.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
