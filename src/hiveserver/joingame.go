package main

import (
	"HiveServer/src/hivegame"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type joinGameHandler struct{}

type joinGameResponse struct {
	Id    uint64             `json:"id"`
	Game  hivegame.HiveGame  `json:"game"`
	Color hivegame.HiveColor `json:"color"`
}

func (h *joinGameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST %s", r.URL.Path)

	playerId, ok := loadPlayerId(w, r)

	if !ok {
		return
	}

	idString := r.PathValue("id")

	idInt, err := strconv.Atoi(idString)
	id := uint64(idInt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// first check if the game already exists
	found, gameExists := state.games.Load(id)
	var currentGame *hostedGame

	if gameExists {
		currentGame, ok = found.(*hostedGame)

		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Game of id %d is not a *hostedGame type, this is a bug", id)
			return
		}

		var playerColor = new(hivegame.HiveColor)

		switch playerId {
		case currentGame.blackPlayer:
			*playerColor = hivegame.ColorBlack
		case currentGame.whitePlayer:
			*playerColor = hivegame.ColorWhite
		default:
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Player hacked into a game that was not theirs, player = %d, game = %d", playerId, id)
			return
		}

		_ = json.NewEncoder(w).Encode(joinGameResponse{
			Id:    id,
			Color: *playerColor,
			Game:  currentGame.game,
		})
		return
	}

	// the game does not already exist;
	// wait for another player to join this game
	state.pendingGameCondition.L.Lock()

	if state.pendingGame != nil && state.pendingGame.playerId != playerId {
		w.WriteHeader(http.StatusNotFound)
		state.pendingGameCondition.L.Unlock()
		return
	}

	for state.pendingGame != nil {
		state.pendingGameCondition.Wait()
	}

	state.pendingGameCondition.L.Unlock()

	found, ok = state.games.Load(id)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("POST /join-game/{id}: 500 game was not added to current games in state")
		return
	}

	currentGame = found.(*hostedGame)

	var color hivegame.HiveColor
	if playerId == currentGame.blackPlayer {
		color = hivegame.ColorBlack
	} else {
		color = hivegame.ColorWhite
	}

	response := joinGameResponse{
		Id:    id,
		Game:  currentGame.game,
		Color: color,
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("POST /join-game/{id}: 500 could not encode response, %v", err)
		return
	}
}
