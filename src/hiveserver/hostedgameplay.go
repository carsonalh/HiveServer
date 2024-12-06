package main

import (
	"HiveServer/src/hivegame"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"os"
)

type HostedGamePlayHandler struct {
	upgrader websocket.Upgrader
	state    *HostedGameState
}

func CreateHostedGamePlayHandler(hostedGameState *HostedGameState) *HostedGamePlayHandler {
	return &HostedGamePlayHandler{
		state: hostedGameState,
	}
}

func (h *HostedGamePlayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /hosted-game/play")

	var conn *websocket.Conn
	var err error

	conn, err = h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection to websocket", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message PlayMessage

	err = conn.ReadJSON(&message)
	if err != nil {
		log.Println("Error reading json from websocket", err)
		_ = conn.Close()
		return
	}

	if message.Event != EventAuthenticate {
		_ = conn.Close()
		return
	}

	givenToken := message.Token
	if givenToken == nil {
		_ = conn.Close()
		return
	}

	var token *jwt.Token

	token, err = jwt.Parse(*givenToken, func(_ *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		_ = conn.Close()
		return
	}

	playerId := uint64(token.Claims.(jwt.MapClaims)["id"].(float64))
	gameId := r.URL.Query().Get("id")

	var got any
	var game *HostedGame
	var ok bool

	got, ok = h.state.games.Load(gameId)
	if !ok {
		_ = conn.Close()
		return
	}

	game, ok = got.(*HostedGame)
	if !ok {
		log.Println("500 Could not cast games map value to *HostedGame")
		_ = conn.Close()
		return
	}

	var playerColor hivegame.HiveColor

	game.condition.L.Lock()

	if game.blackPlayer != 0 && game.whitePlayer != 0 {
		if playerId != game.blackPlayer && playerId != game.whitePlayer {
			// player is trying to play a game that's not theirs
			game.condition.L.Unlock()
			_ = conn.Close()
			return
		}

		// player is reconnecting to a disconnected game (or a token has been leaked)

		goto unlock
	}

	if game.blackPlayer == 0 && game.whitePlayer == 0 {
		// player is the first to join this game
		if rand.Intn(2) == 0 {
			game.blackPlayer = playerId
			game.blackConn = conn
			playerColor = hivegame.ColorBlack
		} else {
			game.whitePlayer = playerId
			game.whiteConn = conn
			playerColor = hivegame.ColorWhite
		}
	} else if game.blackPlayer == 0 {
		// other player already joined as white
		game.blackPlayer = playerId
		game.blackConn = conn
		playerColor = hivegame.ColorBlack
		game.condition.Signal()
		goto unlock
	} else if game.whitePlayer == 0 {
		// other player already joined as black
		game.whitePlayer = playerId
		game.whiteConn = conn
		playerColor = hivegame.ColorWhite
		game.condition.Signal()
		goto unlock
	}

	for game.blackPlayer == 0 || game.whitePlayer == 0 {
		// we were the first to join and are waiting for our opponent to join
		game.condition.Wait()
	}

unlock:
	game.condition.L.Unlock()

	var oppConn *websocket.Conn
	if game.blackPlayer == playerId {
		oppConn = game.whiteConn
	} else {
		oppConn = game.blackConn
	}

	err = conn.WriteJSON(PlayMessage{
		Event: EventConnect,
		Connect: &GameConnect{
			Color: playerColor,
		},
	})
	if err != nil {
		log.Println("Error writing json to websocket", err)
		_ = conn.Close()
		return
	}

	for {
		err = conn.ReadJSON(&message)
		if err != nil {
			// player has disconnected, handle this somehow
		}

		if message.Event == EventPlayMove {
			_ = oppConn.WriteJSON(message)

			if !game.RecordMove(message.Move) {
				// the move was illegal
				_ = conn.WriteJSON(PlayMessage{
					Event: EventRejectedMove,
				})
			} else if over, winner := game.hiveGame.IsOver(); over {
				_ = conn.WriteJSON(PlayMessage{
					Event: EventGameCompleted,
					Complete: &GameComplete{
						Won: winner == playerColor,
					},
				})
			}
		}
	}
}
