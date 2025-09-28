package main

import (
	"context"
	"fmt"
	"log/slog"
)

type Game struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	transport *Transport
	events    chan GameState

	id    string
	state GameState
}

func NewGame() *Game {
	ctx, ctxCancel := context.WithCancel(context.Background())
	events := make(chan GameState)
	return &Game{ctx: ctx, ctxCancel: ctxCancel, events: events, transport: NewTransport(ctx, events)}
}

func (g *Game) Join(id string) error {
	g.id = id

	if !g.transport.IsConnected() {
		err := g.transport.Connect()
		if err != nil {
			return err
		}
	}

	go update(g)

	// join game
	g.transport.Send(1, fmt.Sprintf(`{"t":"d","d":{"r":1,"a":"l","b":{"p":"/games/%s","h":""}}}`, id))
	g.transport.Send(2, fmt.Sprintf(`{"t":"d","d":{"r":2,"a":"p","b":{"p":"/games/%s/participants/fe24478e-0161-0c97-18ef-ab569207ac44","d":{"fullname":"go-cli","id":"fe24478e-0161-0c97-18ef-ab569207ac44"}}}}`, id))
	g.transport.Send(3, fmt.Sprintf(`{"t":"d","d":{"r":3,"a":"o","b":{"p":"/games/%s/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":{".sv":"timestamp"}}}}`, id))
	g.transport.Send(4, fmt.Sprintf(`{"t":"d","d":{"r":4,"a":"p","b":{"p":"/games/%s/participants/fe24478e-0161-0c97-18ef-ab569207ac44/online","d":true}}}`, id))

	return nil
}

func (g *Game) Leave() error {
	slog.Info("Leaving the game")

	// This cancels all derived contexts (i.e. sending and receiving messages)
	g.ctxCancel()

	// close conn
	return g.transport.Disconnect()
}

func update(g *Game) {
	slog.Debug("Waiting for game updates")
	for event := range g.events {
		slog.Info("Game update received, updating...")
		g.state = event
	}
	slog.Debug("Update channel closed")
}
