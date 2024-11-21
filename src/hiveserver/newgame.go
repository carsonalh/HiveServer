package main

import (
	"HiveServer/src/hivegame"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync/atomic"
)

type newGameHandler struct {
	previousGameId   atomic.Uint64
	previousPlayerId atomic.Uint64
}

type newGameResponse struct {
	Id      uint64             `json:"id"`
	Token   string             `json:"token"`
	Game    *hivegame.HiveGame `json:"game,omitempty"`
	Pending bool               `json:"pending"`
}

func (h *newGameHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	response := newGameResponse{}
	playerId := h.previousPlayerId.Add(1)

	log.Printf("GET /new-game, created player id = %d\n", playerId)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": playerId})
	secretString := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secretString))

	if err != nil {
		log.Printf("Error signing token: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Token = tokenString

	if state.pendingGame == nil {
		// we are the only ones who set state.pendingGame, so this is safe

		// we are player 1

		response.Id = h.previousGameId.Add(1)
		state.pendingGame = &pendingGame{
			gameId:   response.Id,
			playerId: playerId,
		}
		response.Pending = true
	} else {
		// here, we are player 2
		// either player 1 forgot to join or has joined and is blocking, either way, they shouldn't bother us here

		game := createHostedGame(playerId, state.pendingGame.playerId)
		response.Id = state.pendingGame.gameId
		response.Game = &game.game
		state.games.Store(response.Id, game)
		state.pendingGame = nil
		state.notifyGameFulfilled.Signal()
		response.Pending = false
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /new-game; 500: failed to encode json response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

func createHostedGame(player1, player2 uint64) hostedGame {
	var game hostedGame

	if rand.Intn(2) == 0 {
		game.blackPlayer = player1
		game.whitePlayer = player2
	} else {
		game.blackPlayer = player2
		game.whitePlayer = player1
	}

	game.game = hivegame.CreateHiveGame()

	return game
}
