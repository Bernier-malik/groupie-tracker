package controllers

import (
	"database/sql"
	"groupie/db"
	"log"
	"net/http"
)


func getScoreByPlayers(pseudo string) (int, error) {
	var score int
	query := "SELECT score FROM players WHERE pseudo = ?"
	err := db.DB.QueryRow(query, pseudo).Scan(&score)
	if err != nil {
		return 0, err
	}
	return score, nil
}

func ScorBoard(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}


	score, err := getScoreByPlayers("coco")

	if err != nil && err != sql.ErrNoRows {
		log.Println("Database error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Score: " + string(score)))
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
