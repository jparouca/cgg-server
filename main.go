package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients sync.Map

func main() {
	go updateActiveConnections()

	http.HandleFunc("/ws", wsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	clients.Store(conn, true)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
	}

	clients.Delete(conn)
}

func updateActiveConnections() {
	for {
		activeConnections := make([]string, 0)

		clients.Range(func(key, value interface{}) bool {
			// key is the connection, value is not used
			activeConnections = append(activeConnections, "Active Connection")
			return true
		})

		message, err := json.Marshal(map[string]interface{}{
			"activeConnections": activeConnections,
		})
		if err != nil {
			log.Println(err)
			return
		}

		clients.Range(func(key, value interface{}) bool {
			conn := key.(*websocket.Conn)
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println(err)
				return false
			}
			return true
		})

		time.Sleep(1 * time.Second)
	}
}
