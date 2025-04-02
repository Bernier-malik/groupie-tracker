package models

import "github.com/gorilla/websocket"

type Player struct {
	Conn *websocket.Conn
	Name string
}

type Lobby struct {
	Code    string
	Players []*Player
}

var Lobbies = make(map[string]*Lobby)
