package rooms

import (
	// "math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Room struct {
	Code      string
	Users     int
	Clients   map[*websocket.Conn]bool
	CreatedAt time.Time
	mu        sync.Mutex
}

func NewRoom(code string) *Room {
	return &Room{
		Code:      code,
		Users:     0,
		Clients:   make(map[*websocket.Conn]bool),
		CreatedAt: time.Now(),
	}
}


func (r *Room) Join() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Users++
}

func (r *Room) Leave() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Users--
	return r.Users
}

func (r *Room) AddClient(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[conn] = true
	r.Users++
}

func (r *Room) RemoveClient(conn *websocket.Conn) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Clients, conn)
	r.Users--
	return r.Users
}

func (r *Room) Broadcast(sender *websocket.Conn, msg interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for client := range r.Clients {
		if client != sender {
			client.WriteJSON(msg)
		}
	}
}
