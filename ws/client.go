package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"

	"github.com/gorilla/websocket"
)

const (
	writeWait             = 10 * time.Second
	pongWait              = 60 * time.Second
	pingPeriod            = (pongWait * 9) / 10
	maxMessageSize        = 8192
	defaultHistoryLimit   = 30
	maxHistoryLimit       = 100
)

type Client struct {
	ctx         context.Context
	conn        *websocket.Conn
	hub         *Hub
	service     ChatService
	callService CallService
	postRepo    *repository.PostRepository
	userID      string
	send        chan []byte
	joinedChats map[string]struct{}
	typingChats map[string]bool
}

func NewClient(ctx context.Context, conn *websocket.Conn, hub *Hub, service ChatService, callService CallService, postRepo *repository.PostRepository, userID string) *Client {
	return &Client{
		ctx:         ctx,
		conn:        conn,
		hub:         hub,
		service:     service,
		callService: callService,
		postRepo:    postRepo,
		userID:      userID,
		send:        make(chan []byte, 256),
		joinedChats: make(map[string]struct{}),
		typingChats: make(map[string]bool),
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
			c.sendError("định dạng tin nhắn không hợp lệ")
			continue
		}

		// Chat events cần ChatService. Call events có nil check riêng.
		if c.service == nil {
			switch event.Type {
			case "chat:join", "chat:history:more", "message:send", "typing:start", "typing:stop", "message:delete", "message:search", "message:pin", "message:unpin":
				continue
			}
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

			c.sendChatHistory("message:history", payload.ChatID, payload.BeforeCursor, resolveHistoryLimit(payload.Limit))

			// Gửi danh sách tin nhắn đã ghim
			if pins, err := c.service.GetPinnedMessages(c.ctx, c.userID, payload.ChatID); err == nil && len(pins) > 0 {
				pinResp, _ := json.Marshal(dto.WsEvent{
					Type:    "message:pinned_list",
					Payload: mustMarshal(map[string]any{"pinned_messages": pins}),
				})
				c.send <- pinResp
			}

		case "chat:history:more":
			var payload dto.ChatHistoryMorePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tải lịch sử không hợp lệ")
				continue
			}
			if _, ok := c.joinedChats[payload.ChatID]; !ok {
				if err := c.service.JoinChat(c.ctx, c.userID, payload.ChatID); err != nil {
					c.sendError(err.Error())
					continue
				}
				c.joinedChats[payload.ChatID] = struct{}{}
				c.hub.JoinChat(payload.ChatID, c)
			}
			c.sendChatHistory("message:history_more", payload.ChatID, &payload.Cursor, resolveHistoryLimit(payload.Limit))

		case "message:send":
			var payload dto.SendMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tin nhắn không hợp lệ")
				continue
			}

			msg, err := c.service.SendMessage(c.ctx, c.userID, payload.ChatID, payload.Content, payload.E2EVersion, payload.EmojiID, payload.MediaID, payload.GifURL, payload.ReplyToMessageID, payload.SharedPostID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			// Tin nhắn E2E: server không đọc được nội dung, chuyển ciphertext
			// nguyên trạng cho các client cùng chat (mỗi client tự giải mã).
			content := msg.Content
			if msg.E2EVersion == 0 {
				decrypted, err := c.service.DecryptMessage(c.ctx, msg.ChatID, msg.Content)
				if err != nil {
					c.sendError("giải mã tin nhắn thất bại")
					continue
				}
				content = decrypted
			}

			messagePayload := dto.MessagePayload{
				ID:               msg.ID,
				ChatID:           msg.ChatID,
				SenderID:         msg.SenderID,
				Content:          content,
				EmojiID:          msg.EmojiID,
				MediaID:          msg.MediaID,
				ReplyToMessageID: msg.ReplyToMessageID,
				SharedPostID:     msg.SharedPostID,
				Type:             msg.Type,
				E2EVersion:       msg.E2EVersion,
				CreatedAt:        msg.CreatedAt,
			}
			if msg.ReplyToMessageID != nil && *msg.ReplyToMessageID != "" {
				if previews := c.service.GetReplyPreviews(c.ctx, []string{*msg.ReplyToMessageID}); previews != nil {
					if preview, ok := previews[*msg.ReplyToMessageID]; ok {
						encKey, _ := c.service.GetEncryptionKey(c.ctx, msg.ChatID)
						if encKey != "" {
							if decrypted, err := utils.DecryptMessage(preview.Content, encKey); err == nil {
								preview.Content = decrypted
							}
						}
						messagePayload.ReplyTo = preview
					}
				}
			}

			if msg.SharedPostID != nil && *msg.SharedPostID != "" && c.postRepo != nil {
				if sharedPosts := loadSharedPostsForMessages(c.ctx, []models.Message{*msg}, c.postRepo); sharedPosts != nil {
					if sp, ok := sharedPosts[*msg.SharedPostID]; ok {
						messagePayload.SharedPost = sp
					}
				}
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

			isStart := event.Type == "typing:start"

			if isStart {
				if c.typingChats[payload.ChatID] {
					continue
				}
				c.typingChats[payload.ChatID] = true
			} else {
				if !c.typingChats[payload.ChatID] {
					continue
				}
				c.typingChats[payload.ChatID] = false
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
					Messages: toMessagePayloads(messages, c.userID),
				}),
			})
			c.send <- resp

		case "message:pin":
			var payload dto.PinMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu ghim không hợp lệ")
				continue
			}

			pinDTO, err := c.service.PinMessage(c.ctx, c.userID, payload.ChatID, payload.MessageID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			pinnedResp, _ := json.Marshal(dto.WsEvent{
				Type:    "message:pinned",
				Payload: mustMarshal(pinDTO),
			})
			c.hub.broadcast <- &BroadcastMessage{ChatID: payload.ChatID, Data: pinnedResp}

		case "message:unpin":
			var payload dto.UnpinMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu bỏ ghim không hợp lệ")
				continue
			}

			if err := c.service.UnpinMessage(c.ctx, c.userID, payload.ChatID, payload.MessageID); err != nil {
				c.sendError(err.Error())
				continue
			}

			unpinnedResp, _ := json.Marshal(dto.WsEvent{
				Type:    "message:unpinned",
				Payload: mustMarshal(dto.MessageUnpinnedPayload{ChatID: payload.ChatID, MessageID: payload.MessageID}),
			})
			c.hub.broadcast <- &BroadcastMessage{ChatID: payload.ChatID, Data: unpinnedResp}

		case "call:busy":
			// Client xác nhận đã nhận được thông báo busy — bỏ qua, không cần xử lý thêm
			continue

		case "call:initiate":
			if c.callService == nil {
				c.sendError("dịch vụ gọi không khả dụng")
				continue
			}
			var payload dto.CallInitiatePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu khởi tạo cuộc gọi không hợp lệ")
				continue
			}
			call, err := c.callService.InitiateCall(c.ctx, c.userID, payload)
			if err != nil {
				c.sendError(err.Error())
				continue
			}
			if call == nil {
				continue
			}
			resp, _ := json.Marshal(dto.WsEvent{
				Type:    "call:initiated",
				Payload: mustMarshal(map[string]string{"call_id": call.ID}),
			})
			c.send <- resp

		case "call:accept":
			if c.callService == nil {
				c.sendError("dịch vụ gọi không khả dụng")
				continue
			}
			var payload dto.CallActionPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu chấp nhận không hợp lệ")
				continue
			}
			if err := c.callService.AcceptCall(c.ctx, c.userID, payload.CallID); err != nil {
				c.sendError(err.Error())
				continue
			}

		case "call:reject":
			if c.callService == nil {
				c.sendError("dịch vụ gọi không khả dụng")
				continue
			}
			var payload dto.CallActionPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu từ chối không hợp lệ")
				continue
			}
			if err := c.callService.RejectCall(c.ctx, c.userID, payload.CallID); err != nil {
				c.sendError(err.Error())
				continue
			}

		case "call:end":
			if c.callService == nil {
				c.sendError("dịch vụ gọi không khả dụng")
				continue
			}
			var payload dto.CallActionPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu kết thúc không hợp lệ")
				continue
			}
			if err := c.callService.EndCall(c.ctx, c.userID, payload.CallID); err != nil {
				c.sendError(err.Error())
				continue
			}

		case "call:signal":
			if c.callService == nil {
				c.sendError("dịch vụ gọi không khả dụng")
				continue
			}
			var payload dto.CallSignalPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tín hiệu không hợp lệ")
				continue
			}
			// Phase 4 fix: Validate signal payload size to prevent DoS
			// via oversized payloads forwarded to the other user.
			if err := payload.Validate(); err != nil {
				c.sendError(err.Error())
				continue
			}
			if err := c.callService.HandleSignal(c.ctx, c.userID, payload.CallID, payload.Signal); err != nil {
				c.sendError(err.Error())
				continue
			}

		case "call:toggle_mute":
			if c.callService == nil {
				c.sendError("dịch vụ gọi không khả dụng")
				continue
			}
			var payload struct {
				CallID string `json:"call_id"`
				Muted  bool   `json:"muted"`
			}
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu mute không hợp lệ")
				continue
			}
			if err := c.callService.ToggleMute(c.ctx, c.userID, payload.CallID, payload.Muted); err != nil {
				c.sendError(err.Error())
				continue
			}

		case "call:video_toggle":
			if c.callService == nil {
				c.sendError("dịch vụ gọi không khả dụng")
				continue
			}
			var payload struct {
				CallID       string `json:"call_id"`
				VideoEnabled bool   `json:"video_enabled"`
			}
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu video toggle không hợp lệ")
				continue
			}
			if err := c.callService.ToggleVideo(c.ctx, c.userID, payload.CallID, payload.VideoEnabled); err != nil {
				c.sendError(err.Error())
				continue
			}

		default:
			// Phase 1 fix: Removed the extra ReadMessage() call that was here.
			// The outer for-loop already calls ReadMessage() at the top of
			// each iteration. The previous code consumed the next client
			// message, silently dropping it with no error response.
			c.sendError("loại sự kiện không xác định")
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

func toMessagePayloads(messages []models.Message, userID string, sharedPosts ...map[string]*dto.SharedPostPayload) []dto.MessagePayload {
	var postsMap map[string]*dto.SharedPostPayload
	if len(sharedPosts) > 0 {
		postsMap = sharedPosts[0]
	}
	result := make([]dto.MessagePayload, 0, len(messages))
	for _, msg := range messages {
		deleted := isMessageDeletedFor(msg, userID)
		content := msg.Content
		if deleted {
			content = ""
		}
		payload := dto.MessagePayload{
			ID:               msg.ID,
			ChatID:           msg.ChatID,
			SenderID:         msg.SenderID,
			Content:          content,
			EmojiID:          msg.EmojiID,
			MediaID:          msg.MediaID,
			ReplyToMessageID: msg.ReplyToMessageID,
			SharedPostID:     msg.SharedPostID,
			Type:             msg.Type,
			E2EVersion:       msg.E2EVersion,
			Deleted:          deleted,
			CreatedAt:        msg.CreatedAt,
		}
		if postsMap != nil && msg.SharedPostID != nil {
			if sp, ok := postsMap[*msg.SharedPostID]; ok {
				payload.SharedPost = sp
			}
		}
		result = append(result, payload)
	}
	return result
}

func loadSharedPostsForMessages(ctx context.Context, messages []models.Message, postRepo *repository.PostRepository) map[string]*dto.SharedPostPayload {
	postIDs := make([]string, 0)
	seen := make(map[string]struct{})
	for _, m := range messages {
		if m.SharedPostID != nil && *m.SharedPostID != "" {
			if _, ok := seen[*m.SharedPostID]; !ok {
				postIDs = append(postIDs, *m.SharedPostID)
				seen[*m.SharedPostID] = struct{}{}
			}
		}
	}
	if len(postIDs) == 0 {
		return nil
	}
	posts, err := postRepo.FindByIDs(ctx, postIDs)
	if err != nil || len(posts) == 0 {
		return nil
	}
	result := make(map[string]*dto.SharedPostPayload, len(posts))
	for i := range posts {
		p := &posts[i]
		result[p.ID] = &dto.SharedPostPayload{
			ID:          p.ID,
			UserID:      p.UserID,
			Username:    p.Username,
			DisplayName: p.DisplayName,
			AvatarURI:   p.AvatarURI,
			Title:       p.Title,
			Content:     p.Content,
		}
		if len(p.Media) > 0 {
			result[p.ID].MediaURI = p.Media[0].FileURI
			result[p.ID].MediaType = p.Media[0].FileType
		}
	}
	return result
}

func isMessageDeletedFor(msg models.Message, userID string) bool {
	if msg.SenderID == userID {
		return msg.DeletedForSender
	}
	return msg.DeletedForReceiver
}

func resolveHistoryLimit(limit *int) int {
	if limit == nil || *limit <= 0 {
		return defaultHistoryLimit
	}
	if *limit > maxHistoryLimit {
		return maxHistoryLimit
	}
	return *limit
}

// sendChatHistory tải một trang lịch sử (mới → cũ theo cursor), enrich reply
// previews + shared posts rồi gửi qua WS. eventType = "message:history" (lần
// đầu) hoặc "message:history_more" (tải thêm trang cũ hơn).
func (c *Client) sendChatHistory(eventType, chatID string, cursor *dto.HistoryCursor, limit int) {
	history, err := c.service.GetMessagesHistory(c.ctx, c.userID, chatID, cursor, limit)
	if err != nil {
		c.sendError(fmt.Sprintf("lấy lịch sử: %v", err))
		return
	}

	hasMore := false
	var nextCursor *dto.HistoryCursor
	if len(history) > limit {
		history = history[:limit]
		hasMore = true
	}
	if len(history) > 0 {
		oldest := history[len(history)-1]
		nextCursor = &dto.HistoryCursor{CreatedAt: oldest.CreatedAt, ID: oldest.ID}
	}

	var sharedPosts map[string]*dto.SharedPostPayload
	if c.postRepo != nil {
		sharedPosts = loadSharedPostsForMessages(c.ctx, history, c.postRepo)
	}
	payloads := toMessagePayloads(history, c.userID, sharedPosts)

	replyIDs := make([]string, 0)
	for _, p := range payloads {
		if p.ReplyToMessageID != nil && *p.ReplyToMessageID != "" {
			replyIDs = append(replyIDs, *p.ReplyToMessageID)
		}
	}
	if previews := c.service.GetReplyPreviews(c.ctx, replyIDs); previews != nil {
		encKey, _ := c.service.GetEncryptionKey(c.ctx, chatID)
		for _, preview := range previews {
			if encKey != "" {
				if decrypted, err := utils.DecryptMessage(preview.Content, encKey); err == nil {
					preview.Content = decrypted
				}
			}
		}
		for i := range payloads {
			if payloads[i].ReplyToMessageID != nil {
				if preview, ok := previews[*payloads[i].ReplyToMessageID]; ok {
					payloads[i].ReplyTo = preview
				}
			}
		}
	}

	resp, _ := json.Marshal(dto.WsEvent{
		Type: eventType,
		Payload: mustMarshal(map[string]any{
			"chat_id":     chatID,
			"messages":    payloads,
			"has_more":    hasMore,
			"next_cursor": nextCursor,
		}),
	})
	c.send <- resp
}

func (c *Client) handleTypingEvent(chatID, eventType string) bool {
	switch eventType {
	case "typing:start":
		if c.typingChats[chatID] {
			return false
		}
		c.typingChats[chatID] = true
		return true
	case "typing:stop":
		if !c.typingChats[chatID] {
			return false
		}
		c.typingChats[chatID] = false
		return true
	default:
		return false
	}
}
