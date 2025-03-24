package main

import (
	"fmt"
	"groupie/server"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "PAGE ACCEUIL")
}
func gameHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Page de jeu")
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Page de connexion")
}
func main() {
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/login", loginHandler)
	server.Start()
}
