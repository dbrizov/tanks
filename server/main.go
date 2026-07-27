package main

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
)

func main() {
	var world = newWorld()
	go world.run()

	http.HandleFunc("/play", func(writer http.ResponseWriter, request *http.Request) {
		handlePlay(world, writer, request)
	})

	http.HandleFunc("/health", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlePlay(world *World, writer http.ResponseWriter, request *http.Request) {
	var id, err = authenticate(request)
	if err != nil {
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(writer, request, &websocket.AcceptOptions{
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

	var ctx = request.Context()
	go player.writeLoop(ctx)

	world.join <- player
	defer func() { world.leave <- player }()

	player.readLoop(ctx, world)
}
