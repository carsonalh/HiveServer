package main

import (
	"HiveServer/src/hivegame"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"sync"
)

type hostedGame struct {
	blackPlayer uint64
	whitePlayer uint64
	gameId      uint64
	game        hivegame.HiveGame
}

type pendingGame struct {
	playerId uint64
	gameId   uint64
}

type serverState struct {
	// map game id (uint64) to hostedGame
	games               sync.Map
	pendingGame         *pendingGame
	notifyGameFulfilled *sync.Cond
}

var state = serverState{}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file:\n%v\n", err)
	}

	state.notifyGameFulfilled = sync.NewCond(&sync.Mutex{})

	http.Handle("GET /new-game", new(newGameHandler))
	http.Handle("POST /join-game/{id}", new(joinGameHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
