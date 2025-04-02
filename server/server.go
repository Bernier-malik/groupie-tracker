package server

import (
	"fmt"
	"groupie/controllers"
	"groupie/db"
	"html"
	"html/template"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/home" {
		http.NotFound(w, r)
		fmt.Printf("Error: handler for %s not found\n", html.EscapeString(r.URL.Path))
		return
	}

	tmpl := template.Must(template.ParseFiles("_templates_/home.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Printf("erreur de template %s:", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.NotFound(w, r)
		fmt.Printf("Error: handler for %s not found\n", html.EscapeString(r.URL.Path))
		return
	}

	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("_templates_/login.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Printf("erreur de template %s:", err)

		}
	} else if r.Method == http.MethodPost {
		controllers.LoginUser(w, r)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register" {
		http.NotFound(w, r)
		fmt.Printf("Error: handler for %s not found\n", html.EscapeString(r.URL.Path))
		return
	}

	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("_templates_/register.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Printf("erreur de template %s:", err)
		}
	} else if r.Method == http.MethodPost {
		controllers.RegisterUser(w, r)
	}
}

func gameHomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/game-home" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.ParseFiles("_templates_/game-home.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		fmt.Println(err)
	}
}
































func guessHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("_templates_/guess-the-song.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		fmt.Println(err)
	}
	guess := controllers.GuessTheSong()
	fmt.Println(guess)
	if r.Method == http.MethodGet {
		fmt.Printf("non rien")
	} else if r.Method == http.MethodPost {
		fmt.Println(controllers.Checkrequet(w,r))
	}
}























func petitBacHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Petit Bac")
}

func BlindTestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Blind test")
}

func Start() {
	db.InitDB()
	defer db.CloseDB()

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("_templates_/css"))))
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("_templates_/"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "")
	})

	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/guess-the-song", guessHandler)
	http.HandleFunc("/petit", petitBacHandler)
	http.HandleFunc("/blind", BlindTestHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/game-home", gameHomeHandler)
	http.HandleFunc("/ws/game-home", controllers.GameWebSocket)

	fmt.Println("Serveur démarré sur le port 8080 ")
	http.ListenAndServe(":8080", nil)
}
