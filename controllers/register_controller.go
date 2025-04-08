package controllers

import (
	"database/sql"
	"fmt"
	"groupie/db"
	"log"
	"net/http"
	"strings"
)

// RegisterUser handles user registration
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	pseudo := r.FormValue("pseudo")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm-password")

	// Step 1: Check if passwords match
	if password != confirmPassword {
		fmt.Println("Les mots de passe ne correspondent pas!")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// Step 2: Check if username already exists
	var existingPseudo string
	queryCheck := "SELECT pseudo FROM players WHERE pseudo = ?"
	err = db.DB.QueryRow(queryCheck, pseudo).Scan(&existingPseudo)

	if err != nil && err != sql.ErrNoRows {
		log.Println("Database error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If the username exists
	if err == nil {
		fmt.Println("Le pseudo est déjà utilisé. Essayez un autre!")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	//Check if password resepected cnil
	if(password == "" || len(password) < 8 || len(password) > 20 || !strings.ContainsAny(password, "0123456789") || !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") || !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")) {
		fmt.Println("Le mot de passe doit contenir entre 8 et 20 caractères, au moins une lettre majuscule, une lettre minuscule et un chiffre.")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// Step 3: Insert new user into the database
	queryInsert := `INSERT INTO players (email, pseudo, password) VALUES (?, ?, ?)`
	_, err = db.DB.Exec(queryInsert, email, pseudo, password)
	if err != nil {
		log.Println("Error inserting user:", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Redirect to login page after successful registration
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
