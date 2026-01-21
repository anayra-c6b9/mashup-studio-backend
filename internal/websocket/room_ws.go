package websocket

import (
	"net/http"
	"log"

	"github.com/gorilla/websocket"
	"anayra-c6b9.net/mashupstudio/internal/rooms"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (ok for now)
	},
}
// go get github.com/gorilla/websocket
// type RoomWS struct {
// 	Manager *rooms.Manager
// }
type RoomWS struct {
	Manager    *rooms.Manager
	InfoLog    *log.Logger
}

type WSMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// func NewRoomWS(m *rooms.Manager) *RoomWS {
// 	return &RoomWS{Manager: m}
// }
func NewRoomWS(m *rooms.Manager, infoLog *log.Logger) *RoomWS {
	return &RoomWS{
		Manager: m,
		InfoLog: infoLog,
	}
}

func (ws *RoomWS) Handle(w http.ResponseWriter, r *http.Request) {
	ws.InfoLog.Printf("WS request: %s %s", r.RemoteAddr, r.URL.String())

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
	conn.WriteJSON(map[string]interface{}{
		"type": "queue",
		"payload": map[string]interface{}{
			"queue": room.QueueSnapshot(),
		},
	})

	// room.Broadcast(nil, map[string]interface{}{
	// 	"type":  "users",
	// 	"count": room.UserCount(),
	// })
	// ws.InfoLog.Printf(
	// 	"Room %s: user joined, users=%d",
	// 	code,
	// 	room.UserCount(),
	// )

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
	
	type WSMessage struct {
		Type    string                 `json:"type"`
		Payload map[string]interface{} `json:"payload,omitempty"`
	}

	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Type {
		case "queue_add":
			tid, _ := msg.Payload["trackId"].(string)
			if tid == "" {
				continue
			}

			queue := room.AddToQueue(tid)

			room.Broadcast(nil, map[string]interface{}{
				"type": "queue_add",
				"payload": map[string]interface{}{
					"trackId": tid,
				},
			})

			// broadcast updated queue to everyone
			room.Broadcast(nil, map[string]interface{}{
				"type": "queue",
				"payload": map[string]interface{}{
					"queue": queue,
				},
			})
		
		case "queue_remove":
			tid, _ := msg.Payload["trackId"].(string)
			if tid == "" {
				continue
			}

			// update server queue
			queue := room.RemoveFromQueueByID(tid)

			// 1) broadcast the event (fast for clients)
			room.Broadcast(nil, map[string]interface{}{
				"type": "queue_remove",
				"payload": map[string]interface{}{
					"trackId": tid,
				},
			})

			// 2) broadcast the full updated queue (source of truth)
			room.Broadcast(nil, map[string]interface{}{
				"type": "queue",
				"payload": map[string]interface{}{
					"queue": queue,
				},
			})
		
		case "play":
			tid, _ := msg.Payload["trackId"].(string) // optional
			_, current := room.Play(tid)

			// event (optional)
			room.Broadcast(nil, map[string]interface{}{
				"type": "play",
				"payload": map[string]interface{}{
					"currentTrack": current,
				},
			})

			// authoritative state
			room.Broadcast(nil, map[string]interface{}{
				"type":    "playback",
				"payload": room.PlaybackSnapshot(),
			})

		case "pause":
			_, _ = room.Pause()

			room.Broadcast(nil, map[string]interface{}{
				"type": "pause",
			})

			room.Broadcast(nil, map[string]interface{}{
				"type":    "playback",
				"payload": room.PlaybackSnapshot(),
			})

		case "next":
			_, _ = room.NextTrack()

			room.Broadcast(nil, map[string]interface{}{
				"type": "next",
			})

			room.Broadcast(nil, map[string]interface{}{
				"type":    "playback",
				"payload": room.PlaybackSnapshot(),
			})

		case "prev":
			_, _ = room.PreviousTrack()

			room.Broadcast(nil, map[string]interface{}{
				"type": "prev",
			})

			room.Broadcast(nil, map[string]interface{}{
				"type":    "playback",
				"payload": room.PlaybackSnapshot(),
			})


		default:
			// optional: broadcast other messages or ignore
		}
	}



}
