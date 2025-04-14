package controllers

import (
	"fmt"
	"html/template"
	"net/http"
)

func ServeWaitingRoom(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("id")
	if gameID == "" {
		http.Error(w, "Missing game ID", http.StatusBadRequest)
		return
	}

	clientID := r.URL.Query().Get("client")
	if clientID == "" {
		http.Error(w, "Missing client ID", http.StatusBadRequest)
		return
	}

	game, ok := games[gameID]
	if !ok {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	cookie, err := r.Cookie("pseudo")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	pseudo := cookie.Value

	found := false
	for _, c := range game.Clients {
		if c.Pseudo == pseudo {
			found = true
			break
		}
	}

	if !found {
		game.Clients = append(game.Clients, &ClientInfo{
			ClientID: clientID,
			Pseudo:   pseudo,
		})
		fmt.Println("✅", pseudo, "a rejoint la salle", gameID)
	}

	tmpl, err := template.ParseFiles("_templates_/waiting-room.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, game)
}
