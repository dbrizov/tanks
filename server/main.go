package main

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
)

func main() {
	var world = newWorld()
	go world.run()

	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		handlePlay(world, w, r)
	})

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlePlay(world *World, w http.ResponseWriter, r *http.Request) {
	var id, err = authenticate(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})

	if err != nil {
		log.Printf("accept: %v", err)
		return
	}

	defer conn.CloseNow()

	var player = newPlayer(id, conn)
	log.Printf("[%s] connected", player.Id)
	defer log.Printf("[%s] disconnected", player.Id)

	var ctx = r.Context()
	go player.writeLoop(ctx)

	world.join <- player
	defer func() { world.leave <- player }()

	player.readLoop(ctx, world)
}
