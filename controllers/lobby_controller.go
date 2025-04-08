package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID         string
	Pseudo     string
	Connection *websocket.Conn
}

var clients = make(map[string]*Client)

func HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()
	cookie, err := r.Cookie("pseudo")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	pseudo := cookie.Value
	fmt.Println("WebSocket connected by:", pseudo)

	clientID := uuid.New().String()
	clients[clientID] = &Client{
		ID:         clientID,
		Pseudo:     pseudo,
		Connection: conn,
	}
	fmt.Println("New client connected. ID:", clientID, "| Pseudo:", pseudo)

	payload := map[string]string{
		"method":   "connect",
		"pseudo":   pseudo,
		"clientId": clientID,
	}
	payloadBytes, _ := json.Marshal(payload)
	conn.WriteMessage(websocket.TextMessage, payloadBytes)

}

func ServeLobbyPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "_templates_/lobby-room.html")
}
