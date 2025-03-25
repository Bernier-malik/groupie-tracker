
//curl -X POST "https://accounts.spotify.com/api/token"      -H "Content-Type: application/x-www-form-urlencoded"      -d "grant_type=client_credentials&client_id=454055576c254e1e802bf0b10f687298&client_secret=3fe51de5734a4052bceac38e4cb35a42"
package main

import (
	"context"
	"log"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	authConfig := &clientcredentials.Config{
		ClientID:     "454055576c254e1e802bf0b10f687298",
		ClientSecret: "3fe51de5734a4052bceac38e4cb35a42",
		TokenURL:     spotify.TokenURL,
	}

	accessToken, err := authConfig.Token(context.Background())
	if err != nil {
		log.Fatalf("error retrieve access token: %v", err)
	}

	client := spotify.Authenticator{}.NewClient(accessToken)

	playlistID := spotify.ID("0s3c0flbAQIvlvG5v1d7X8")
	playlist, err := client.GetPlaylist(playlistID)
	if err != nil {
		log.Fatalf("error retrieve playlist data: %v", err)
	}
	
	log.Println("playlist name:", playlist.Name)
}