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

	id, err := strconv.Atoi(idString)

	if err != nil || state.pendingGame == nil || state.pendingGame.gameId != uint64(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// wait for another player to join this game
	state.pendingGameCondition.L.Lock()

	if state.pendingGame != nil && state.pendingGame.playerId != playerId {
		w.WriteHeader(http.StatusUnauthorized)
		state.pendingGameCondition.L.Unlock()
		return
	}

	for state.pendingGame != nil {
		state.pendingGameCondition.Wait()
	}

	state.pendingGameCondition.L.Unlock()

	found, ok := state.games.Load(uint64(id))

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("POST /join-game/{id}: 500 game was not added to current games in state")
		return
	}

	hostedGame := found.(*hostedGame)

	var color hivegame.HiveColor
	if playerId == hostedGame.blackPlayer {
		color = hivegame.ColorBlack
	} else {
		color = hivegame.ColorWhite
	}

	response := joinGameResponse{
		Id:    uint64(id),
		Game:  hostedGame.game,
		Color: color,
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("POST /join-game/{id}: 500 could not encode response, %v", err)
		return
	}
}
