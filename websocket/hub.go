package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// Prevent cross-site request forgery
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub Struct
type Hub struct {
	clients    []*Client
	register   chan *Client
	unregister chan *Client
	mutex      *sync.Mutex
}

// Constructor of Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make([]*Client, 0),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		mutex:      &sync.Mutex{},
	}
}

// WebSocket Handler
func (hub *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	socket, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	// Register New Client to WebSocket
	client := NewClient(hub, socket)

	hub.register <- client

	// GoRoutine wich send messages to registered clients
	go client.Write()
}

// Connected Client Handler
func (hub *Hub) onConnect(client *Client) {
	// Console message
	log.Println("Client connected", client.socket.RemoteAddr())

	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	// Assign id for client
	client.id = client.socket.RemoteAddr().String()

	// Add client to Hub
	hub.clients = append(hub.clients, client)
}

// Disconnected Client Handler
func (hub *Hub) onDisconnect(client *Client) {
	// Console message
	log.Println("Client disconnected", client.socket.RemoteAddr())

	// Close the connection with the socket
	client.socket.Close()

	// Remove the client from the Hub by index in the slice
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	i := -1
	for j, cli := range hub.clients {
		if cli.id == client.id {
			i = j
		}
	}

	// Algorythm to delete an element from slice
	copy(hub.clients[i:], hub.clients[i+1:])
	hub.clients[len(hub.clients)-1] = nil
	hub.clients = hub.clients[:len(hub.clients)-1]
}

// Run WebSocket
func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.onConnect(client)
		case client := <-hub.unregister:
			hub.onDisconnect(client)
		}
	}
}

// Sent Messages Handler
func (hub *Hub) Broadcast(message interface{}, ignore *Client) {

	// Serialize the message
	data, _ := json.Marshal(message)

	// Send the message to all clients, ignoring the client who sends the data
	for _, client := range hub.clients {
		if client != ignore {
			client.outbound <- data
		}
	}
}
