package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"io/ioutil"

	"github.com/PuerkitoBio/goquery"
)
//Structure de recheche
type SearchResponse struct {
	Tracks struct {
		Items []Track `json:"items"`
	} `json:"tracks"`
}

//Structure de l'api spotify
type Track struct {
	Album            Album          `json:"album"`
	Artists          []Artist       `json:"artists"`
	AvailableMarkets []string       `json:"available_markets"`
	DiscNumber       int            `json:"disc_number"`
	DurationMs       int            `json:"duration_ms"`
	Explicit         bool           `json:"explicit"`
	ExternalIDs      ExternalIDs    `json:"external_ids"`
	ExternalURLs     ExternalURLs   `json:"external_urls"`
	Href             string         `json:"href"`
	ID               string         `json:"id"`
	IsPlayable       bool           `json:"is_playable"`
	LinkedFrom       json.RawMessage `json:"linked_from"` // champ vide ou inconnu
	Restrictions     *Restrictions  `json:"restrictions,omitempty"`
	Name             string         `json:"name"`
	Popularity       int            `json:"popularity"`
	PreviewURL       string         `json:"preview_url"`
	TrackNumber      int            `json:"track_number"`
	Type             string         `json:"type"`
	URI              string         `json:"uri"`
	IsLocal          bool           `json:"is_local"`
}

type Album struct {
	AlbumType            string         `json:"album_type"`
	TotalTracks          int            `json:"total_tracks"`
	AvailableMarkets     []string       `json:"available_markets"`
	ExternalURLs         ExternalURLs   `json:"external_urls"`
	Href                 string         `json:"href"`
	ID                   string         `json:"id"`
	Images               []Image        `json:"images"`
	Name                 string         `json:"name"`
	ReleaseDate          string         `json:"release_date"`
	ReleaseDatePrecision string         `json:"release_date_precision"`
	Restrictions         *Restrictions  `json:"restrictions,omitempty"`
	Type                 string         `json:"type"`
	URI                  string         `json:"uri"`
	Artists              []Artist       `json:"artists"`
}

type Artist struct {
	ExternalURLs ExternalURLs `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}

type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

type ExternalIDs struct {
	ISRC string `json:"isrc"`
	EAN  string `json:"ean"`
	UPC  string `json:"upc"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type Restrictions struct {
	Reason string `json:"reason"`
}


type TrackWithPreview struct {
	TrackID    string `json:"track_id"`
	TrackName  string `json:"track_name"`
	ArtistName string `json:"artist_name"`
	PreviewURL string `json:"preview_url"`
}

type GeniusSearchResponse struct {
	Response struct {
		Hits []struct {
			Result struct {
				FullTitle string `json:"full_title"`
				URL       string `json:"url"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

//Get track info from spotify
func getSpotifyTrackId(token string, trackId string) (Track, error){
	url := "https://api.spotify.com/v1/tracks/" + trackId

	// Création de la requête
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Attention : ne pas mettre ":" dans le nom du header
	req.Header.Set("Authorization", "Bearer "+token)

	// Envoi de la requête avec un client HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Lecture de la réponse
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	// Unmarshal du JSON dans ta structure
	var track Track
	jsonErr := json.Unmarshal(body, &track)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return track, nil
}

func getLyricsFromGeniusPage(pageURL string) (string, error) {
	res, err := http.Get(pageURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("Erreur HTTP %d via l'url %s", res.StatusCode, pageURL)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	lyrics := ""
	doc.Find("div[data-lyrics-container='true']").Each(func(i int, s *goquery.Selection) {
		lyrics += s.Text() + "\n"
	})

	if lyrics == "" {
		return "", fmt.Errorf("Paroles non trouvees")
	}

	return strings.TrimSpace(lyrics), nil
}

func searchLyricsOnGenius(title, artist, geniusToken string) (string, error) {
	query := fmt.Sprintf("%s %s", title, artist)
	searchURL := "https://api.genius.com/search?q=" + url.QueryEscape(query)

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("Authorization", "Bearer "+geniusToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result GeniusSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Response.Hits) == 0 {
		return "", fmt.Errorf("Aucun resultat trouve pour %s", query)
	}

	songURL := result.Response.Hits[0].Result.URL
	return getLyricsFromGeniusPage(songURL)
}

func  delet(str string) string {
	var frames int = len(str)
	var result string = ""
	var in bool = true
	for i:= 0; i<frames; i++ {
		if in == true{
			if str[i] == '[' {
				in = false 
			} else {
				result += str[i]
			}
		} else if in == false {
			if str[i] == ']' {
				in = true
			}
		}
		
	}
	return result
}

func main() {
	trackName := "One More Time"
	artistName := "Daft Punk"

	geniusToken := "kqpwlWVEknmSRiSnXiFLtXbFW9pv0Nn92i9jWe9qywhY8jkD0W7TaHYwDxLSigYz"

	lyrics, err := searchLyricsOnGenius(trackName, artistName, geniusToken)
	if err != nil {
		log.Fatal("Erreur :", err)
	}

	//fmt.Println("Track name : ", trackName, " artiste : ", artistName)
	//fmt.Println("----------------------------------------")
	fmt.Println(delet(lyrics))
}
