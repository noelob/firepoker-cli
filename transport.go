package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"time"
)

type Transport struct {
	ctx context.Context

	heartbeat *time.Ticker
	conn      *websocket.Conn

	ack chan Acknowledgement
}

func NewTransport(ctx context.Context) *Transport {
	return &Transport{ctx: ctx, heartbeat: time.NewTicker(45 * time.Second), ack: make(chan Acknowledgement, 5)}
}

func (t *Transport) Connect() error {
	// establish the websocket connection

	ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	fmt.Println("Opening websocket")

	c, _, err := websocket.Dial(ctx, "wss://firepoker-75089.firebaseio.com/.ws?v=5", nil)
	//c, _, err := websocket.Dial(ctx, "wss://s-usc1b-nss-2107.firebaseio.com/.ws?v=5&ns=firepoker-75089", nil)

	if err != nil {
		return err
	}

	t.conn = c

	fmt.Println("Establishing heartbeat")
	go func() {
		for range t.heartbeat.C {
			keepalive(t)
		}
	}()

	// start listening for messages
	fmt.Println("Listening for messages")
	go listen(t)

	return nil
}

func (t *Transport) Disconnect() error {
	fmt.Println("Disconnecting..")

	close(t.ack)

	// Stop heartbeat ticker
	t.heartbeat.Stop()

	// close conn
	fmt.Println("Closing websocket")
	return t.conn.Close(websocket.StatusNormalClosure, "")
}

func (t *Transport) Send(ref int, payload string) error {
	ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	println(">>> " + payload)
	err := t.conn.Write(ctx, websocket.MessageText, []byte(payload))
	if err != nil {
		return err
	}

	// Block until an ack message is received
	if ref > 0 {
		for {
			fmt.Printf(">>>> awaiting acknowledgement...\n")
			ack := <-t.ack
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

func (t *Transport) IsConnected() bool {
	return t.conn != nil
}

func keepalive(t *Transport) {
	err := t.Send(0, "0")
	if err != nil {
		msg := fmt.Sprintf(">>> Unable to Send keepalive: %v", err)
		fmt.Printf(msg)
	}
}

func listen(t *Transport) {
	for {
		ctx, cancel := context.WithCancel(t.ctx)

		// Blocks until a message is received or the context is cancelled.
		_, bytes, err := t.conn.Read(ctx)
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
				t.ack <- msg
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
