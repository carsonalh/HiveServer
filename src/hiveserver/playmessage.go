package main

import "HiveServer/src/hivegame"

const (
	EventAuthenticate = "AUTHENTICATE"
	EventConnect      = "CONNECT"
	// EventDisconnect should be sent when an opponent disconnects from the game
	EventDisconnect = "DISCONNECT"
	// EventReconnect should be sent when an opponent reconnects to a disconnected game
	EventReconnect     = "RECONNECT"
	EventPlayMove      = "PLAY_MOVE"
	EventRejectedMove  = "REJECT_MOVE"
	EventGameCompleted = "GAME_COMPLETED"
)

type PlayMessage struct {
	Event    string        `json:"event"`
	Move     *HiveMove     `json:"move,omitempty"`
	Connect  *GameConnect  `json:"connect,omitempty"`
	Complete *GameComplete `json:"complete,omitempty"`
	Token    *string       `json:"token,omitempty"`
}

type GameConnect struct {
	Color hivegame.HiveColor `json:"color"`
}

type GameComplete struct {
	Won bool `json:"won"`
}

const (
	MoveTypePlacement = "PLACE"
	MoveTypeMovement  = "MOVE"
)

type HiveMove struct {
	MoveType  string         `json:"moveType"`
	Placement *HivePlacement `json:"placement,omitempty"`
	Movement  *HiveMovement  `json:"movement,omitempty"`
}

type HivePlacement struct {
	PieceType hivegame.HivePieceType `json:"pieceType"`
	Position  hivegame.HexVectorInt  `json:"position"`
}

type HiveMovement struct {
	From hivegame.HexVectorInt `json:"from"`
	To   hivegame.HexVectorInt `json:"to"`
}
