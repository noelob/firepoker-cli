package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"time"
)

type deck struct {
	name   string
	values []string
}

type story string

type participant struct {
	name string
}

type Game struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	id string

	heartbeat *time.Ticker

	conn         *websocket.Conn
	name         string
	description  string
	stories      []story
	deck         deck
	participants []participant

	ack chan Acknowledgement
}

func NewGame() *Game {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return &Game{ctx: ctx, ctxCancel: ctxCancel, heartbeat: time.NewTicker(45 * time.Second), ack: make(chan Acknowledgement, 5)}
}

func (g *Game) Connect() error {
	// establish the websocket connection

	ctx, cancel := context.WithTimeout(g.ctx, 5*time.Second)
	defer cancel()

	fmt.Println("Opening websocket")

	c, _, err := websocket.Dial(ctx, "wss://firepoker-75089.firebaseio.com/.ws?v=5", nil)
	//c, _, err := websocket.Dial(ctx, "wss://s-usc1b-nss-2107.firebaseio.com/.ws?v=5&ns=firepoker-75089", nil)

	if err != nil {
		return err
	}

	g.conn = c

	fmt.Println("Establishing heartbeat")
	go func() {
		for range g.heartbeat.C {
			keepalive(g)
		}
	}()

	// start listening for messages
	fmt.Println("Listening for messages")
	go listen(g)

	return nil
}

func (g *Game) Join(id string) error {
	g.id = id

	// join game
	g.send(1, `{"t":"d","d":{"r":1,"a":"l","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d","h":""}}}`)
	g.send(2, `{"t":"d","d":{"r":2,"a":"p","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/fe24478e-0161-0c97-18ef-ab569207ac44","d":{"fullname":"go-cli","id":"fe24478e-0161-0c97-18ef-ab569207ac44"}}}}`)
	g.send(3, `{"t":"d","d":{"r":3,"a":"o","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":{".sv":"timestamp"}}}}`)
	g.send(4, `{"t":"d","d":{"r":4,"a":"p","b":{"p":"/games/d2538816-2f8e-a8b0-6534-30857b5e932d/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":true}}}`)

	return nil
}

func (g *Game) Leave() error {
	fmt.Println("Leaving the game")

	close(g.ack)

	// This cancels all derived contexts (i.e. sending and receiving messages)
	g.ctxCancel()

	// Stop heartbeat ticker
	g.heartbeat.Stop()

	// close conn
	fmt.Println("Closing websocket")
	return g.conn.Close(websocket.StatusNormalClosure, "")
}

func (g *Game) send(ref int, payload string) error {
	ctx, cancel := context.WithTimeout(g.ctx, 5*time.Second)
	defer cancel()

	println(">>> " + payload)
	err := g.conn.Write(ctx, websocket.MessageText, []byte(payload))
	if err != nil {
		return err
	}

	// Block until an ack message is received
	if ref > 0 {
		for {
			fmt.Printf(">>>> awaiting acknowledgement...\n")
			ack := <-g.ack
			if ack.Ref == uint16(ref) {
				if ack.S == "ok" {
					fmt.Printf(">>>> message with ref %d acknowledged\n", ref)
					return nil
				} else {
					return Error("unknown acknowledgement status: " + ack.S)
				}
			} else {
				fmt.Printf(">>>> unexpected acknowledgement received; expected %d, discarding: %v\n", ref, ack)
			}
		}
	}
	return nil
}

func keepalive(g *Game) {
	err := g.send(0, "0")
	if err != nil {
		msg := fmt.Sprintf(">>> Unable to send keepalive: %v", err)
		fmt.Printf(msg)
	}
}

func listen(g *Game) {
	for {
		ctx, cancel := context.WithCancel(g.ctx)

		// Blocks until a message is received or the context is cancelled.
		_, bytes, err := g.conn.Read(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				msg := fmt.Sprintf("<<< Unable to read from websocket: %v", err)
				fmt.Println(msg)
			}
			cancel()
			break
		}

		message, err := ParseMessage(bytes)
		if err != nil {
			msg := fmt.Sprintf("<<< Error parsing message: %v: %s", err, string(bytes))
			fmt.Println(msg)
		} else {
			switch msg := message.(type) {
			case Handshake:
				fmt.Printf("<<< Received Handshake: %+v\n", message)
			case Acknowledgement:
				fmt.Printf("<<< Received Acknowledgement: %+v\n", message)
				fmt.Printf("<<< Sending on Acknowledgement channel\n")
				g.ack <- msg
			case User:
				fmt.Printf("<<< Received User (has joined the game): %+v\n", message)
			case Presence:
				fmt.Printf("<<< Received Presence (online/offline): %+v\n", message)
			case GameState:
				fmt.Printf("<<< Received GameState: %+v\n", message)
			default:
				fmt.Printf("<<< Received message: %+v\n", message)
			}
		}

		cancel()
	}
}
