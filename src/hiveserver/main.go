package main

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"sync"
)

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

	mux.Handle("GET /join", new(joinHandler))
	mux.Handle("GET /play", NewPlayHandler())

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
