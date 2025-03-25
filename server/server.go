package server

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.ParseGlob("_templates_/*.html"))

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/home" {
		http.NotFound(w, r)
		fmt.Printf("Error: handler for %s not found\n", html.EscapeString(r.URL.Path))
		return
	}

	err := tmpl.ExecuteTemplate(w, "pages/home/home.html", nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Printf("Template error: %v\n", err)
	}
}

func guessSoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Guess The Sound")
}
func petitBacHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Petit Bac")
}
func BlindTestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Blind test")
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Page de connexion")
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Inscrivez vous")
}

func Start() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "")
	})

	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/guess", guessSoundHandler)
	http.HandleFunc("/petit", petitBacHandler)
	http.HandleFunc("/blind", BlindTestHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)

	fmt.Println("Serveur démarré sur le port 8080 ")
	http.ListenAndServe(":8080", nil)
}
