package server

import (
	"encoding/json"
	"groupie/db"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var scoreboardClients = make(map[string][]*websocket.Conn)
var scoreboardMutex = &sync.Mutex{}

func scoreboardWSHandler(w http.ResponseWriter, r *http.Request) {
	gameId := r.URL.Query().Get("gameId")
	if gameId == "" {
		http.Error(w, "gameId required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket error:", err)
		return
	}

	scoreboardMutex.Lock()
	scoreboardClients[gameId] = append(scoreboardClients[gameId], conn)
	scoreboardMutex.Unlock()

	
	for {
		if _, _, err := conn.NextReader(); err != nil {
			conn.Close()
			break
		}
	}
}

func broadcastScoreboard(gameId string) {
	rows, err := db.DB.Query("SELECT pseudo, score FROM scores WHERE gameId = ? ORDER BY score DESC", gameId)
	if err != nil {
		log.Println("Erreur récupération score:", err)
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

	jsonScores, _ := json.Marshal(scores)

	scoreboardMutex.Lock()
	for _, conn := range scoreboardClients[gameId] {
		conn.WriteMessage(websocket.TextMessage, jsonScores)
	}
	scoreboardMutex.Unlock()
}
