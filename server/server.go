package server

import (
	"fmt"
	petitbac "groupie/PetitBac"
	"groupie/blindtest"
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

func petitBacHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Petit Bac")
}

func blindTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		track, err := blindtest.GetRandomTrack()
		if err != nil {
			http.Error(w, "Erreur lors du chargement du son", http.StatusInternalServerError)
			return
		}

		data := struct {
			Preview string
			Answer  string
			Result  string
			Score   int
		}{
			Preview: track.Preview,
			Answer:  track.Title,
			Result:  "",
			Score:   1,
		}

		tmpl := template.Must(template.ParseFiles("_templates_/blindTest.html"))
		tmpl.Execute(w, data)

	} else if r.Method == http.MethodPost {
		r.ParseForm()
		guess := r.FormValue("guess")
		answer := r.FormValue("answer")
		score := 1
		var result string
		if blindtest.CheckAnswer(guess, answer) {
			score++
			result = fmt.Sprintf("✅ Bravo ! C'était bien : <strong>%s</strong>", answer)
		} else {
			result = fmt.Sprintf("❌ Mauvais ! La bonne réponse était : <strong>%s</strong>", answer)
		}

		data := struct {
			Preview string
			Answer  string
			Result  string
			Score   int
		}{
			Preview: "",
			Answer:  answer,
			Result:  result,
			Score:   score,
		}

		tmpl := template.Must(template.ParseFiles("_templates_/blindTest.html"))
		tmpl.Execute(w, data)
	}
}

func gameRoomHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	game, exists := petitbac.PetitBacGames[code]
	if !exists {
		http.NotFound(w, r)
		return
	}

	data := struct {
		Code       string
		Letter     string
		Categories []string
	}{
		Code:       game.Code,
		Letter:     game.Letter,
		Categories: []string{"Animal", "Ville", "Objet"},
	}

	tmpl := template.Must(template.ParseFiles("_templates_/game-room.html"))
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		fmt.Println(err)
	}
}

func guessHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("_templates_/guess-the-sound.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		fmt.Println(err)
	}
}

func createRoomHandler(w http.ResponseWriter, r *http.Request) {
	petitbac.Start(w, r)
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
	http.HandleFunc("/guess", guessHandler)
	http.HandleFunc("/petit", petitBacHandler)
	http.HandleFunc("/blind", blindTestHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/game-home", gameHomeHandler)
	http.HandleFunc("/game-room", gameRoomHandler)
	http.HandleFunc("/ws/game-home", controllers.GameWebSocket)
	http.HandleFunc("/create-room", createRoomHandler)
	http.HandleFunc("/submit-answer", petitbac.SubmitAnswer)

	fmt.Println("Serveur démarré sur le port 8080 ")
	http.ListenAndServe(":8080", nil)
}
