package main

import (
	"HiveServer/src/hivegame"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type latestOpponentMoveHandler struct{}

type latestOpponentMoveResponse struct {
	Game *hivegame.HiveGame `json:"game"`
}

func (h *latestOpponentMoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	playerId, ok := loadPlayerId(w, r)

	if !ok {
		return
	}

	gameIdString := r.PathValue("id")

	gameId, err := strconv.Atoi(gameIdString)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	value, gameExists := state.games.Load(uint64(gameId))

	hostedGame, casted := value.(*hostedGame)

	if !casted || !gameExists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if hostedGame.blackPlayer != playerId && hostedGame.whitePlayer != playerId {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var playerColor hivegame.HiveColor

	if playerId == hostedGame.blackPlayer {
		playerColor = hivegame.ColorBlack
	} else {
		playerColor = hivegame.ColorWhite
	}

	var playerLastSeenTick *uint = nil

	if playerColor == hivegame.ColorBlack {
		playerLastSeenTick = &hostedGame.blackLastSeenTick
	} else {
		playerLastSeenTick = &hostedGame.whiteLastSeenTick
	}

	// wait for the next move to happen, if it has not already
	hostedGame.nextMove.L.Lock()

	latestTick := toTick(uint(hostedGame.game.Move), uint(hostedGame.game.ColorToMove))

	if *playerLastSeenTick == latestTick {
		nextTick := latestTick
		for latestTick == nextTick {
			hostedGame.nextMove.Wait()
			nextTick = toTick(uint(hostedGame.game.Move), uint(hostedGame.game.ColorToMove))
		}
	} else if *playerLastSeenTick > latestTick {
		w.WriteHeader(http.StatusInternalServerError)
		hostedGame.nextMove.L.Unlock()
		return
	}

	*playerLastSeenTick = latestTick
	hostedGame.nextMove.L.Unlock()

	err = json.NewEncoder(w).Encode(latestOpponentMoveResponse{
		Game: &hostedGame.game,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s %s, 500 could not encode JSON, %v", r.Method, r.URL.Path, err)
		return
	}
}
