package ws

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"linkup/dto"
	"linkup/services"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type Client struct {
	ctx         context.Context
	conn        *websocket.Conn
	hub         *Hub
	service     *services.ChatService
	userID      string
	send        chan []byte
	joinedChats map[string]struct{}
}

func NewClient(ctx context.Context, conn *websocket.Conn, hub *Hub, service *services.ChatService, userID string) *Client {
	return &Client{
		ctx:         ctx,
		conn:        conn,
		hub:         hub,
		service:     service,
		userID:      userID,
		send:        make(chan []byte, 256),
		joinedChats: make(map[string]struct{}),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws read error: %v", err)
			}
			return
		}

		var event dto.WsEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			c.sendError("invalid message format")
			continue
		}

		switch event.Type {
		case "chat:join":
			var payload dto.ChatJoinPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("invalid join payload")
				continue
			}
			if err := c.service.JoinChat(c.ctx, c.userID, payload.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}
			c.joinedChats[payload.ChatID] = struct{}{}
			c.hub.JoinChat(payload.ChatID, c)

		case "message:send":
			var payload dto.SendMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("invalid message payload")
				continue
			}

			msg, err := c.service.SendMessage(c.ctx, c.userID, payload.ChatID, payload.Content, payload.EmojiID, payload.MediaID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			messagePayload := dto.MessagePayload{
				ID:        msg.ID,
				ChatID:    msg.ChatID,
				SenderID:  msg.SenderID,
				Content:   msg.Content,
				EmojiID:   msg.EmojiID,
				MediaID:   msg.MediaID,
				CreatedAt: msg.CreatedAt,
			}

			resp, _ := json.Marshal(dto.WsEvent{
				Type:    "message:new",
				Payload: mustMarshal(messagePayload),
			})
			c.hub.broadcast <- &BroadcastMessage{ChatID: payload.ChatID, Data: resp}

		case "typing:start", "typing:stop":
			var payload dto.TypingPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("invalid typing payload")
				continue
			}
			payload.UserID = c.userID
			payload.IsTyping = event.Type == "typing:start"
			resp, _ := json.Marshal(dto.WsEvent{
				Type:    "typing",
				Payload: mustMarshal(payload),
			})
			c.hub.broadcast <- &BroadcastMessage{ChatID: payload.ChatID, Data: resp}

		case "message:delete":
			var payload dto.DeleteMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("invalid delete payload")
				continue
			}

			msg, err := c.service.DeleteMessage(c.ctx, c.userID, payload.MessageID, payload.Mode)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			deletedPayload := dto.MessageDeletedPayload{
				ChatID:    payload.ChatID,
				MessageID: msg.ID,
				DeletedBy: c.userID,
				Mode:      payload.Mode,
			}

			ack, _ := json.Marshal(dto.WsEvent{
				Type:    "message:deleted",
				Payload: mustMarshal(deletedPayload),
			})

			if strings.EqualFold(payload.Mode, "all") {
				c.hub.broadcast <- &BroadcastMessage{ChatID: payload.ChatID, Data: ack}
			} else {
				c.send <- ack
			}

		default:
			c.sendError("unknown event type")
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendError(text string) {
	payload := dto.WsEvent{
		Type:    "error",
		Payload: mustMarshal(map[string]string{"message": text}),
	}
	data, _ := json.Marshal(payload)
	c.send <- data
}

func mustMarshal(v any) json.RawMessage {
	out, _ := json.Marshal(v)
	return out
}
