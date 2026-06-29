package ws

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"linkup/dto"
	"linkup/models"

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
	service     ChatService
	userID      string
	send        chan []byte
	joinedChats map[string]struct{}
}

func NewClient(ctx context.Context, conn *websocket.Conn, hub *Hub, service ChatService, userID string) *Client {
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

		if c.service == nil {
			continue
		}

		var event dto.WsEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			c.sendError("định dạng tin nhắn không hợp lệ")
			continue
		}

		switch event.Type {
		case "chat:join":
			var payload dto.ChatJoinPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tham gia không hợp lệ")
				continue
			}
			if err := c.service.JoinChat(c.ctx, c.userID, payload.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}
			c.joinedChats[payload.ChatID] = struct{}{}
			c.hub.JoinChat(payload.ChatID, c)

			history, err := c.service.GetAllMessages(c.ctx, c.userID, payload.ChatID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			resp, _ := json.Marshal(dto.WsEvent{
				Type: "message:history",
				Payload: mustMarshal(map[string]any{
					"chat_id":  payload.ChatID,
					"messages": toMessagePayloads(history),
				}),
			})
			c.send <- resp

		case "message:send":
			var payload dto.SendMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tin nhắn không hợp lệ")
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
				c.sendError("dữ liệu gõ chữ không hợp lệ")
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
				c.sendError("dữ liệu xóa không hợp lệ")
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

		case "message:search":
			var payload dto.SearchMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tìm kiếm không hợp lệ")
				continue
			}

			messages, err := c.service.SearchMessages(c.ctx, c.userID, payload.ChatID, payload.Keyword)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			resp, _ := json.Marshal(dto.WsEvent{
				Type: "message:search_result",
				Payload: mustMarshal(dto.SearchMessageResultPayload{
					ChatID:   payload.ChatID,
					Keyword:  payload.Keyword,
					Messages: toMessagePayloads(messages),
				}),
			})
			c.send <- resp

		default:
			c.sendError("loại sự kiện không xác định")
			_, _, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("ws read error: %v", err)
				return
			}
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

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
			}

			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write([]byte("\n")); err != nil {
					return
				}
				if _, err := w.Write(<-c.send); err != nil {
					return
				}
			}

			if err := w.Close(); err != nil {
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

func toMessagePayloads(messages []models.Message) []dto.MessagePayload {
	result := make([]dto.MessagePayload, 0, len(messages))
	for _, msg := range messages {
		result = append(result, dto.MessagePayload{
			ID:        msg.ID,
			ChatID:    msg.ChatID,
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			EmojiID:   msg.EmojiID,
			MediaID:   msg.MediaID,
			CreatedAt: msg.CreatedAt,
		})
	}
	return result
}
