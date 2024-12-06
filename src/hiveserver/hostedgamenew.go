package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"math/rand"
	"net/http"
	"os"
)

type HostedGameNewHandler struct {
	state *HostedGameState
}

func CreateHostedGameNewHandler(hostedGameState *HostedGameState) *HostedGameNewHandler {
	return &HostedGameNewHandler{
		state: hostedGameState,
	}
}

type HostedGameNewResponse struct {
	Id string `json:"id"`
}

func (h *HostedGameNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /hosted-game/new")

	var err error
	var givenToken string

	_, err = fmt.Sscanf(r.Header.Get("Authorization"), "Bearer %s", &givenToken)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = jwt.Parse(givenToken, func(_ *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id := generateGameId()

	if _, ok := h.state.games.Load(id); ok {
		log.Println("Generated an id that is already in use; time to re-think this solution")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.state.games.Store(id, NewHostedGame())

	err = json.NewEncoder(w).Encode(HostedGameNewResponse{
		Id: id,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func generateGameId() string {
	const Length = 6

	var result [Length]byte

	for i := range Length {
		r := rand.Intn(26 + 10)

		if r < 26 {
			result[i] = byte('A') + byte(r)
		} else {
			result[i] = byte('0') + byte(r-26)
		}
	}

	return string(result[:])
}
