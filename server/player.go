package main

import (
	"context"

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

func (p *Player) readLoop(ctx context.Context) {
	for {
		_, _, err := p.conn.Read(ctx)
		if err != nil {
			return // client gone
		}
		// TODO decode the message into an Input and hand it to the world.
	}
}
