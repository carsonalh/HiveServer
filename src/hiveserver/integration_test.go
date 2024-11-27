package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"testing"
)

// launches the server in a blank state
func launchServer() func() {
	if err := godotenv.Load("integration_test.env"); err != nil {
		fmt.Println("Could not load .env file")
		return nil
	}

	server := createServer()

	return func() {
		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Println(err)
		}

		fmt.Println("Server successfully shut down")
	}
}

func TestMain(m *testing.M) {
	cleanup := launchServer()
	_ = m.Run()
	cleanup()
	os.Exit(0)
}

func TestSimpleRequest(t *testing.T) {
	_, err := http.Get("http://localhost:8080")

	if err != nil {
		t.Errorf("Could not make a GET request: %v", err)
	}
}

func TestTwoPlayersJoinAGame(t *testing.T) {
	response, err := http.Get("http://localhost:8080/new-game")

	if err != nil {
		t.Errorf("Player 1 could not make a GET request: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Player 1 returned unexpected status code: %d", response.StatusCode)
	}

	decoded := newGameResponse{}

	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatal(err)
	}

	token, err := jwt.NewParser().Parse(decoded.Token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if !token.Valid {
		t.Fatal("Invalid jwt")
	}

	player1IdFloat, ok := token.Claims.(jwt.MapClaims)["id"].(float64)
	player1Id := uint64(player1IdFloat)

	if !ok {
		t.Fatal("Failed to cast player 1 id")
	}

	player1GameId := decoded.Id

	fmt.Printf("Player 1 id = %d, game id = %d\n", player1Id, player1GameId)

	go func() {
		request, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8080/join-game/%d", player1GameId), nil)
		if err != nil {
			t.Error(err)
			return
		}
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token.Raw))
		response, err = http.DefaultClient.Do(request)

		if err != nil {
			t.Error(err)
			return
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Failed to join the game (player 1), %d", response.StatusCode)
			return
		}
	}()

	response, err = http.Get("http://localhost:8080/new-game")

	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Failed to join the game (player 2), %d", response.StatusCode)
	}

	decoded = newGameResponse{}
	err = json.NewDecoder(response.Body).Decode(&decoded)

	if err != nil {
		t.Fatal(err)
	}

	token, err = jwt.NewParser().Parse(decoded.Token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		t.Fatal(err)
	}

	player2IdFloat, ok := token.Claims.(jwt.MapClaims)["id"].(float64)
	player2Id := uint64(player2IdFloat)

	if !ok {
		t.Errorf("Failed to cast player 2 id")
	} else if player1Id == player2Id {
		t.Errorf("Player 1 and player 2 should be different, both are %d", player1Id)
	}
}

func TestManyPlayersJoinAtOnce(t *testing.T) {
	type playerInfo struct {
		gameId   uint64
		playerId uint64
	}

	// post the empty struct on error
	joinedGames := make(chan playerInfo, 64)

	const NumPlayers = 100

	joinGame := func() {
		response, err := http.Get("http://localhost:8080/new-game")

		if err != nil {
			t.Error(err)
			joinedGames <- playerInfo{}
			return
		}

		decoded := newGameResponse{}

		err = json.NewDecoder(response.Body).Decode(&decoded)

		if err != nil {
			t.Error(err)
			joinedGames <- playerInfo{}
			return
		}

		token, err := jwt.Parse(decoded.Token, func(_ *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			t.Error(err)
			joinedGames <- playerInfo{}
			return
		}

		playerIdFloat := token.Claims.(jwt.MapClaims)["id"].(float64)
		playerId := uint64(playerIdFloat)

		if decoded.Game == nil {
			request, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8080/join-game/%d", decoded.Id), nil)
			if err != nil {
				t.Error(err)
				joinedGames <- playerInfo{}
				return
			}

			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", decoded.Token))

			response, err = http.DefaultClient.Do(request)

			if err != nil {
				t.Error(err)
				joinedGames <- playerInfo{}
				return
			}

			if response.StatusCode != http.StatusOK {
				t.Error("Failed to join the game")
				joinedGames <- playerInfo{}
				return
			}

			joinedGameResponse := joinGameResponse{}
			err = json.NewDecoder(response.Body).Decode(&joinedGameResponse)

			if err != nil {
				t.Error(err)
				joinedGames <- playerInfo{}
				return
			}
		}

		joinedGames <- playerInfo{decoded.Id, playerId}
	}

	for range NumPlayers {
		go joinGame()
	}

	received := make([]playerInfo, NumPlayers)

	for i := range NumPlayers {
		received[i] = <-joinedGames
	}

	badValues := 0

	for _, r := range received {
		if r.playerId == 0 || r.gameId == 0 {
			badValues++
		}
	}

	if badValues > 0 {
		t.Fatalf("Received %d non-responses from the server", badValues)
	}

	gameCounts := make(map[uint64]int)

	for _, r := range received {
		if _, ok := gameCounts[r.gameId]; !ok {
			gameCounts[r.gameId] = 1
		} else {
			gameCounts[r.gameId]++
		}
	}

	for id, count := range gameCounts {
		if count != 2 {
			t.Errorf("Game id %d had %d player(s) receive its id", id, count)
		}
	}

	playerIds := make(map[uint64]struct{})

	for _, r := range received {
		if _, ok := playerIds[r.playerId]; ok {
			t.Errorf("Server handed out player id %d at least twice", r.playerId)
		} else {
			playerIds[r.playerId] = struct{}{}
		}
	}
}
