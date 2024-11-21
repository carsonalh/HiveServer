package main

import (
	"HiveServer/src/hivegame"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"strconv"
)

type joinGameHandler struct{}

type joinGameResponse struct {
	Id   uint64            `json:"id"`
	Game hivegame.HiveGame `json:"game"`
}

func (h *joinGameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST %s", r.URL.Path)

	authorization := r.Header.Get("Authorization")
	var tokenString string
	_, err := fmt.Sscanf(authorization, "Bearer %s", &tokenString)

	if err != nil {
		log.Printf("POST /join-game/{id}: could not parse bearer token, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		log.Printf("POST /join-game/{id}: could not decode jwt, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	idString := r.PathValue("id")

	id, err := strconv.Atoi(idString)

	if err != nil || state.pendingGame == nil || state.pendingGame.gameId != uint64(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// wait for another player to join this game
	state.notifyGameFulfilled.L.Lock()

	for state.pendingGame != nil {
		state.notifyGameFulfilled.Wait()
	}

	state.notifyGameFulfilled.L.Unlock()

	found, ok := state.games.Load(uint64(id))

	hostedGame := found.(hostedGame)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("POST /join-game/{id}: 500 game was not added to current games in state")
		return
	}

	response := joinGameResponse{
		Id:   uint64(id),
		Game: hostedGame.game,
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("POST /join-game/{id}: 500 could not encode response, %v", err)
		return
	}
}
