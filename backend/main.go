package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients = make(map[*websocket.Conn]bool)
	mu      sync.Mutex
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Convert HTTP connection to WebSocket.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	addClient(conn)
	broadcastCount() // tell everyone a new user joined

	// When this function ends (user left), remove them and send the new count.
	defer func() {
		removeClient(conn)
		broadcastCount()
	}()

	fmt.Println("Client connected")

	// Keep reading messages until the client disconnects.
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected")
			return
		}

		fmt.Println("Received:", string(message))
		broadcastMessage(message)
	}
}

func addClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()

	clients[conn] = true
}

func removeClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()

	delete(clients, conn)
}

func broadcastMessage(message []byte) {
	mu.Lock()
	defer mu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

// broadcastCount sends the number of online users to every client.
// Example message: {"type":"count","count":3}
func broadcastCount() {
	mu.Lock()
	count := len(clients)
	mu.Unlock()

	message := fmt.Sprintf(`{"type":"count","count":%d}`, count)
	broadcastMessage([]byte(message))
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("WebSocket server started on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed:", err)
	}
}
