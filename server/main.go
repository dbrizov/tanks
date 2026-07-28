package main

import (
	"log"
	"net/http"
	"os"

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

	var addr = "127.0.0.1:" + portOrDefault()
	log.Printf("listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func portOrDefault() string {
	var port = os.Getenv("PORT")
	if port != "" {
		return port
	}

	return "8101"
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
