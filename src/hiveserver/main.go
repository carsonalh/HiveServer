package main

import (
	"errors"
	"flag"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

type ServerState struct {
	hostedGameState HostedGameState
}

var state *ServerState

func withHeaders(next http.Handler, restrictOrigin bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if restrictOrigin {
			w.Header().Set("Access-Control-Allow-Origin", "https://hivegame.io")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		next.ServeHTTP(w, r)
	})
}

func createServer() *http.Server {
	var useTls bool
	var restrictOrigin bool

	var development bool

	flag.BoolVar(&development, "development", false, "Set to run in development mode")
	flag.Parse()

	useTls = !development
	restrictOrigin = !development

	state = new(ServerState)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.Handle("GET /join", withHeaders(new(joinHandler), restrictOrigin))
	mux.Handle("GET /hosted-game/new", withHeaders(CreateHostedGameNewHandler(&state.hostedGameState), restrictOrigin))
	mux.Handle("GET /hosted-game/play", withHeaders(CreateHostedGamePlayHandler(&state.hostedGameState), restrictOrigin))

	mux.Handle("/", http.FileServer(http.Dir("./static/")))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if useTls {
			log.Println("Server starting (HTTPS)")
			if err := server.ListenAndServeTLS("master.crt", "hivegame.pem"); !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		} else {
			log.Println("Server starting (HTTP)")
			if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
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
