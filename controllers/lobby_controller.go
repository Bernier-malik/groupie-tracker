package controllers

import (
	"fmt"
	"groupie/models"
	"math/rand"
	"net/http"
)

func generateCode() string {
	return fmt.Sprintf("%d", rand.Intn(9000)+1000)
}

func CreateLobbyHandler(w http.ResponseWriter, r *http.Request) {
	code := generateCode()

	models.Lobbies[code] = &models.Lobby{
		Code:    code,
		Players: []*models.Player{},
	}

	fmt.Println("Lobby créé avec code:", code)
	http.Redirect(w, r, "/game-room?code="+code, http.StatusSeeOther)
}
