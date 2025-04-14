package server

import (
	"fmt"
	"groupie/db"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type PetitBacState struct {
	GameID     string
	Pseudo     string
	Round      int
	Score      int
	AnswerChan chan bool
	RoundStart time.Time
	Conn       *websocket.Conn
}

var (
	petitBacTemplate = template.Must(template.ParseFiles("_templates_/lePetitBac.html"))
	petitBacGames    = make(map[string]map[string]*PetitBacState)
	petitBacMutex    = &sync.Mutex{}
	roundLetters     = make(map[string][]string)
)

func generateLetters() []string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var result []string
	for i := 0; i < 5; i++ {
		result = append(result, string(letters[rand.Intn(len(letters))]))
	}
	return result
}

func startPetitBacTimer(ps *PetitBacState) {
	ps.RoundStart = time.Now()
	ps.AnswerChan = make(chan bool)

	go func(round int, answerCh chan bool) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		timeout := time.After(30 * time.Second)
		remaining := 30

		for {
			select {
			case <-ticker.C:
				remaining--
				if ps.Conn != nil {
					ps.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d", remaining)))
				}
			case <-timeout:
				return
			case <-answerCh:
				return
			}
		}
	}(ps.Round, ps.AnswerChan)
}

func petitBacHandler(w http.ResponseWriter, r *http.Request) {
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

		petitBacMutex.Lock()
		if petitBacGames[gameId] == nil {
			petitBacGames[gameId] = make(map[string]*PetitBacState)
		}
		if _, ok := roundLetters[gameId]; !ok {
			roundLetters[gameId] = generateLetters()
		}
		ps, exists := petitBacGames[gameId][pseudo]
		if !exists {
			ps = &PetitBacState{GameID: gameId, Pseudo: pseudo, Round: 1, Score: 0}
			petitBacGames[gameId][pseudo] = ps
		}
		if ps.Round > 5 {
			petitBacMutex.Unlock()
			http.Redirect(w, r, "/scoreboard?gameId="+gameId, http.StatusSeeOther)
			return
		}

		letter := roundLetters[gameId][ps.Round-1]
		currentRound := ps.Round
		startPetitBacTimer(ps)
		petitBacMutex.Unlock()

		elapsed := time.Since(ps.RoundStart)
		remaining := 30 - int(elapsed.Seconds())
		if remaining < 0 {
			remaining = 0
		}

		data := struct {
			GameID  string
			Round   int
			Timer   int
			Letter  string
			NextURL string
		}{
			GameID:  gameId,
			Round:   currentRound,
			Timer:   remaining,
			Letter:  letter,
			NextURL: "/petitbac?gameId=" + gameId,
		}

		if err := petitBacTemplate.Execute(w, data); err != nil {
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

		r.ParseForm()
		album := r.FormValue("album")
		groupe := r.FormValue("groupe")
		instrument := r.FormValue("instrument")
		featuring := r.FormValue("featuring")

		petitBacMutex.Lock()
		ps, ok := petitBacGames[gameId][pseudo]
		if !ok {
			petitBacMutex.Unlock()
			http.Error(w, "Game session not found", http.StatusBadRequest)
			return
		}

		// --- STOCKAGE DES RÉPONSES ---
		if reponses[gameId] == nil {
			reponses[gameId] = make(map[int]map[string]ReponseSet)
		}
		if reponses[gameId][ps.Round] == nil {
			reponses[gameId][ps.Round] = make(map[string]ReponseSet)
		}
		reponses[gameId][ps.Round][pseudo] = ReponseSet{
			Album:      album,
			Groupe:     groupe,
			Instrument: instrument,
			Featuring:  featuring,
		}

		// --- CALCUL DES SCORES ---
		answers := reponses[gameId][ps.Round]
		myAnswers := answers[pseudo]
		letter := strings.ToLower(roundLetters[gameId][ps.Round-1])

		isValid := func(val string) bool {
			val = strings.ToLower(strings.TrimSpace(val))
			return strings.HasPrefix(val, letter) && len(val) > 1
		}

		countOccurrences := func(field func(ReponseSet) string, target string) int {
			count := 0
			for p, r := range answers {
				if p != pseudo && strings.EqualFold(strings.TrimSpace(field(r)), strings.TrimSpace(target)) {
					count++
				}
			}
			return count
		}

		score := 0

		if isValid(myAnswers.Album) {
			if countOccurrences(func(r ReponseSet) string { return r.Album }, myAnswers.Album) == 0 {
				score += 2
			} else {
				score += 1
			}
		}
		if isValid(myAnswers.Groupe) {
			if countOccurrences(func(r ReponseSet) string { return r.Groupe }, myAnswers.Groupe) == 0 {
				score += 2
			} else {
				score += 1
			}
		}
		if isValid(myAnswers.Instrument) {
			if countOccurrences(func(r ReponseSet) string { return r.Instrument }, myAnswers.Instrument) == 0 {
				score += 2
			} else {
				score += 1
			}
		}
		if isValid(myAnswers.Featuring) {
			if countOccurrences(func(r ReponseSet) string { return r.Featuring }, myAnswers.Featuring) == 0 {
				score += 2
			} else {
				score += 1
			}
		}

		ps.Score += score
		ps.Round++

		if ps.Round > 5 {
			db.DB.Exec("INSERT OR REPLACE INTO scores(gameId, pseudo, score) VALUES(?, ?, ?)", gameId, pseudo, ps.Score)
			go broadcastScoreboard(gameId)
		}

		if ps.AnswerChan != nil {
			select {
			case ps.AnswerChan <- true:
			default:
			}
		}

		petitBacMutex.Unlock()

		if ps.Round > 5 {
			http.Redirect(w, r, "/scoreboard?gameId="+gameId, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/petitbac?gameId="+gameId, http.StatusSeeOther)
		}
	}
}

func petitBacWSHandler(w http.ResponseWriter, r *http.Request) {
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

	petitBacMutex.Lock()
	ps, ok := petitBacGames[gameId][pseudo]
	if ok {
		ps.Conn = conn
	}
	petitBacMutex.Unlock()

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
		case <-ps.AnswerChan:
			conn.WriteMessage(websocket.TextMessage, []byte("Temps arrêté"))
			conn.Close()
			return
		}
	}
	conn.WriteMessage(websocket.TextMessage, []byte("0"))
	conn.Close()
}
