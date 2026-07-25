package main

import (
	"encoding/json"
	"time"
)

const ticksPerSecond = 20

type TankState struct {
	Id    PlayerId `json:"id"`
	X     float64  `json:"x"`
	Y     float64  `json:"y"`
	Angle float64  `json:"angle"`
}

type Snapshot struct {
	Tick  int         `json:"tick"`
	Tanks []TankState `json:"tanks"`
}

type World struct {
	players map[PlayerId]*Player
	tanks   map[PlayerId]*TankState
	join    chan *Player
	leave   chan *Player
	tick    int
}

func newWorld() *World {
	return &World{
		players: make(map[PlayerId]*Player),
		tanks:   make(map[PlayerId]*TankState),
		join:    make(chan *Player),
		leave:   make(chan *Player),
	}
}

func (w *World) run() {
	ticker := time.NewTicker(time.Second / ticksPerSecond)
	defer ticker.Stop()

	for {
		select {
		case p := <-w.join:
			w.players[p.Id] = p
			w.tanks[p.Id] = &TankState{Id: p.Id, X: 400, Y: 300}

		case p := <-w.leave:
			delete(w.players, p.Id)
			delete(w.tanks, p.Id)
			close(p.send)

		case <-ticker.C:
			w.broadcast()
			w.tick++
		}
	}
}

func (w *World) broadcast() {
	snap := Snapshot{Tick: w.tick}
	for _, t := range w.tanks {
		snap.Tanks = append(snap.Tanks, *t)
	}

	data, err := json.Marshal(snap)
	if err != nil {
		return
	}

	for _, p := range w.players {
		select {
		case p.send <- data:
		default: // buffer full → drop this frame
		}
	}
}
