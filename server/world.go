package main

import (
	"encoding/json"
	"log"
	"time"
)

const ticksPerSecond = 20

type World struct {
	players       map[PlayerId]*Player
	tanks         map[PlayerId]*TankState
	currentInputs map[PlayerId]Input // each player's most recent intent, read every tick
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
			delete(w.currentInputs, p.Id)
			close(p.send)

		case in := <-w.inputs:
			w.applyInput(in)

		case <-ticker.C:
			w.broadcast()
			w.tick++
		}
	}
}

func (w *World) applyInput(in Input) {
	_, ok := w.tanks[in.PlayerId]
	if !ok {
		return // player already gone; ignore stray buffered input
	}

	w.currentInputs[in.PlayerId] = in

	log.Printf("[%s] input ax=%.2f ay=%.2f fire=%v", in.PlayerId, in.Ax, in.Ay, in.Fire) // TEMP: remove once step() uses inputs in Step 5
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
