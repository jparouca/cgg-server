package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// fuck http connections me and my bois are web sockets
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ClientData struct {
	Nickname string
}

// active connections and nick
var clients sync.Map

func main() {
	go updateActiveConnections()

	http.HandleFunc("/ws", wsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// prevent connection refused, everyone is welcome, even chinese bots
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	nickname := r.URL.Query().Get("nickname")
	if nickname == "" {
		nickname = "Anonymous"
	}


// fuck http connections me and my bois are web sockets
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	clientData := &ClientData{Nickname: nickname}
	clients.Store(conn, clientData)

	//read message
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		if messageType == websocket.TextMessage {
			var data map[string]interface{}
			err = json.Unmarshal(message, &data)
			if err != nil {
				log.Println(err)
				continue
			}

			if action, ok := data["action"].(string); ok {
				if action == "ping" {
					targetNickname, ok := data["targetNickname"].(string)
					if !ok {
						log.Println("Invalid 'targetNickname' field in the message")
						continue
					}

					clients.Range(func(key, value interface{}) bool {
						clientData := value.(*ClientData)
						if clientData.Nickname == targetNickname {
							pingMessage := []byte("ping")
							conn := key.(*websocket.Conn)
							conn.WriteMessage(websocket.TextMessage, pingMessage)
						}
						return true
					})
				}
			}
		}
	}

	// Remove the client from the clients map when the connection is closed
	clients.Delete(conn)
}

// updateActiveConnections periodically sends the list of active connections to all connected clients
func updateActiveConnections() {
	for {
		// connected bois
		var activeConnections []string
		clients.Range(func(key, value interface{}) bool {
			clientData := value.(*ClientData)
			activeConnections = append(activeConnections, clientData.Nickname)
			return true
		})

		// map with the active connections and nick
		data := map[string]interface{}{
			"activeConnections": activeConnections,
		}

		// marshal jasuon boy
		message, err := json.Marshal(data)
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

		time.Sleep(2 * time.Second)
	}
}
