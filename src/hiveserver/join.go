package main

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type joinHandler struct {
	previousPlayerId atomic.Uint64
}

type joinResponse struct {
	Token string `json:"token"`
}

func (h *joinHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	response := joinResponse{}
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

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error writing JSON response")
		return
	}
}
