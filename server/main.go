package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func main() {
	http.HandleFunc("/play", handlePlay)

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlePlay(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})

	if err != nil {
		log.Printf("accept: %v", err)
	}

	defer conn.CloseNow()

	log.Println("client conneted")

	for {
		ctx, cancel := context.WithTimeout(r.Context(), time.Minute)

		var msg map[string]any
		err := wsjson.Read(ctx, conn, &msg)
		cancel()

		if err != nil {
			log.Printf("client gone: %v", err)
		}

		log.Printf("received: %v", msg)

		if err := wsjson.Write(r.Context(), conn, msg); err != nil {
			log.Printf("write: %v", err)
			return
		}
	}
}
