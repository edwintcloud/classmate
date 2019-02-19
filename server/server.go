package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Message is our message struct
type Message struct {
	IP      string `json:"ip"`
	Message string `json:"message"`
}

var clients sync.Map
var messages sync.Map
var broadcast = make(chan Message)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {

	// create simple http file server
	fs := http.FileServer(http.Dir("../client/build"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients.Store(ws, true)

	// convert messages from sync map to regular map
	tmpMessages := make(map[string][]string)
	messages.Range(func(k, v interface{}) bool {
		tmpMessages[k.(string)] = v.([]string)
		return true
	})

	// send messages to new client
	ws.WriteJSON(tmpMessages)

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			clients.Delete(ws)
			break
		}

		// add message to messages map
		if v, ok := messages.Load(msg.IP); ok {
			v = append(v.([]string), msg.Message)
			messages.Store(msg.IP, v)
		} else {
			messages.Store(msg.IP, []string{msg.Message})
		}

		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		clients.Range(func(k, v interface{}) bool {
			err := k.(*websocket.Conn).WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				k.(*websocket.Conn).Close()
				clients.Delete(k)
				return false
			}
			return true
		})
	}
}
