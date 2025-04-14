package server

import (
	"fmt"
	"groupie/controllers"
	"groupie/db"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

type PlayerState struct {
	GameID     string
	Pseudo     string
	Round      int
	Score      int
	AnswerChan chan bool
	RoundStart time.Time
	Conn       *websocket.Conn
}

type ScoreEntry struct {
	Pseudo string
	Score  int
}

var (
	upgrader           = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	games              = make(map[string]map[string]*PlayerState)
	gameMutex          = &sync.Mutex{}
	guess              = controllers.GuessTheSong()
	gameTemplate       = template.Must(template.ParseFiles("_templates_/guess-the-song.html"))
	scoreboardTemplate = template.Must(template.ParseFiles("_templates_/scoreboard.html"))
)

func startRoundTimer(ps *PlayerState) {
	ps.RoundStart = time.Now()
	ps.AnswerChan = make(chan bool)
	go func(round int, answerCh chan bool) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		timeout := time.After(30 * time.Second)
		remaining := 30

		for {
			select {
			//Chaque seconde, on décrémente le timer et l'envoie au joueur via WebSocket
			case <-ticker.C:
				remaining--
				if ps.Conn != nil {
					ps.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d", remaining)))
				}
			//Si 30 secondes sont passées sans réponse
			case <-timeout:
				gameMutex.Lock()
				if ps.Round == round {
					if ps.Round > 5 {
						db.DB.Exec("INSERT OR REPLACE INTO scores(gameId, pseudo, score) VALUES(?, ?, ?)", ps.GameID, ps.Pseudo, ps.Score)
					}
				}
				gameMutex.Unlock()
				return
			//Si le joueur a répondu, on arrête le timer immédiatement
			case <-answerCh:
				return
			}
		}
	}(ps.Round, ps.AnswerChan)
}

func guessTheSongWSHandler(w http.ResponseWriter, r *http.Request) {
	gameId := r.URL.Query().Get("gameId")
	cookie, err := r.Cookie("pseudo")
	if err != nil || gameId == "" {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}
	pseudo := cookie.Value

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	gameMutex.Lock()
	ps, ok := games[gameId][pseudo]
	if ok {
		ps.Conn = conn
	}
	gameMutex.Unlock()

	if !ok {
		conn.Close()
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for remaining := 30; remaining >= 0; remaining-- {
		err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d", remaining)))
		if err != nil {
			log.Println("WebSocket write error:", err)
			break
		}

		select {
		case <-time.After(1 * time.Second):
			// continue
		case <-ps.AnswerChan:
			conn.WriteMessage(websocket.TextMessage, []byte(" Temps arrêté"))
			conn.Close()
			return
		}
	}
	conn.WriteMessage(websocket.TextMessage, []byte("0"))
	conn.Close()
}
