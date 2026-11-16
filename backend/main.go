package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrade HTTP to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins
		return true
	},
}

// Connected clients registry
var (
	clients = make(map[*websocket.Conn]bool)
	mu      sync.Mutex
)

// Handle one connection
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade failed:", err)
		return
	}
	// Close on exit
	defer conn.Close()

	// Register client
	addClient(conn)
	broadcastCount()

	// Cleanup on disconnect
	defer func() {
		removeClient(conn)
		broadcastCount()
	}()

	fmt.Println("Client connected")

	// Read message loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected")
			return
		}

		// Broadcast to everyone
		fmt.Println("Received:", string(message))
		broadcastMessage(message)
	}
}

// Add client safely
func addClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()

	clients[conn] = true
}

// Remove client safely
func removeClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()

	delete(clients, conn)
}

// Send to all clients
func broadcastMessage(message []byte) {
	mu.Lock()
	defer mu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			// Drop failed client
			client.Close()
			delete(clients, client)
		}
	}
}

// broadcastCount sends the number of online users to every client.
// Example message: {"type":"count","count":3}
func broadcastCount() {
	mu.Lock()
	// Count online clients
	count := len(clients)
	mu.Unlock()

	// Build and send
	message := fmt.Sprintf(`{"type":"count","count":%d}`, count)
	broadcastMessage([]byte(message))
}

func main() {
	// Register route
	http.HandleFunc("/ws", handleWebSocket)

	// Start server
	fmt.Println("WebSocket server started on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed:", err)
	}
}
