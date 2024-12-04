package main

import (
	"HiveServer/src/hivegame"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"
)

type PlayHandler struct {
	upgrader               websocket.Upgrader
	waitingPlayer          *WaitingPlayer
	waitingPlayerCondition *sync.Cond
}

type WaitingPlayer struct {
	id       uint64
	conn     *websocket.Conn
	swapConn chan *websocket.Conn
	game     *LiveGame
}

type LiveGame struct {
	game  hivegame.HiveGame
	mutex sync.Mutex
}

const (
	EventAuthenticate  = "AUTHENTICATE"
	EventConnect       = "CONNECT"
	EventPlayMove      = "PLAY_MOVE"
	EventRejectedMove  = "REJECT_MOVE"
	EventGameCompleted = "GAME_COMPLETED"
)

type PlayMessage struct {
	Event    string        `json:"event"`
	Move     *HiveMove     `json:"move,omitempty"`
	Connect  *GameConnect  `json:"connect,omitempty"`
	Complete *GameComplete `json:"complete,omitempty"`
	Token    *string       `json:"token,omitempty"`
}

type GameConnect struct {
	Color hivegame.HiveColor `json:"color"`
}

type GameComplete struct {
	Won bool `json:"won"`
}

const (
	MoveTypePlacement = "PLACE"
	MoveTypeMovement  = "MOVE"
)

type HiveMove struct {
	MoveType  string         `json:"moveType"`
	Placement *HivePlacement `json:"placement,omitempty"`
	Movement  *HiveMovement  `json:"movement,omitempty"`
}

type HivePlacement struct {
	PieceType hivegame.HivePieceType `json:"pieceType"`
	Position  hivegame.HexVectorInt  `json:"position"`
}

type HiveMovement struct {
	From hivegame.HexVectorInt `json:"from"`
	To   hivegame.HexVectorInt `json:"to"`
}

func NewPlayHandler() *PlayHandler {
	h := new(PlayHandler)

	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	h.waitingPlayerCondition = sync.NewCond(&sync.Mutex{})

	return h
}

func (h *PlayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	var message PlayMessage

	_ = conn.ReadJSON(&message)

	if message.Event != EventAuthenticate {
		_ = conn.Close()
		return
	}

	if message.Token == nil {
		_ = conn.Close()
		return
	}

	token, err := jwt.Parse(*message.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		_ = conn.Close()
		return
	}

	idFloat, ok := token.Claims.(jwt.MapClaims)["id"]

	if !ok {
		_ = conn.Close()
		log.Println("could not get id from token claims, token has been signed without a player id")
		return
	}

	playerId := uint64(idFloat.(float64))

	var oppConn *websocket.Conn
	var swapConn chan *websocket.Conn

	var liveGame *LiveGame

	var firstIn bool

	h.waitingPlayerCondition.L.Lock()

	if h.waitingPlayer == nil {
		firstIn = true
		liveGame = new(LiveGame)
		liveGame.game = hivegame.CreateHiveGame()

		h.waitingPlayer = &WaitingPlayer{
			id:       playerId,
			conn:     conn,
			game:     liveGame,
			swapConn: make(chan *websocket.Conn, 1),
		}

		swapConn = h.waitingPlayer.swapConn

		for h.waitingPlayer != nil {
			h.waitingPlayerCondition.Wait()
		}

		oppConn = <-swapConn
	} else {
		firstIn = false
		wp := h.waitingPlayer
		h.waitingPlayer = nil
		oppConn = wp.conn
		liveGame = wp.game
		h.waitingPlayerCondition.Signal()
		wp.swapConn <- conn
	}

	h.waitingPlayerCondition.L.Unlock()

	var color hivegame.HiveColor

	if firstIn {
		color = hivegame.ColorBlack
	} else {
		color = hivegame.ColorWhite
	}

	message.Event = EventConnect
	message.Connect = &GameConnect{
		Color: color,
	}

	_ = conn.WriteJSON(message)

	message = PlayMessage{}

	for {
		_ = conn.ReadJSON(&message)

		switch message.Event {
		case EventPlayMove:
			var legal bool

			liveGame.mutex.Lock()

			if color != liveGame.game.ColorToMove {
				legal = false
				goto checkLegal
			}

			switch message.Move.MoveType {
			case MoveTypeMovement:
				legal = liveGame.game.MoveTile(message.Move.Movement.From, message.Move.Movement.To)
			case MoveTypePlacement:
				legal = liveGame.game.PlaceTile(message.Move.Placement.Position, message.Move.Placement.PieceType)
			}

		checkLegal:
			liveGame.mutex.Unlock()

			if !legal {
				_ = conn.WriteJSON(PlayMessage{
					Event: EventRejectedMove,
				})

				continue
			}
		case EventConnect:
			fallthrough
		case EventRejectedMove:
			fallthrough
		default:
			continue
		}

		_ = oppConn.WriteJSON(message)

		liveGame.mutex.Lock()
		if over, winner := liveGame.game.IsOver(); over {
			message.Event = EventGameCompleted
			message.Complete = &GameComplete{
				Won: winner == color,
			}

			_ = conn.WriteJSON(message)
			_ = conn.Close()

			message.Complete.Won = winner != color

			_ = oppConn.WriteJSON(message)
			_ = oppConn.Close()

			liveGame.mutex.Unlock()
			break
		}
		liveGame.mutex.Unlock()
	}
}
