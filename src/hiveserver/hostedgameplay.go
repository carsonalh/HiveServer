package main

import (
	"HiveServer/src/hivegame"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const DisconnectTimeout = 45 * time.Second

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
	log.Printf("GET %s?%s\n", r.URL.Path, r.URL.RawQuery)

	var conn *websocket.Conn
	var err error

	h.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err = h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection to websocket", err)
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
	var isReconnect = false

	game.condition.L.Lock()

	if game.blackPlayer != 0 && game.whitePlayer != 0 {
		switch playerId {
		case game.blackPlayer:
			game.blackConn = conn
			game.blackLastDisconnected = nil
			isReconnect = true
		case game.whitePlayer:
			game.whiteConn = conn
			game.whiteLastDisconnected = nil
			isReconnect = true
		default:
			// player is trying to play a game that's not theirs
			game.condition.L.Unlock()
			_ = conn.Close()
			return
		}

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

		go game.WatchForDisconnect(func() {
			h.state.OnGameCompleted(gameId)
		})
	}

unlock:
	game.condition.L.Unlock()

	var oppConn *websocket.Conn
	if game.blackPlayer == playerId {
		oppConn = game.whiteConn
	} else {
		oppConn = game.blackConn
	}

	if isReconnect {
		if oppConn != nil {
			err = oppConn.WriteJSON(PlayMessage{
				Event: EventReconnect,
			})
			if err != nil {
				goto wsWriteError
			}
		}
	} else {
		err = conn.WriteJSON(PlayMessage{
			Event: EventConnect,
			Connect: &GameConnect{
				Color: playerColor,
			},
		})
		if err != nil {
			goto wsWriteError
		}
	}

	for {
		err = conn.ReadJSON(&message)
		if err != nil {
			// player has disconnected; or the server terminated the connection because
			// the game is over

			select {
			case <-game.shutdown:
				// only happens when the server ended the game because the opponent disconnected
				return
			default:
				break
			}

			if over, _ := game.hiveGame.IsOver(); !over {
				when := time.Now()
				if playerColor == hivegame.ColorBlack {
					game.blackConn = nil
					game.blackLastDisconnected = &when
				} else {
					game.whiteConn = nil
					game.whiteLastDisconnected = &when
				}

				game.onDisconnect <- playerColor

				err = oppConn.WriteJSON(PlayMessage{
					Event: EventDisconnect,
				})
				if err != nil {
					goto wsWriteError
				}
			}

			break
		}

		if message.Event == EventPlayMove {
			err = oppConn.WriteJSON(message)
			if err != nil {
				goto wsWriteError
			}

			if !game.RecordMove(message.Move) {
				// the move was illegal
				err = conn.WriteJSON(PlayMessage{
					Event: EventRejectedMove,
				})
				if err != nil {
					goto wsWriteError
				}
			} else if over, winner := game.hiveGame.IsOver(); over {
				err = conn.WriteJSON(PlayMessage{
					Event: EventGameCompleted,
					Complete: &GameComplete{
						Won: winner == playerColor,
					},
				})
				if err != nil {
					goto wsWriteError
				}

				err = oppConn.WriteJSON(PlayMessage{
					Event: EventGameCompleted,
					Complete: &GameComplete{
						Won: winner != playerColor,
					},
				})
				if err != nil {
					goto wsWriteError
				}

				_ = conn.Close()
				_ = oppConn.Close()

				h.state.OnGameCompleted(gameId)

				break
			}
		}
	}

	return

wsWriteError:
	log.Println("Error writing to websocket", err)
	_ = conn.Close()
	return
}

func (hg *HostedGame) WatchForDisconnect(onGameComplete func()) {
	for {
		color := <-hg.onDisconnect

		var whenDisconnected time.Time

		hg.disconnectMutex.Lock()
		if color == hivegame.ColorBlack {
			whenDisconnected = *hg.blackLastDisconnected
		} else {
			whenDisconnected = *hg.whiteLastDisconnected
		}
		hg.disconnectMutex.Unlock()

		toWait := DisconnectTimeout - time.Now().Sub(whenDisconnected)

		if toWait < 0 {
			continue
		}

		var reconnectFailed bool

		<-time.After(toWait)
		hg.disconnectMutex.Lock()
		var lastDisconnected *time.Time
		if color == hivegame.ColorBlack {
			lastDisconnected = hg.blackLastDisconnected
		} else {
			lastDisconnected = hg.whiteLastDisconnected
		}

		if lastDisconnected != nil {
			// player of color 'color' should lose because they disconnected
			reconnectFailed = true
		}
		hg.disconnectMutex.Unlock()

		if reconnectFailed {
			hg.condition.L.Lock()
			var err error
			if color == hivegame.ColorBlack && hg.whiteConn != nil {
				err = hg.whiteConn.WriteJSON(PlayMessage{
					Event: EventGameCompleted,
					Complete: &GameComplete{
						Won: true,
					},
				})
				hg.shutdown <- struct{}{}
				_ = hg.whiteConn.Close()
				onGameComplete()
			} else if color == hivegame.ColorWhite && hg.blackConn != nil {
				err = hg.blackConn.WriteJSON(PlayMessage{
					Event: EventGameCompleted,
					Complete: &GameComplete{
						Won: true,
					},
				})
				hg.shutdown <- struct{}{}
				_ = hg.blackConn.Close()
				onGameComplete()
			}

			if err != nil {
				log.Println("Error writing to websocket after disconnect", err)
			}

			hg.condition.L.Unlock()

			return
		}
	}
}

func (state *HostedGameState) OnGameCompleted(id string) {
	if _, ok := state.games.Load(id); !ok {
		log.Printf("OnGameCompleted() called with id %s and cannot be found in the map of current games", id)
		return
	}

	state.games.Delete(id)
}
