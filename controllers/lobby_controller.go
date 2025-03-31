package controllers

import (
	"fmt"
	"html"
	"html/template"
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
)

var lobbyConns []*websocket.Conn

var lobbyUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func generateCode() string {
	return fmt.Sprintf("%d", rand.Intn(9000)+1000)
}

func GameRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/game-room" {
		http.NotFound(w, r)
		fmt.Printf("Error: handler for %s not found\n", html.EscapeString(r.URL.Path))
		return
	}

	tmpl := template.Must(template.ParseFiles("_templates_/game-room.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Printf("Erreur de template: %s\n", err)
	}
}

// POST /create
func CreateLobbyHandler(w http.ResponseWriter, r *http.Request) {
	code := generateCode()
	http.Redirect(w, r, "/game-room?code="+code, http.StatusSeeOther)
}
