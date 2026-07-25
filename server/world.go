package main

import (
	"encoding/json"
	"time"
)

const ticksPerSecond = 20

type World struct {
	players       map[PlayerId]*Player
	tanks         map[PlayerId]*TankState
	currentInputs map[PlayerId]Input
	join          chan *Player
	leave         chan *Player
	inputs        chan Input
	tick          int
}

func newWorld() *World {
	return &World{
		players:       make(map[PlayerId]*Player),
		tanks:         make(map[PlayerId]*TankState),
		currentInputs: make(map[PlayerId]Input),
		join:          make(chan *Player),
		leave:         make(chan *Player),
		inputs:        make(chan Input, 64),
	}
}

func (w *World) run() {
	var ticker = time.NewTicker(time.Second / ticksPerSecond)
	defer ticker.Stop()

	for {
		select {
		case p := <-w.join:
			w.players[p.Id] = p
			w.tanks[p.Id] = spawnTank(p.Id)

		case p := <-w.leave:
			delete(w.players, p.Id)
			delete(w.tanks, p.Id)
			delete(w.currentInputs, p.Id)
			close(p.send)

		case in := <-w.inputs:
			w.applyInput(in)

		case <-ticker.C:
			w.step(1.0 / ticksPerSecond)
			w.broadcast()
			w.tick++
		}
	}
}

func (w *World) applyInput(in Input) {
	var _, ok = w.tanks[in.PlayerId]
	if !ok {
		return // player already gone; ignore stray buffered input
	}

	w.currentInputs[in.PlayerId] = in
}

func (w *World) broadcast() {
	var snap = Snapshot{Tick: w.tick}
	for _, t := range w.tanks {
		snap.Tanks = append(snap.Tanks, *t)
	}

	var data, err = json.Marshal(snap)
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
