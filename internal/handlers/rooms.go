package handlers

import (
	"encoding/json"
	"net/http"

	"anayra-c6b9.net/mashupstudio/internal/rooms"
)

type RoomHandler struct {
	Manager *rooms.Manager
}

func NewRoomHandler(m *rooms.Manager) *RoomHandler {
	return &RoomHandler{Manager: m}
}

// POST /api/rooms/create
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	room := h.Manager.CreateRoom()

	json.NewEncoder(w).Encode(map[string]string{
		"code": room.Code,
	})
}

// POST /api/rooms/join
func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Code string `json:"code"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.Code == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	room, err := h.Manager.JoinRoom(body.Code)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":  room.Code,
		"users": room.Users,
	})
}
