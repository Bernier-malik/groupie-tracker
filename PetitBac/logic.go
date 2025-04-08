package PetitBac

import (
	"math/rand"
	"net/http"
	"strings"
	"sync"
)

type Game struct {
	Code       string
	Letter     string
	Categories []string
	Answers    map[string]map[string]string
	Scores     map[string]int
}

var PetitBacGames = make(map[string]*Game)
var mutex sync.Mutex

func randomLetter() string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return string(letters[rand.Intn(len(letters))])
}

func generateGameCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := ""
	for i := 0; i < 6; i++ {
		code += string(charset[rand.Intn(len(charset))])
	}
	return code
}

func Start(w http.ResponseWriter, r *http.Request) {
	code := generateGameCode()
	letter := randomLetter()
	game := &Game{
		Code:       code,
		Letter:     letter,
		Categories: []string{"Animal", "Ville", "Objet"},
		Answers:    make(map[string]map[string]string),
		Scores:     make(map[string]int),
	}

	mutex.Lock()
	PetitBacGames[code] = game
	mutex.Unlock()

	http.Redirect(w, r, "/game-room?code="+code, http.StatusSeeOther)
}

func SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	code := r.FormValue("code")
	pseudo := r.FormValue("pseudo")

	mutex.Lock()
	game, exists := PetitBacGames[code]
	mutex.Unlock()

	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	userAnswers := make(map[string]string)
	score := 0

	for _, category := range game.Categories {
		answer := r.FormValue(category)
		userAnswers[category] = answer

		if IsValidAnswer(answer, game.Letter) {
			score++
		}
	}

	mutex.Lock()
	game.Answers[pseudo] = userAnswers
	game.Scores[pseudo] = score
	mutex.Unlock()

}

func IsValidAnswer(answer, letter string) bool {
	answer = strings.TrimSpace(strings.ToUpper(answer))
	return strings.HasPrefix(answer, strings.ToUpper(letter))
}
