package groupws

import (
    "encoding/json"
    "sync"
)

type BroadcastMessage struct {
    ChatID string
    Data   []byte
}

type Hub struct {
    rooms      map[string]map[*Client]bool
    clients    map[string]map[*Client]bool
    register   chan *Client
    unregister chan *Client
    broadcast  chan *BroadcastMessage
    mu         sync.RWMutex
}

func NewHub() *Hub {
    return &Hub{
        rooms:      make(map[string]map[*Client]bool),
        clients:    make(map[string]map[*Client]bool),
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
            if h.clients[client.userID] == nil {
                h.clients[client.userID] = make(map[*Client]bool)
            }
            h.clients[client.userID][client] = true
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            for chatID, clients := range h.rooms {
                delete(clients, client)
                if len(clients) == 0 {
                    delete(h.rooms, chatID)
                }
            }
            if clients, ok := h.clients[client.userID]; ok {
                delete(clients, client)
                if len(clients) == 0 {
                    delete(h.clients, client.userID)
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
                    h.mu.RUnlock()
                    h.mu.Lock()
                    delete(clients, client)
                    close(client.send)
                    h.mu.Unlock()
                    h.mu.RLock()
                }
            }
            h.mu.RUnlock()
        }
    }
}

func (h *Hub) JoinChat(chatID string, client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if h.rooms[chatID] == nil {
        h.rooms[chatID] = make(map[*Client]bool)
    }
    h.rooms[chatID][client] = true
}

func (h *Hub) LeaveChat(chatID string, client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if clients, ok := h.rooms[chatID]; ok {
        delete(clients, client)
        if len(clients) == 0 {
            delete(h.rooms, chatID)
        }
    }
}

func (h *Hub) RegisterClient(client *Client) {
    h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
    h.unregister <- client
}

func (h *Hub) Broadcast(chatID string, msg any) {
    data, err := json.Marshal(msg)
    if err != nil {
        return
    }
    h.broadcast <- &BroadcastMessage{
        ChatID: chatID,
        Data:   data,
    }
}

func (h *Hub) SendToUser(userID string, msg any) {
    data, err := json.Marshal(msg)
    if err != nil {
        return
    }

    h.mu.RLock()
    clients := h.clients[userID]
    h.mu.RUnlock()

    for client := range clients {
        select {
        case client.send <- data:
        default:
            h.mu.Lock()
            delete(clients, client)
            close(client.send)
            if len(clients) == 0 {
                delete(h.clients, userID)
            }
            h.mu.Unlock()
        }
    }
}