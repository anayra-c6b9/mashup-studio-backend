package rooms

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type Manager struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

func generateCode() string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = letters[rand.Intn(len(letters))]
	}
	return string(code)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (m *Manager) CreateRoom() *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	var code string
	for {
		code = generateCode()
		if _, exists := m.rooms[code]; !exists {
			break
		}
	}

	room := NewRoom(code)
	m.rooms[code] = room
	return room
}

func (m *Manager) JoinRoom(code string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[code]
	if !ok {
		return nil, errors.New("room not found")
	}

	return room, nil
}

func (m *Manager) LeaveRoom(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[code]
	if !ok {
		return
	}

	room.Leave()

	if room.Users == 0 {
		delete(m.rooms, code)
	}
}


