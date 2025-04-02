package controllers

import (
	"fmt"
	"groupie/models"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func GameWebSocket(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code manquant", http.StatusBadRequest)
		return
	}

	lobby, ok := models.Lobbies[code]
	if !ok {
		http.Error(w, "Salle introuvable", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Erreur WebSocket:", err)
		return
	}
	defer conn.Close()

	lobby.Players = append(lobby.Players, &models.Player{Conn: conn})

	for _, p := range lobby.Players {
		p.Conn.WriteMessage(websocket.TextMessage, []byte("Un joueur a rejoint la salle "+code))
	}

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		for _, p := range lobby.Players {
			p.Conn.WriteMessage(msgType, msg)
		}
	}
}
