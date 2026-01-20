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
	Queue	 []string
	CurrentTrack string // trackId
	IsPlaying    bool

}

func NewRoom(code string) *Room {
	return &Room{
		Code:      code,
		Users:     0,
		Clients:   make(map[*websocket.Conn]bool),
		CreatedAt: time.Now(),
		Queue:   []string{},
		CurrentTrack: "",
		IsPlaying:    false,
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
	// r.Users++
}

func (r *Room) RemoveClient(conn *websocket.Conn) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Clients, conn)
	// r.Users--
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

func (r *Room) UserCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Clients)
}

func (r *Room) AddToQueue(trackID string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Queue = append(r.Queue, trackID)
	// return a copy (safer)
	out := make([]string, len(r.Queue))
	copy(out, r.Queue)
	return out
}

func (r *Room) QueueSnapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]string, len(r.Queue))
	copy(out, r.Queue)
	return out
}

func (r *Room) RemoveFromQueueByID(trackID string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if trackID == "" {
		out := make([]string, len(r.Queue))
		copy(out, r.Queue)
		return out
	}

	newQueue := make([]string, 0, len(r.Queue))
	for _, id := range r.Queue {
		if id != trackID {
			newQueue = append(newQueue, id)
		}
	}

	r.Queue = newQueue

	out := make([]string, len(r.Queue))
	copy(out, r.Queue)
	return out
}

func (r *Room) EnsureCurrentFromQueue() {
	if r.CurrentTrack == "" && len(r.Queue) > 0 {
		r.CurrentTrack = r.Queue[0]
	}
}

func (r *Room) Play(trackID string) (bool, string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if trackID != "" {
		r.CurrentTrack = trackID
	} else {
		r.EnsureCurrentFromQueue()
	}

	if r.CurrentTrack == "" {
		r.IsPlaying = false
		return r.IsPlaying, r.CurrentTrack
	}

	r.IsPlaying = true
	return r.IsPlaying, r.CurrentTrack
}

func (r *Room) Pause() (bool, string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.IsPlaying = false
	return r.IsPlaying, r.CurrentTrack
}

// FIFO "next": move to next item in queue
func (r *Room) NextTrack() (bool, string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Queue) == 0 {
		r.CurrentTrack = ""
		r.IsPlaying = false
		return r.IsPlaying, r.CurrentTrack
	}

	// If no current, start from first
	if r.CurrentTrack == "" {
		r.CurrentTrack = r.Queue[0]
		r.IsPlaying = true
		return r.IsPlaying, r.CurrentTrack
	}

	// Find current in queue
	idx := -1
	for i, id := range r.Queue {
		if id == r.CurrentTrack {
			idx = i
			break
		}
	}

	// If not found, go to first
	if idx == -1 {
		r.CurrentTrack = r.Queue[0]
		r.IsPlaying = true
		return r.IsPlaying, r.CurrentTrack
	}

	// Move to next, clamp at end (no wrap)
	if idx+1 >= len(r.Queue) {
		// end reached: stop (or keep last paused)
		r.IsPlaying = true
		r.CurrentTrack = r.Queue[len(r.Queue)-1]
		return r.IsPlaying, r.CurrentTrack
	}

	r.CurrentTrack = r.Queue[idx+1]
	r.IsPlaying = true
	return r.IsPlaying, r.CurrentTrack
}

func (r *Room) PreviousTrack() (bool, string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Queue) == 0 {
		r.CurrentTrack = ""
		r.IsPlaying = false
		return r.IsPlaying, r.CurrentTrack
	}

	if r.CurrentTrack == "" {
		r.CurrentTrack = r.Queue[0]
		r.IsPlaying = true
		return r.IsPlaying, r.CurrentTrack
	}

	idx := -1
	for i, id := range r.Queue {
		if id == r.CurrentTrack {
			idx = i
			break
		}
	}

	if idx == -1 {
		r.CurrentTrack = r.Queue[0]
		r.IsPlaying = true
		return r.IsPlaying, r.CurrentTrack
	}

	// Move to prev, clamp at start
	if idx-1 < 0 {
		// already at first: keep as-is
		r.IsPlaying = true
		return r.IsPlaying, r.CurrentTrack
	}

	r.CurrentTrack = r.Queue[idx-1]
	r.IsPlaying = true
	return r.IsPlaying, r.CurrentTrack
}

func (r *Room) PlaybackSnapshot() map[string]interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	return map[string]interface{}{
		"isPlaying":    r.IsPlaying,
		"currentTrack": r.CurrentTrack,
	}
}
