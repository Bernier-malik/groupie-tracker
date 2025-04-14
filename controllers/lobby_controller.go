package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
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

	for id, client := range clients {
		if client.Pseudo == pseudo {
			fmt.Println("Removing old connection for pseudo:", pseudo, "with ID:", id)
			delete(clients, id)
			break
		}
	}

	clientID := uuid.New().String()
	clients[clientID] = &ClientInfo{
		ClientID:   clientID,
		Pseudo:     pseudo,
		Connection: conn,
	}

	fmt.Println("client connected. ID:", clientID, "| Pseudo:", pseudo)

	payload := map[string]interface{}{
		"method":   "connect",
		"pseudo":   pseudo,
		"clientId": clientID,
	}
	payloadBytes, _ := json.Marshal(payload)
	conn.WriteMessage(websocket.TextMessage, payloadBytes)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", pseudo)
			//delete(clients, clientID)
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
			conn.WriteMessage(websocket.TextMessage, payloadBytes)
		case "start":
			gameID := result["gameId"].(string)
			gameType := result["gameType"].(string)
			pseudo := result["pseudo"].(string)

			game, ok := games[gameID]
			if !ok {
				fmt.Println("Game not found:", gameID)
				continue
			}

			if len(game.Clients) < 2 {
				warning := map[string]interface{}{
					"method":  "alert",
					"message": "Il faut au moins 2 joueurs pour démarrer la partie.",
				}
				payloadBytes, _ := json.Marshal(warning)
				conn.WriteMessage(websocket.TextMessage, payloadBytes)
				continue
			}
			var url string
			switch gameType {
			case "guess-the-song":
				url = fmt.Sprintf("/guess-the-song?gameId=%s", gameID)
			case "petit-bac":
				url = fmt.Sprintf("/petit-bac?gameId=%s", gameID)
			case "blind-test":
				url = fmt.Sprintf("/blind?gameId=%s&pseudo=%s&game=%s", gameID, pseudo, gameType)
			default:
				url = fmt.Sprintf("/unknown?id=%s", gameID)
			}

			startPayload := map[string]interface{}{
				"method": "redirect",
				"url":    url,
			}

			broadcastToGame(game, startPayload)
		case "leave":
			gameID := result["gameId"].(string)
			pseudo := result["pseudo"].(string)

			game, ok := games[gameID]
			if !ok {
				fmt.Println("Game not found:", gameID)
				continue
			}

			newClients := []*ClientInfo{}
			for _, c := range game.Clients {
				if c.Pseudo != pseudo {
					newClients = append(newClients, c)
				}
			}
			game.Clients = newClients
			fmt.Println(pseudo, "has left the game:", gameID)

			if len(game.Clients) == 0 {
				delete(games, gameID)
				fmt.Println("Game", gameID, "deleted because it's empty.")
				break
			}
			visibleClients := []map[string]string{}
			for _, c := range game.Clients {
				visibleClients = append(visibleClients, map[string]string{
					"pseudo":   c.Pseudo,
					"clientId": c.ClientID,
				})
			}

			updatePayload := map[string]interface{}{
				"method": "update",
				"game": map[string]interface{}{
					"id":        game.GameID,
					"creatorId": game.CreatorID,
					"clients":   visibleClients,
				},
			}
			broadcastToGame(game, updatePayload)

		case "rejoin":
			gameID := result["gameId"].(string)
			pseudo := result["pseudo"].(string)

			game, ok := games[gameID]
			if !ok {
				fmt.Println("Game not found:", gameID)
				continue
			}

			// Check if player is already in the game
			alreadyIn := false
			for _, c := range game.Clients {
				if c.Pseudo == pseudo {
					alreadyIn = true
					break
				}
			}

			if !alreadyIn {
				var foundClient *ClientInfo
				for _, client := range clients {
					if client.Pseudo == pseudo {
						foundClient = client
						break
					}
				}

				if foundClient == nil || foundClient.Connection == nil {
					fmt.Println(" Connection not found for pseudo:", pseudo)
					continue
				}

				game.Clients = append(game.Clients, &ClientInfo{
					ClientID:   foundClient.ClientID,
					Pseudo:     foundClient.Pseudo,
					Connection: foundClient.Connection,
				})

				fmt.Println(pseudo, "rejoined game", gameID)
			}

			var simpleClients []map[string]string
			for _, c := range game.Clients {
				simpleClients = append(simpleClients, map[string]string{
					"clientId": c.ClientID,
					"pseudo":   c.Pseudo,
				})
			}

			updatePayload := map[string]interface{}{
				"method": "update",
				"game": map[string]interface{}{
					"id":        game.GameID,
					"creatorId": game.CreatorID,
					"clients":   simpleClients,
				},
			}

			fmt.Println(" List of players in game:")
			for _, c := range game.Clients {
				fmt.Printf(" Pseudo: %s\n", c.Pseudo)
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

	for _, gameClient := range game.Clients {
		var foundClient *ClientInfo

		for _, c := range clients {
			if c.Pseudo == gameClient.Pseudo {
				foundClient = c
				break
			}
		}

		if foundClient == nil || foundClient.Connection == nil {
			fmt.Println("Connection not found for:", gameClient.Pseudo)
			continue
		}

		err := foundClient.Connection.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Println("Error sending to:", gameClient.Pseudo, "-", err)
		}
	}
}

func ServeLobbyPage(w http.ResponseWriter, r *http.Request) {
	gameType := r.URL.Query().Get("game")
	tmpl := template.Must(template.ParseFiles("_templates_/lobby-room.html"))
	tmpl.Execute(w, map[string]string{
		"GameType": gameType,
	})
}
