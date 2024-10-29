package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ChatRoom mantém as conexões ativas e gerencia mensagens
type ChatRoom struct {
	connections map[*websocket.Conn]bool
	mu          sync.Mutex
}

func newChatRoom() *ChatRoom {
	return &ChatRoom{
		connections: make(map[*websocket.Conn]bool),
	}
}

// Adiciona uma nova conexão ao chat
func (room *ChatRoom) addConnection(conn *websocket.Conn) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.connections[conn] = true
}

// Remove uma conexão do chat
func (room *ChatRoom) removeConnection(conn *websocket.Conn) {
	room.mu.Lock()
	defer room.mu.Unlock()
	delete(room.connections, conn)
	conn.Close()
}

// Envia uma mensagem para todas as conexões ativas
func (room *ChatRoom) broadcastMessage(msg []byte) {
	room.mu.Lock()
	defer room.mu.Unlock()
	for conn := range room.connections {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			room.removeConnection(conn)
		}
	}
}

// Manipulador WebSocket
func (room *ChatRoom) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Erro ao atualizar conexão:", err)
		return
	}
	defer room.removeConnection(conn)

	room.addConnection(conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		room.broadcastMessage(msg)
	}
}

func main() {
	chatRoom := newChatRoom()

	http.HandleFunc("/ws", chatRoom.handleWebSocket)

	// Servindo o arquivo HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})

	fmt.Println("Servidor iniciado na porta 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
	}
}
