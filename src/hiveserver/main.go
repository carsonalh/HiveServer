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
	// We invent the concept of a `tick`, which is a kind of atomic move.
	// Sometimes turns are skipped etc. so we want to know the last thing a player has 'seen' for
	// synchronisation purposes later
	blackLastSeenTick uint
	whitePlayer       uint64
	whiteLastSeenTick uint
	gameId            uint64
	nextMove          *sync.Cond
	game              hivegame.HiveGame
}

func toTick(moveNumber, playerToMove uint) uint {
	return 2*moveNumber + playerToMove
}

type pendingGame struct {
	playerId uint64
	gameId   uint64
}

type serverState struct {
	// map game id (uint64) to hostedGame
	games                sync.Map
	pendingGame          *pendingGame
	pendingGameCondition *sync.Cond
}

var state *serverState

func main() {
	state = new(serverState)
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file:\n%v\n", err)
	}

	state.pendingGameCondition = sync.NewCond(&sync.Mutex{})

	http.Handle("GET /new-game", new(newGameHandler))
	http.Handle("POST /join-game/{id}", new(joinGameHandler))
	http.Handle("GET /game/{id}/latest-opponent-move", new(latestOpponentMoveHandler))
	http.Handle("POST /game/{id}/moves", new(makeMoveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
