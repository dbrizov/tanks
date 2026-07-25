package main

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
)

type PlayerId string

type Player struct {
	Id       PlayerId
	conn     *websocket.Conn
	toClient chan []byte
}

func newPlayer(id PlayerId, conn *websocket.Conn) *Player {
	return &Player{
		Id:       id,
		conn:     conn,
		toClient: make(chan []byte, 16),
	}
}

func (p *Player) writeLoop(ctx context.Context) {
	for data := range p.toClient {
		var err = p.conn.Write(ctx, websocket.MessageText, data)
		if err != nil {
			return
		}
	}
}

func (p *Player) readLoop(ctx context.Context, world *World) {
	for {
		var _, data, err = p.conn.Read(ctx)
		if err != nil {
			return
		}

		var input Input
		if err := json.Unmarshal(data, &input); err != nil {
			continue
		}

		input.PlayerId = p.Id

		world.inputs <- input
	}
}
