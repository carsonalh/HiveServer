package main

import (
	"HiveServer/src/hivegame"
	"errors"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
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

func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func createServer() *http.Server {
	state = new(serverState)
	state.pendingGameCondition = sync.NewCond(&sync.Mutex{})

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.Handle("GET /new-game", new(newGameHandler))
	mux.Handle("POST /join-game/{id}", new(joinGameHandler))
	mux.Handle("GET /game/{id}/latest-opponent-move", new(latestOpponentMoveHandler))
	mux.Handle("POST /game/{id}/moves", new(makeMoveHandler))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: withHeaders(mux),
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	return server
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file:\n%v\n", err)
	}

	createServer()
	select {}
}
