package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Structure de recheche
type SearchResponse struct {
	Tracks struct {
		Items []Track `json:"items"`
	} `json:"tracks"`
}

// Structure de l'api Deezer
type Track struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Duration      int    `json:"duration"`
	Public        bool   `json:"public"`
	IsLovedTrack  bool   `json:"is_loved_track"`
	Collaborative bool   `json:"collaborative"`
	NbTracks      int    `json:"nb_tracks"`
	Fans          int    `json:"fans"`
	Link          string `json:"link"`
	Share         string `json:"share"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXl     string `json:"picture_xl"`
	Checksum      string `json:"checksum"`
	Tracklist     string `json:"tracklist"`
	CreationDate  string `json:"creation_date"`
	AddDate       string `json:"add_date"`
	ModDate       string `json:"mod_date"`
	Md5Image      string `json:"md5_image"`
	PictureType   string `json:"picture_type"`
	Creator       struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Tracklist string `json:"tracklist"`
		Type      string `json:"type"`
	} `json:"creator"`
	Type   string `json:"type"`
	Tracks struct {
		Data []struct {
			ID                    int    `json:"id"`
			Readable              bool   `json:"readable"`
			Title                 string `json:"title"`
			TitleShort            string `json:"title_short"`
			TitleVersion          string `json:"title_version,omitempty"`
			Isrc                  string `json:"isrc"`
			Link                  string `json:"link"`
			Duration              int    `json:"duration"`
			Rank                  int    `json:"rank"`
			ExplicitLyrics        bool   `json:"explicit_lyrics"`
			ExplicitContentLyrics int    `json:"explicit_content_lyrics"`
			ExplicitContentCover  int    `json:"explicit_content_cover"`
			Preview               string `json:"preview"`
			Md5Image              string `json:"md5_image"`
			TimeAdd               int    `json:"time_add"`
			Artist                struct {
				ID        int    `json:"id"`
				Name      string `json:"name"`
				Link      string `json:"link"`
				Tracklist string `json:"tracklist"`
				Type      string `json:"type"`
			} `json:"artist"`
			Album struct {
				ID          int    `json:"id"`
				Title       string `json:"title"`
				Upc         string `json:"upc"`
				Cover       string `json:"cover"`
				CoverSmall  string `json:"cover_small"`
				CoverMedium string `json:"cover_medium"`
				CoverBig    string `json:"cover_big"`
				CoverXl     string `json:"cover_xl"`
				Md5Image    string `json:"md5_image"`
				Tracklist   string `json:"tracklist"`
				Type        string `json:"type"`
			} `json:"album"`
			Type string `json:"type"`
		} `json:"data"`
		Checksum string `json:"checksum"`
	} `json:"tracks"`
}

type TrackInfo struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Artist  string `json:"artist"`
	Album   string `json:"album"`
	Preview string `json:"preview"`
	Lyrics  string `json:"lyrics"`
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

// Get track info from spotify
func getTrack() (Track, error) {
	url := "https://api.deezer.com/playlist/13701736741"

	// Création de la requête
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

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

func delete(str string) string {
	var frames int = len(str)
	var result string
	var in bool = true
	for i := 0; i < frames; i++ {
		if in == true {
			if str[i] == '[' {
				in = false
			} else {
				result += string(str[i])
			}
		} else if in == false {
			if str[i] == ']' {
				in = true
			}
		}

	}
	return result
}

func getInfoTrack() ([]TrackInfo, string) {
	geniusToken := "kqpwlWVEknmSRiSnXiFLtXbFW9pv0Nn92i9jWe9qywhY8jkD0W7TaHYwDxLSigYz"

	trackInfo := []TrackInfo{}
	track, nil := getTrack()

	if nil != nil {
		log.Fatal(nil)
	}

	for i := 0; i < len(track.Tracks.Data); i++ {
		lyrics, err := searchLyricsOnGenius(track.Tracks.Data[i].Title, track.Tracks.Data[i].Artist.Name, geniusToken)
		if err != nil {
			log.Fatal("Erreur :", err)
		}
		trackInfo = append(trackInfo, TrackInfo{
			ID:      track.Tracks.Data[i].ID,
			Title:   track.Tracks.Data[i].Title,
			Artist:  track.Tracks.Data[i].Artist.Name,
			Album:   track.Tracks.Data[i].Album.Title,
			Preview: track.Tracks.Data[i].Preview,
			Lyrics:  delete(lyrics),
		})
	}

	return trackInfo, ""
}

func main() {
	trackInfo, _ := getInfoTrack()

	for _, track := range trackInfo {
		fmt.Println("----------------------------------------")
		fmt.Println("Titre:", track.Title)
		fmt.Println("Artiste:", track.Artist)
		fmt.Println("Album:", track.Album)
		fmt.Println("Preview:", track.Preview)
		fmt.Println()
	}
}
