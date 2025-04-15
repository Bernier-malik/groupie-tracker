package server

import (
	"fmt"
	"groupie/blindtest"
	"groupie/controllers"
	"groupie/db"
	"html"
	"html/template"
	"log"
	"net/http"
	"strings"
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

var blindTracks []blindtest.Track

func init() {
	tracks, err := blindtest.GetTracks()
	if err != nil {
		log.Fatal(" Failed to load blind test tracks:", err)
	}
	blindTracks = tracks
}

func blindTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		gameId := r.URL.Query().Get("gameId")
		cookie, err := r.Cookie("pseudo")
		if err != nil || gameId == "" {
			http.Error(w, "Invalid game or player", http.StatusBadRequest)
			return
		}
		pseudo := cookie.Value

		gameMutex.Lock()
		if games[gameId] == nil {
			games[gameId] = make(map[string]*PlayerState)
		}
		ps, exists := games[gameId][pseudo]
		if !exists {
			ps = &PlayerState{GameID: gameId, Pseudo: pseudo, Round: 1, Score: 0}
			games[gameId][pseudo] = ps
		}

		if ps.Round > 5 {
			gameMutex.Unlock()
			http.Redirect(w, r, "/scoreboard?gameId="+gameId, http.StatusSeeOther)
			return
		}

		currentRound := ps.Round
		track := blindTracks[(currentRound-1)%len(blindTracks)]

		ps.RoundStart = time.Now()
		ps.AnswerChan = make(chan bool)
		go startRoundTimer(ps)
		gameMutex.Unlock()

		elapsed := time.Since(ps.RoundStart)
		remaining := 30 - int(elapsed.Seconds())
		if remaining < 0 {
			remaining = 0
		}

		data := struct {
			GameID   string
			Round    int
			Preview  string
			Answer   string
			Timer    int
			NextURL  string
			GameOver bool
			Score    int
		}{
			GameID:   gameId,
			Round:    currentRound,
			Preview:  track.Preview,
			Answer:   track.Title,
			Timer:    remaining,
			NextURL:  "/blind?gameId=" + gameId,
			GameOver: false,
			Score:    ps.Score,
		}

		tmpl := template.Must(template.ParseFiles("_templates_/blindTest.html"))
		tmpl.Execute(w, data)

	} else if r.Method == http.MethodPost {
		gameId := r.URL.Query().Get("gameId")
		cookie, err := r.Cookie("pseudo")
		if err != nil || gameId == "" {
			http.Error(w, "Invalid game or player", http.StatusBadRequest)
			return
		}
		pseudo := cookie.Value
		userAnswer := r.FormValue("userReponse")
		correctAnswer := r.FormValue("correctAnswer")

		gameMutex.Lock()
		ps, ok := games[gameId][pseudo]
		if !ok {
			gameMutex.Unlock()
			http.Error(w, "Game session not found", http.StatusBadRequest)
			return
		}

		currentRound := ps.Round
		if currentRound > 5 || currentRound != ps.Round {
			gameMutex.Unlock()
			http.Redirect(w, r, "/blind?gameId="+gameId, http.StatusSeeOther)
			return
		}

		correct := blindtest.CheckAnswer(userAnswer, correctAnswer)
		if correct {
			elapsed := time.Since(ps.RoundStart)
			remaining := int(30 - elapsed.Seconds())
			if remaining < 0 {
				remaining = 0
			}
			ps.Score += remaining
		}

		ps.Round++
		if ps.Round > 5 {
			_, _ = db.DB.Exec("INSERT OR REPLACE INTO scores(gameId, pseudo, score) VALUES(?, ?, ?)", gameId, pseudo, ps.Score)
			go broadcastScoreboard(gameId)
		}

		if ps.AnswerChan != nil {
			select {
			case ps.AnswerChan <- correct:
			default:
			}
		}
		gameMutex.Unlock()

		if ps.Round > 5 {
			http.Redirect(w, r, "/scoreboard?gameId="+gameId, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/blind?gameId="+gameId, http.StatusSeeOther)
		}
	}
}

func guessHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		gameId := r.URL.Query().Get("gameId")
		if gameId == "" {
			http.Error(w, "gameId required", http.StatusBadRequest)
			return
		}
		cookie, err := r.Cookie("pseudo")
		if err != nil {
			http.Error(w, "Player pseudo not found in cookie", http.StatusUnauthorized)
			return
		}
		pseudo := cookie.Value

		gameMutex.Lock()
		if games[gameId] == nil {
			games[gameId] = make(map[string]*PlayerState)
		}
		ps, exists := games[gameId][pseudo]
		if !exists {
			ps = &PlayerState{GameID: gameId, Pseudo: pseudo, Round: 1, Score: 0}
			games[gameId][pseudo] = ps
		}

		if ps.Round > 5 {
			gameMutex.Unlock()
			http.Redirect(w, r, "/scoreboard?gameId="+gameId, http.StatusSeeOther)
			return
		}

		currentRound := ps.Round
		lyricData := guess[currentRound-1].Lyrics
		Titre := guess[currentRound-1].Title
		fmt.Println("correct answer is :", Titre)
		startRoundTimer(ps)
		gameMutex.Unlock()
		elapsed := time.Since(ps.RoundStart)
		remaining := 30 - int(elapsed.Seconds())
		if remaining < 0 {
			remaining = 0
		}
		data := struct {
			GameID  string
			Round   int
			Lyric   string
			Timer   int
			NextURL string
		}{
			GameID: gameId,
			Round:  currentRound,
			Lyric:  lyricData,
			Timer:  remaining,
		}
		if currentRound < 5 {
			data.NextURL = "/guess-the-song?gameId=" + gameId
		} else {
			data.NextURL = "/scoreboard?gameId=" + gameId
		}
		if err := gameTemplate.Execute(w, data); err != nil {
			log.Println("Template execution error:", err)
		}
	} else if r.Method == http.MethodPost {
		gameId := r.URL.Query().Get("gameId")
		cookie, err := r.Cookie("pseudo")
		if err != nil || gameId == "" {
			http.Error(w, "Invalid game or player", http.StatusBadRequest)
			return
		}
		pseudo := cookie.Value
		userAnswer := r.FormValue("userReponse")
		gameMutex.Lock()
		ps, ok := games[gameId][pseudo]
		if !ok {
			gameMutex.Unlock()
			http.Error(w, "Game session not found", http.StatusBadRequest)
			return
		}
		currentRound := ps.Round
		if currentRound > 5 || currentRound != ps.Round {
			gameMutex.Unlock()
			http.Redirect(w, r, "/guess-the-song?gameId="+gameId, http.StatusSeeOther)
			return
		}
		lyricData := guess[currentRound-1].Lyrics
		correct := false
		if lyricData != "" {
			answer := guess[currentRound-1].Title
			if strings.TrimSpace(strings.ToLower(userAnswer)) == strings.ToLower(answer) {
				correct = true
			}
		}
		if correct {
			elapsed := time.Since(ps.RoundStart)
			remainingSec := int(30 - elapsed.Seconds())
			if remainingSec < 0 {
				remainingSec = 0
			}
			ps.Score += remainingSec
		}
		ps.Round++
		if ps.Round > 5 {
			_, _ = db.DB.Exec("INSERT OR REPLACE INTO scores(gameId, pseudo, score) VALUES(?, ?, ?)",
				gameId, pseudo, ps.Score)
			go broadcastScoreboard(gameId)
		}
		if ps.AnswerChan != nil {
			select {
			case ps.AnswerChan <- correct:
			default:
			}
		}
		gameMutex.Unlock()

		if ps.Round > 5 {
			http.Redirect(w, r, "/scoreboard?gameId="+gameId, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/guess-the-song?gameId="+gameId, http.StatusSeeOther)
		}
	}
}

func scoreboardHandler(w http.ResponseWriter, r *http.Request) {
	gameId := r.URL.Query().Get("gameId")
	if gameId == "" {
		http.Error(w, "gameId required", http.StatusBadRequest)
		return
	}
	// Query all players in this game, order by score ascending
	rows, err := db.DB.Query("SELECT pseudo, score FROM scores WHERE gameId = ? ORDER BY score DESC", gameId)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var scores []ScoreEntry
	for rows.Next() {
		var entry ScoreEntry
		if err := rows.Scan(&entry.Pseudo, &entry.Score); err == nil {
			scores = append(scores, entry)
		}
	}

	data := struct {
		GameID string
		Scores []ScoreEntry
	}{
		GameID: gameId,
		Scores: scores,
	}
	fmt.Println(data.Scores)
	if err := scoreboardTemplate.Execute(w, data); err != nil {
		log.Println("Template error:", err)
	}
}

type ReponseSet struct { // pour Petit bac
	Album      string
	Groupe     string
	Instrument string
	Featuring  string
}

var reponses = make(map[string]map[int]map[string]ReponseSet)

func BlindTestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Blind test")
}

func gameRoomHandler(w http.ResponseWriter, r *http.Request) {
}

func createRoomHandler(w http.ResponseWriter, r *http.Request) {
}

func Start() {
	db.InitDB()
	defer db.CloseDB()

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("_templates_/css"))))
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("_templates_/"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/guess-the-song", guessHandler)
	http.HandleFunc("/guess-the-song/ws", guessTheSongWSHandler)
	http.HandleFunc("/blind", blindTestHandler)
	http.HandleFunc("/blind/ws", blindTestWSHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/game-home", gameHomeHandler)
	http.HandleFunc("/game-room", gameRoomHandler)
	http.HandleFunc("/ws/game-home", controllers.GameWebSocket)
	http.HandleFunc("/create-room", createRoomHandler)
	http.HandleFunc("/lobby", controllers.ServeLobbyPage)
	http.HandleFunc("/lobby/ws", controllers.HandleWS)
	http.HandleFunc("/waiting-room", controllers.ServeWaitingRoom)
	http.HandleFunc("/scoreboard", scoreboardHandler)
	http.HandleFunc("/scoreboard/ws", scoreboardWSHandler)
	http.HandleFunc("/petitbac", petitBacHandler)
	http.HandleFunc("/petitbac/ws", petitBacWSHandler)

	fmt.Println("Serveur démarré sur le port 8080 ")
	http.ListenAndServe(":8080", nil)
}
