package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"log/slog"
	"time"
)

type Transport struct {
	ctx context.Context

	heartbeat *time.Ticker
	conn      *websocket.Conn

	ack    chan Acknowledgement
	events chan GameState
}

func NewTransport(ctx context.Context, events chan GameState) *Transport {
	return &Transport{ctx: ctx, heartbeat: time.NewTicker(45 * time.Second), ack: make(chan Acknowledgement, 5), events: events}
}

func (t *Transport) Connect() error {
	// establish the websocket connection

	ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	slog.Info("Opening websocket")

	c, _, err := websocket.Dial(ctx, "wss://firepoker-75089.firebaseio.com/.ws?v=5", nil)
	//c, _, err := websocket.Dial(ctx, "wss://s-usc1b-nss-2107.firebaseio.com/.ws?v=5&ns=firepoker-75089", nil)

	if err != nil {
		return err
	}

	t.conn = c

	slog.Info("Establishing heartbeat")
	go func() {
		for range t.heartbeat.C {
			keepalive(t)
		}
	}()

	// start listening for messages
	slog.Info("Listening for messages")
	go listen(t)

	return nil
}

func (t *Transport) Disconnect() error {
	slog.Info("Disconnecting..")

	close(t.ack)
	close(t.events)

	// Stop heartbeat ticker
	t.heartbeat.Stop()

	// close conn
	slog.Info("Closing websocket")
	return t.conn.Close(websocket.StatusNormalClosure, "")
}

func (t *Transport) Send(ref int, payload string) error {
	ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	slog.Debug(">>> " + payload)
	err := t.conn.Write(ctx, websocket.MessageText, []byte(payload))
	if err != nil {
		return err
	}

	// Block until an ack message is received
	if ref > 0 {
		for {
			slog.Debug(">>>> awaiting acknowledgement...")
			ack := <-t.ack
			if ack.Ref == uint16(ref) {
				if ack.S == "ok" {
					slog.Debug(fmt.Sprintf(">>>> message with ref %d acknowledged", ref))
					return nil
				} else {
					return Error("unknown acknowledgement status: " + ack.S)
				}
			} else {
				slog.Warn(fmt.Sprintf(">>>> unexpected acknowledgement received; expected %d, discarding: %v", ref, ack))
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
		slog.Warn(fmt.Sprintf(">>> Unable to Send keepalive: %v", err))
	}
}

func listen(t *Transport) {
	for {
		ctx, cancel := context.WithCancel(t.ctx)

		// Blocks until a message is received or the context is cancelled.
		_, bytes, err := t.conn.Read(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				slog.Error(fmt.Sprintf("<<< Unable to read from websocket: %v", err))
			}
			cancel()
			break
		}

		message, err := ParseMessage(bytes)
		if err != nil {
			slog.Error(fmt.Sprintf("<<< Error parsing message: %v: %s", err, string(bytes)))
		} else {
			switch msg := message.(type) {
			case Handshake:
				slog.Debug(fmt.Sprintf("<<< Received Handshake: %+v", message))
			case Acknowledgement:
				slog.Debug(fmt.Sprintf("<<< Received Acknowledgement: %+v", message))
				slog.Debug("<<< Sending on Acknowledgement channel")
				t.ack <- msg
			case User:
				slog.Debug(fmt.Sprintf("<<< Received User (has joined the game): %+v", message))
			case Presence:
				slog.Debug(fmt.Sprintf("<<< Received Presence (online/offline): %+v", message))
			case GameState:
				slog.Debug(fmt.Sprintf("<<< Received GameState: %+v", message))
				slog.Debug("<<< Sending on Events channel")
				t.events <- msg
			default:
				slog.Debug(fmt.Sprintf("<<< Received message: %+v", message))
			}
		}

		cancel()
	}
}
