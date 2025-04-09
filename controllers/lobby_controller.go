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
	GameID    string        `json:"id"`
	Clients   []*ClientInfo `json:"clients"`
	CreatorID string        `json:"creatorID"`
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

	// Check if a client with the same pseudo is already connected
	exists := false
	for _, client := range clients {
		if client.Pseudo == pseudo {
			exists = true
			break
		}
	}

	if !exists {
		clientID := uuid.New().String()
		clients[clientID] = &ClientInfo{
			ClientID:   clientID,
			Pseudo:     pseudo,
			Connection: conn,
		}

		fmt.Println("New client connected. ID:", clientID, "| Pseudo:", pseudo)

		payload := map[string]interface{}{
			"method":   "connect",
			"pseudo":   pseudo,
			"clientId": clientID,
		}
		payloadBytes, _ := json.Marshal(payload)
		conn.WriteMessage(websocket.TextMessage, payloadBytes)
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", pseudo)
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
			fmt.Println(pseudo, "created a new game")

			// Add client to new game
			game := &Game{
				GameID:    gameID,
				CreatorID: clientID,
				Clients: []*ClientInfo{
					{ClientID: clientID, Pseudo: pseudo, Connection: conn},
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
		case "rejoin":
			gameID := result["gameId"].(string)
			clientID := result["clientId"].(string)

			game, ok := games[gameID]
			if !ok {
				fmt.Println("Game not found:", gameID)
				continue
			}

			// Get pseudo from cookie
			cookie, err := r.Cookie("pseudo")
			if err != nil {
				fmt.Println("Cookie error:", err)
				continue
			}
			pseudo := cookie.Value

			// Check if player is already in the game
			alreadyIn := false
			for _, c := range game.Clients {
				if c.Pseudo == pseudo {
					alreadyIn = true
					break
				}
			}

			if !alreadyIn {
				game.Clients = append(game.Clients, &ClientInfo{
					ClientID:   clientID,
					Pseudo:     pseudo,
					Connection: clients[clientID].Connection,
				})
				fmt.Println("🔁", pseudo, "rejoined game", gameID)
			}

			// Broadcast updated game to all clients
			updatePayload := map[string]interface{}{
				"method": "update",
				"game": map[string]interface{}{
					"id":        game.GameID,
					"creatorId": game.CreatorID,
					"clients":   game.Clients,
				},
			}
			fmt.Println(" List of players in game:")
			for _, c := range game.Clients {
				connStatus := "NOT FOUND"
				if storedClient, ok := clients[c.ClientID]; ok && storedClient.Connection != nil {
					if storedClient.Connection == c.Connection {
						connStatus = "Connection OK"
					} else {
						connStatus = "Mismatch"
					}
				}
				fmt.Printf("- ID: %s | Pseudo: %s | Conn: %s\n", c.ClientID, c.Pseudo, connStatus)
			}

			broadcastToGame(game, updatePayload)

		}
	}

}

func broadcastToGame(game *Game, payload map[string]interface{}) {
	msg, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}

	for _, client := range game.Clients {
		if c, ok := clients[client.ClientID]; ok {
			c.Connection.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func ServeLobbyPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "_templates_/lobby-room.html")
}
