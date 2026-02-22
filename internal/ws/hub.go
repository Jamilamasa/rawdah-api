package ws

import (
	"encoding/json"
	"sync"
)

type Hub struct {
	rooms      map[string]map[*Client]bool
	broadcast  chan BroadcastMsg
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type BroadcastMsg struct {
	Room  string
	Event interface{}
}

type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan BroadcastMsg, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			for _, room := range client.rooms {
				if h.rooms[room] == nil {
					h.rooms[room] = make(map[*Client]bool)
				}
				h.rooms[room][client] = true
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			for _, room := range client.rooms {
				if clients, ok := h.rooms[room]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.rooms, room)
					}
				}
			}
			h.mu.Unlock()
			close(client.send)

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.rooms[msg.Room]
			data, _ := json.Marshal(msg.Event)
			for client := range clients {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.rooms[msg.Room], client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Broadcast(msg BroadcastMsg) {
	h.broadcast <- msg
}
