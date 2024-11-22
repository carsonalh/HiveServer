package main

import (
	"HiveServer/src/hivegame"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

type newGameHandler struct {
	previousGameId   atomic.Uint64
	previousPlayerId atomic.Uint64
}

type newGameResponse struct {
	Id      uint64              `json:"id"`
	Token   string              `json:"token"`
	Game    *hivegame.HiveGame  `json:"game,omitempty"`
	Color   *hivegame.HiveColor `json:"color,omitempty"`
	Pending bool                `json:"pending"`
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

	state.pendingGameCondition.L.Lock()
	if state.pendingGame == nil {
		// we are the only ones who set state.pendingGame, so this is safe

		// we are player 1

		response.Id = h.previousGameId.Add(1)
		response.Pending = true

		state.pendingGame = &pendingGame{
			gameId:   response.Id,
			playerId: playerId,
		}
	} else {
		// here, we are player 2
		// either player 1 forgot to join or has joined and is blocking, either way, they shouldn't bother us here

		gameId := state.pendingGame.gameId
		game := createHostedGame(playerId, state.pendingGame.playerId)
		game.gameId = gameId
		response.Id = gameId
		response.Game = &game.game
		color := new(hivegame.HiveColor)
		if game.blackPlayer == playerId {
			*color = hivegame.ColorBlack
		} else {
			*color = hivegame.ColorWhite
		}
		response.Color = color
		response.Pending = false
		state.games.Store(gameId, &game)
		state.pendingGame = nil
		state.pendingGameCondition.Signal()
	}
	state.pendingGameCondition.L.Unlock()

	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /new-game, 500 failed to encode json response")
		return
	}
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
	game.nextMove = sync.NewCond(&sync.Mutex{})

	return game
}
