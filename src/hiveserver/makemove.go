package main

import (
	"HiveServer/src/hivegame"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type makeMoveHandler struct {
}

type moveType = string

const (
	moveTypePlace = "PLACE"
	moveTypeMove  = "MOVE"
)

type makeMoveRequest struct {
	MoveType  moveType `json:"moveType"`
	Placement struct {
		PieceType hivegame.HivePieceType `json:"pieceType"`
		Position  hivegame.HexVectorInt  `json:"position"`
	} `json:"placement,omitempty"`
	Movement struct {
		From hivegame.HexVectorInt `json:"from"`
		To   hivegame.HexVectorInt `json:"to"`
	} `json:"movement,omitempty"`
}

type makeMoveResponse struct {
	Game *hivegame.HiveGame `json:"game"`
}

func (h *makeMoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	playerId, ok := loadPlayerId(w, r)

	if !ok {
		return
	}

	gameIdString := r.PathValue("id")

	gameIdInt, err := strconv.Atoi(gameIdString)

	if err != nil {
		//log.Printf("%s %s, 404 error parsing game id, %v", r.Method, r.URL.Path, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	gameId := uint64(gameIdInt)

	value, inMap := state.games.Load(gameId)

	if !inMap {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hostedGame, ok := value.(*hostedGame)

	if !ok || hostedGame.gameId != gameId {
		log.Printf("%s %s, 500 invalid state in hosted game map", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if playerId != hostedGame.blackPlayer && playerId != hostedGame.whitePlayer {
		log.Printf("%s %s, 403 player %d trying to access another game, %v", r.Method, r.URL.Path, playerId, err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var playerColor hivegame.HiveColor

	if hostedGame.blackPlayer == playerId {
		playerColor = hivegame.ColorBlack
	} else {
		playerColor = hivegame.ColorWhite
	}

	game := &hostedGame.game

	if game.ColorToMove != playerColor {
		w.WriteHeader(http.StatusForbidden)
		_, err = fmt.Fprintln(w, "Tried to move on the opponent's turn")

		if err != nil {
			log.Printf("%s %s, Error writing to socket (tried to move on opponent's turn)", r.Method, r.URL.Path)
		}

		return
	}

	request := makeMoveRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch request.MoveType {
	case moveTypeMove:
		game.MoveTile(request.Movement.From, request.Movement.To)
	case moveTypePlace:
		game.PlaceTile(request.Placement.Position, request.Placement.PieceType)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(makeMoveResponse{game})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = fmt.Fprintf(w, "%s %s, 500 could not encode json response, %v", r.Method, r.URL.Path, err)

		if err != nil {
			log.Printf("%s %s, Error writing to socket (tried to encode json response)", r.Method, r.URL.Path)
		}

		return
	}
}
