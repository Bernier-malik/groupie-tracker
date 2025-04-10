package blindtest

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Track struct {
	Title   string
	Artist  string
	Album   string
	Preview string
}

func GetTracks() ([]Track, error) {
	resp, err := http.Get("https://api.deezer.com/playlist/13701736741")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	tracks := data["tracks"].(map[string]interface{})["data"].([]interface{})
	var trackList []Track

	for _, t := range tracks {
		track := t.(map[string]interface{})
		trackList = append(trackList, Track{
			Title:   track["title"].(string),
			Artist:  track["artist"].(map[string]interface{})["name"].(string),
			Album:   track["album"].(map[string]interface{})["title"].(string),
			Preview: track["preview"].(string),
		})
	}
	return trackList, nil
}

func GetRandomTrack() (Track, error) {
	tracks, err := GetTracks()
	if err != nil {
		return Track{}, err
	}
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(tracks))
	return tracks[randomIndex], nil
}

func CheckAnswer(userInput, correctTitle string) bool {
	return normalize(userInput) == normalize(correctTitle)
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "'", "")
	return s
}
