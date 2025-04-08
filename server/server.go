package server

import (
	"fmt"
	"groupie/controllers"
	"groupie/db"
	"html"
	"html/template"
	"net/http"
	"time"
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






































type Data struct {
	Parole string
	Tours  int
	Timer  int
}

var guess = controllers.GuessTheSong()
var tours = 0

func guessHandler(w http.ResponseWriter, r *http.Request) {
	data := Data{
		Parole: guess[tours].Lyrics,
		Tours:  tours + 1,
		Timer:  0,
	}

	go func() {
		stop := time.After(30 * time.Second)
		i := 0
		for {
			select {
			case <-stop:
				fmt.Println("EXIT: 30 seconds")
				return
			case <-time.After(1 * time.Second):
				fmt.Println(data.Timer,"second")
			}
			i++
			data.Timer = i
		}
		//tours++
		//data.Tours =tours
	}()
	

	if tours > 4 {
		tours = 0
	}

	if r.Method == http.MethodGet {
		data := Data{
			Parole: guess[tours].Lyrics,
			Tours:  tours + 1,
		}

		tmpl := template.Must(template.ParseFiles("_templates_/guess-the-song.html"))
		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			fmt.Println("Erreur template :", err)
		}
	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Erreur dans le formulaire", http.StatusBadRequest)
			fmt.Println("Erreur de parsing :", err)
			return
		}

		userResponse := r.FormValue("userReponse")
		fmt.Println("User response :", userResponse)

		correct := controllers.CheckRep(userResponse, guess[tours].Title)
		fmt.Println("Réponse correcte ?", correct)

		tours++

		data := Data{
			Parole: guess[tours].Lyrics,
			Tours:  tours + 1,
		}
		fmt.Println(guess[tours].Title)
		tmpl := template.Must(template.ParseFiles("_templates_/guess-the-song.html"))
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			fmt.Println("Erreur template :", err)
		}
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
