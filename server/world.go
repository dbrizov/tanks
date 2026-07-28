package main

import (
	"encoding/json"
	"time"
)

type World struct {
	players          map[PlayerId]*Player
	tanks            map[PlayerId]*Tank
	currentInputs    map[PlayerId]Input
	projectiles      map[int]*Projectile
	nextProjectileId int
	join             chan *Player
	leave            chan *Player
	inputs           chan Input
	tick             int
}

func newWorld() *World {
	return &World{
		players:       make(map[PlayerId]*Player),
		tanks:         make(map[PlayerId]*Tank),
		currentInputs: make(map[PlayerId]Input),
		projectiles:   make(map[int]*Projectile),
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
		case player := <-w.join:
			w.players[player.Id] = player
			w.tanks[player.Id] = spawnTank(player.Id)
			w.sendJoined(player)

		case player := <-w.leave:
			delete(w.players, player.Id)
			delete(w.tanks, player.Id)
			delete(w.currentInputs, player.Id)
			w.removeProjectilesOf(player.Id)
			close(player.toClient)

		case input := <-w.inputs:
			w.applyInput(input)

		case <-ticker.C:
			w.step(1.0 / ticksPerSecond)
			w.broadcast()
			w.tick++
		}
	}
}

func (w *World) sendJoined(player *Player) {
	var data, err = json.Marshal(Message{Type: MessageJoined, Data: player.Id})
	if err != nil {
		return
	}

	player.toClient <- data
}

func (w *World) removeProjectilesOf(playerId PlayerId) {
	for id, projectile := range w.projectiles {
		if projectile.owner == playerId {
			delete(w.projectiles, id)
		}
	}
}

func (w *World) applyInput(input Input) {
	var _, ok = w.tanks[input.PlayerId]
	if !ok {
		return
	}

	w.currentInputs[input.PlayerId] = input
}

func (w *World) broadcast() {
	var snapshot = Snapshot{}
	for _, tank := range w.tanks {
		snapshot.Tanks = append(snapshot.Tanks, *tank)
	}

	for _, projectile := range w.projectiles {
		snapshot.Projectiles = append(snapshot.Projectiles, *projectile)
	}

	var data, err = json.Marshal(Message{Type: MessageSnapshot, Data: snapshot})
	if err != nil {
		return
	}

	for _, player := range w.players {
		select {
		case player.toClient <- data:
		default: // buffer full, drop this frame
		}
	}
}
