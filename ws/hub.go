package ws

import (
    "sync"
)

type BroadcastMessage struct {
    ChatID string
    Data   []byte
}

type Hub struct {
    rooms      map[string]map[*Client]bool
    register   chan *Client
    unregister chan *Client
    broadcast  chan *BroadcastMessage
    mu         sync.RWMutex
}

func NewHub() *Hub {
    return &Hub{
        rooms:      make(map[string]map[*Client]bool),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan *BroadcastMessage),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            if _, ok := h.rooms["__global__"]; !ok {
                h.rooms["__global__"] = make(map[*Client]bool)
            }
            h.rooms["__global__"][client] = true
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            for chatID, clients := range h.rooms {
                delete(clients, client)
                if len(clients) == 0 {
                    delete(h.rooms, chatID)
                }
            }
            h.mu.Unlock()
            close(client.send)

        case message := <-h.broadcast:
            h.mu.RLock()
            clients := h.rooms[message.ChatID]
            for client := range clients {
                select {
                case client.send <- message.Data:
                default:
                    close(client.send)
                    delete(clients, client)
                }
            }
            h.mu.RUnlock()
        }
    }
}

func (h *Hub) JoinChat(chatID string, client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if _, ok := h.rooms[chatID]; !ok {
        h.rooms[chatID] = make(map[*Client]bool)
    }
    h.rooms[chatID][client] = true
}

func (h *Hub) RegisterClient(client *Client)  {
	h.register <- client
}