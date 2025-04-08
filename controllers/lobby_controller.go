package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ClientInfo struct {
	ClientID   string
	Pseudo     string
	Connection *websocket.Conn
}
type Game struct {
	GameID  string        `json:"id"`
	Clients []*ClientInfo `json:"clients"`
}

var games = make(map[string]*Game)

var clients = make(map[string]*ClientInfo)

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
	clients[clientID] = &ClientInfo{
		ClientID:   clientID,
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

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", clientID)
			// delete(clients, clientID)
			break
		}

		var result map[string]interface{}
		err = json.Unmarshal(msg, &result)

		if err != nil {
			fmt.Println("Invalid message from client")
			continue
		}
		method, ok := result["method"].(string)
		if !ok {
			continue
		}

		switch method {
		case "create":
			clientID, ok := result["clientId"].(string)
			if !ok {
				continue
			}
			gameID := uuid.New().String()

			fmt.Println("Game created with ID:", gameID)
			fmt.Println(pseudo, "with client ID", clientID, "created a new game")

			// Add client to new game
			game := &Game{
				GameID: gameID,
				Clients: []*ClientInfo{
					{ClientID: clientID, Pseudo: pseudo},
				},
			}
			games[gameID] = game

			// Send game info to client
			createPayload := map[string]interface{}{
				"method": "create",
				"gameId": gameID,
			}
			payloadBytes, _ := json.Marshal(createPayload)
			clients[clientID].Connection.WriteMessage(websocket.TextMessage, payloadBytes)
		}
	}

}

func ServeLobbyPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "_templates_/lobby-room.html")
}
