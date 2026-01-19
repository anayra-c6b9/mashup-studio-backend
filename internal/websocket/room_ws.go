package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"anayra-c6b9.net/mashupstudio/internal/rooms"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (ok for now)
	},
}
// go get github.com/gorilla/websocket
type RoomWS struct {
	Manager *rooms.Manager
}

func NewRoomWS(m *rooms.Manager) *RoomWS {
	return &RoomWS{Manager: m}
}

// func (ws *RoomWS) Handle(w http.ResponseWriter, r *http.Request) {
// 	code := r.URL.Query().Get("code")
// 	if code == "" {
// 		http.Error(w, "Room code required", http.StatusBadRequest)
// 		return
// 	}

// 	room, err := ws.Manager.JoinRoom(code)
// 	if err != nil {
// 		http.Error(w, "Room not found", http.StatusNotFound)
// 		return
// 	}

// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		return
// 	}
// 	defer conn.Close()

// 	// When connection closes → leave room
// 	defer ws.Manager.LeaveRoom(code)

// 	// Notify join
// 	conn.WriteJSON(map[string]interface{}{
// 		"type":  "joined",
// 		"users": room.Users,
// 	})

// 	// Keep connection alive
// 	for {
// 		_, _, err := conn.ReadMessage()
// 		if err != nil {
// 			break // client disconnected
// 		}
// 	}
// }
func (ws *RoomWS) Handle(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Room code required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	room, err := ws.Manager.JoinRoom(code)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// ✅ register client safely
	room.AddClient(conn)

	// ✅ cleanup on disconnect
	defer func() {
		usersLeft := room.RemoveClient(conn)
		if usersLeft <= 0 {
			ws.Manager.LeaveRoom(code)
		}
	}()

	// notify join
	conn.WriteJSON(map[string]interface{}{
		"type":  "joined",
		"users": room.Users,
	})

	// read + broadcast
	for {
		type WSMessage struct {
			Type    string                 `json:"type"`
			Payload map[string]interface{} `json:"payload,omitempty"`
		}

		for {
			var msg WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
					break
				}

				// for now: broadcast everything
				room.Broadcast(conn, msg)
		}

	}
}
