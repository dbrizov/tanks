package main

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
)

type PlayerId string

type Player struct {
	Id   PlayerId
	conn *websocket.Conn
	send chan []byte
}

func newPlayer(id PlayerId, conn *websocket.Conn) *Player {
	return &Player{
		Id:   id,
		conn: conn,
		send: make(chan []byte, 16),
	}
}

func (p *Player) writeLoop(ctx context.Context) {
	for data := range p.send {
		err := p.conn.Write(ctx, websocket.MessageText, data)
		if err != nil {
			return
		}
	}
}

func (p *Player) readLoop(ctx context.Context, world *World) {
	for {
		_, data, err := p.conn.Read(ctx)
		if err != nil {
			return // client gone
		}

		var in Input
		if err := json.Unmarshal(data, &in); err != nil {
			continue // ignore malformed messages, keep the connection alive
		}
		in.PlayerId = p.Id // identity comes from the connection, not the client

		world.inputs <- in
	}
}
