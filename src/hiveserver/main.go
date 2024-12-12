package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

type ServerState struct {
	hostedGameState HostedGameState
}

var state *ServerState

func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("CORS_ORIGIN"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}

func createServer() *http.Server {
	var useTls bool
	var development bool

	flag.BoolVar(&development, "development", false, "Set to run in development mode")
	flag.Parse()

	useTls = !development

	if os.Getenv("CORS_ORIGIN") == "" {
		panic("Cannot start the server without the CORS_ORIGIN environment variable")
	} else {
		fmt.Printf("Using origin %s\n", os.Getenv("CORS_ORIGIN"))
	}

	state = new(ServerState)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.Handle("GET /join", withHeaders(new(joinHandler)))
	mux.Handle("GET /hosted-game/new", withHeaders(CreateHostedGameNewHandler(&state.hostedGameState)))
	mux.Handle("GET /hosted-game/play", withHeaders(CreateHostedGamePlayHandler(&state.hostedGameState)))
	mux.HandleFunc("OPTIONS /hosted-game/new", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("OPTIONS %s", r.URL.Path)
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("CORS_ORIGIN"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
	})

	mux.Handle("/", http.FileServer(SpaFileServer(http.Dir("./static/"))))

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
