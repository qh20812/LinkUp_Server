package groupws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/services"
	"linkup/utils"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type Client struct {
	ctx            context.Context
	conn           *websocket.Conn
	hub            *Hub
	messageService *services.GroupMessageService
	groupService   *services.GroupChatService
	userID         string
	send           chan []byte
	joinedChats    map[string]struct{}
	typingChats    map[string]bool
	mode           string
	groupChatHub   *Hub
	activeCallID   string
	callCreator    bool
}

func NewClientWithMode(
	ctx context.Context,
	conn *websocket.Conn,
	hub *Hub,
	messageService *services.GroupMessageService,
	groupService *services.GroupChatService,
	userID string,
	mode string,
	groupChatHub *Hub,
) *Client {
	return &Client{
		ctx:            ctx,
		conn:           conn,
		hub:            hub,
		messageService: messageService,
		groupService:   groupService,
		userID:         userID,
		send:           make(chan []byte, 256),
		joinedChats:    make(map[string]struct{}),
		typingChats:    make(map[string]bool),
		mode:           mode,
		groupChatHub:   groupChatHub,
	}
}

func NewClient(
	ctx context.Context,
	conn *websocket.Conn,
	hub *Hub,
	messageService *services.GroupMessageService,
	groupService *services.GroupChatService,
	userID string,
) *Client {
	return NewClientWithMode(ctx, conn, hub, messageService, groupService, userID, "chat", nil)
}

func (c *Client) ReadPump() {
	defer func() {
		// Khi socket rơi đột ngột, dọn toàn bộ session mà user đang tham gia để tránh treo call.
		c.cleanupDisconnectedCallSessions()
		c.hub.UnregisterClient(c)
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
				log.Printf("group ws read error: %v", err)
			}
			return
		}

		var event dto.WsEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			c.sendError("định dạng tin nhắn không hợp lệ")
			continue
		}

		if c.mode == "call" {
			switch event.Type {
			case "group:call:create", "group:call:request-join", "group:call:approve-join", "group:call:reject-join", "group:call:signal", "group:call:end", "group:call:toggle-mic", "group:call:toggle-mute", "group:call:toggle-video", "group:call:list-participants", "group:call:get-participants":
			default:
				c.sendError("event này chỉ dùng trên endpoint call riêng")
				continue
			}
		}
		if c.mode == "chat" && strings.HasPrefix(event.Type, "group:call:") {
			c.sendError("group call chỉ dùng trên endpoint /api/group-calls/ws")
			continue
		}

		switch event.Type {

		case "group:join":
			var payload dto.GroupJoinPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tham gia nhóm không hợp lệ")
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, payload.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			c.joinedChats[payload.ChatID] = struct{}{}
			c.hub.JoinChat(payload.ChatID, c)

			history, err := c.messageService.GetAllMessagesRaw(c.ctx, c.userID, payload.ChatID)
			if err != nil {
				c.sendError(fmt.Sprintf("lấy lịch sử thất bại: %v", err))
				continue
			}

		for i := range history {
			_ = c.messageService.DecryptMessageContent(c.ctx, &history[i])
		}

		senderIDs := make([]string, 0, len(history))
		seen := make(map[string]struct{})
		for _, m := range history {
			if m.SenderID != "" && m.SenderID != "SYSTEM" {
				if _, ok := seen[m.SenderID]; !ok {
					senderIDs = append(senderIDs, m.SenderID)
					seen[m.SenderID] = struct{}{}
				}
			}
		}
		profiles := c.messageService.GetMemberProfiles(c.ctx, payload.ChatID, senderIDs)
		sharedPosts := c.messageService.LoadSharedPosts(c.ctx, history)

		msgs := make([]dto.MessagePayload, 0, len(history))
		for _, m := range history {
			p := dto.MessagePayload{
				ID:               m.ID,
				ChatID:           m.ChatID,
				SenderID:         m.SenderID,
				Content:          m.Content,
				EmojiID:          m.EmojiID,
				MediaID:          m.MediaID,
				ReplyToMessageID: m.ReplyToMessageID,
				SharedPostID:     m.SharedPostID,
				Type:             m.Type,
				IsAnonymized:     m.IsAnonymized,
				AnonymousName:    m.AnonymousName,
				Deleted:          m.DeletedAt != nil,
				CreatedAt:        m.CreatedAt,
			}
			if prof, ok := profiles[m.SenderID]; ok {
				p.SenderName = prof.DisplayName
				p.SenderAvatar = prof.AvatarURI
			}
			if sharedPosts != nil && m.SharedPostID != nil {
				if sp, ok := sharedPosts[*m.SharedPostID]; ok {
					p.SharedPost = sp
				}
			}
			msgs = append(msgs, p)
		}

		replyIDs := make([]string, 0)
		for _, m := range msgs {
			if m.ReplyToMessageID != nil && *m.ReplyToMessageID != "" {
				replyIDs = append(replyIDs, *m.ReplyToMessageID)
			}
		}
		if previews := c.messageService.GetChatRepo().GetReplyPreviews(c.ctx, replyIDs); previews != nil {
			if encKey, err := c.messageService.GetChatRepo().GetEncryptionKey(c.ctx, payload.ChatID); err == nil {
				for _, preview := range previews {
					if decrypted, err := utils.DecryptMessage(preview.Content, encKey); err == nil {
						preview.Content = decrypted
					}
				}
			}
			for i := range msgs {
				if msgs[i].ReplyToMessageID != nil {
					if preview, ok := previews[*msgs[i].ReplyToMessageID]; ok {
						msgs[i].ReplyTo = preview
					}
				}
			}
		}

			callDocs, err := c.messageService.GetGroupCallsByChatID(c.ctx, c.userID, payload.ChatID)
			if err != nil {
				c.sendError(fmt.Sprintf("lấy lịch sử cuộc gọi thất bại: %v", err))
				continue
			}
			callItems := make([]dto.GroupCallHistoryItem, 0, len(callDocs))
			for _, call := range callDocs {
				var endedAt *string
				if call.EndedAt != nil {
					s := call.EndedAt.UTC().Format(time.RFC3339)
					endedAt = &s
				}
				callItems = append(callItems, dto.GroupCallHistoryItem{
					CallID:       call.CallID,
					ChatID:       call.ChatID,
					CallerID:     call.CallerID,
					Participants: call.Participants,
					Status:       call.Status,
					CreatedAt:    call.CreatedAt.UTC().Format(time.RFC3339),
					EndedAt:      endedAt,
				})
			}

			c.sendEvent("group:history", dto.GroupHistoryPayload{
				ChatID:   payload.ChatID,
				Messages: msgs,
				Calls:    callItems,
			})

		case "group:message:send":
			var payload dto.GroupSendMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tin nhắn không hợp lệ")
				continue
			}

		msg, err := c.messageService.SendMessage(
			c.ctx,
			c.userID,
			payload.ChatID,
			payload.Content,
			payload.EmojiID,
			payload.MediaID,
			payload.GifURL,
			payload.ReplyToMessageID,
			payload.SharedPostID,
		)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

		_ = c.messageService.DecryptMessageContent(c.ctx, msg)

		newPayload := dto.MessagePayload{
			ID:               msg.ID,
			ChatID:           msg.ChatID,
			SenderID:         msg.SenderID,
			Content:          msg.Content,
			EmojiID:          msg.EmojiID,
			MediaID:          msg.MediaID,
			ReplyToMessageID: msg.ReplyToMessageID,
			SharedPostID:     msg.SharedPostID,
			Type:             msg.Type,
			CreatedAt:        msg.CreatedAt,
		}
		profiles := c.messageService.GetMemberProfiles(c.ctx, payload.ChatID, []string{msg.SenderID})
		if prof, ok := profiles[msg.SenderID]; ok {
			newPayload.SenderName = prof.DisplayName
			newPayload.SenderAvatar = prof.AvatarURI
		}
		if msg.ReplyToMessageID != nil && *msg.ReplyToMessageID != "" {
			if previews := c.messageService.GetChatRepo().GetReplyPreviews(c.ctx, []string{*msg.ReplyToMessageID}); previews != nil {
				if preview, ok := previews[*msg.ReplyToMessageID]; ok {
					if encKey, err := c.messageService.GetChatRepo().GetEncryptionKey(c.ctx, payload.ChatID); err == nil {
						if decrypted, err := utils.DecryptMessage(preview.Content, encKey); err == nil {
							preview.Content = decrypted
						}
					}
					newPayload.ReplyTo = preview
				}
			}
		}

		if msg.SharedPostID != nil && *msg.SharedPostID != "" {
			if sharedPosts := c.messageService.LoadSharedPosts(c.ctx, []models.Message{*msg}); sharedPosts != nil {
				if sp, ok := sharedPosts[*msg.SharedPostID]; ok {
					newPayload.SharedPost = sp
				}
			}
		}

		c.hub.Broadcast(payload.ChatID, dto.WsEvent{
			Type:    "group:message:new",
			Payload: mustMarshal(newPayload),
		})

		case "group:typing:start", "group:typing:stop":
			var payload dto.GroupTypingPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu typing không hợp lệ")
				continue
			}

			isTyping := event.Type == "group:typing:start"
			if c.typingChats[payload.ChatID] == isTyping {
				continue
			}
			c.typingChats[payload.ChatID] = isTyping

			c.hub.Broadcast(payload.ChatID, dto.WsEvent{
				Type: "group:typing",
				Payload: mustMarshal(map[string]any{
					"chat_id":   payload.ChatID,
					"user_id":   c.userID,
					"is_typing": isTyping,
				}),
			})

		case "group:message:search":
			var payload dto.GroupSearchPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tìm kiếm không hợp lệ")
				continue
			}

			messages, err := c.messageService.SearchMessages(c.ctx, c.userID, payload.ChatID, payload.Keyword)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			out := make([]dto.MessagePayload, 0, len(messages))
			for _, m := range messages {
				out = append(out, dto.MessagePayload{
					ID:        m.ID,
					ChatID:    m.ChatID,
					SenderID:  m.SenderID,
					Content:   m.Content,
					EmojiID:   m.EmojiID,
					MediaID:   m.MediaID,
					CreatedAt: m.CreatedAt,
				})
			}

		c.sendEvent("group:message:search_result", map[string]any{
			"chat_id":  payload.ChatID,
			"keyword":  payload.Keyword,
			"messages": out,
		})

		case "group:message:delete":
			var payload dto.DeleteMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu xóa không hợp lệ")
				continue
			}

			msg, err := c.messageService.DeleteMessage(c.ctx, c.userID, payload.MessageID, payload.Mode)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			deletedPayload := dto.MessageDeletedPayload{
				ChatID:    msg.ChatID,
				MessageID: msg.ID,
				DeletedBy: c.userID,
				Mode:      payload.Mode,
			}

			ack, _ := json.Marshal(dto.WsEvent{
				Type:    "group:message:deleted",
				Payload: mustMarshal(deletedPayload),
			})

			if strings.EqualFold(payload.Mode, "all") {
				c.hub.broadcast <- &BroadcastMessage{ChatID: msg.ChatID, Data: ack}
			} else {
				c.send <- ack
			}

		case "group:leave":
			var payload dto.GroupLeavePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu rời nhóm không hợp lệ")
				continue
			}

			if err := c.groupService.LeaveGroup(c.ctx, payload.ChatID, c.userID, payload.LeaveMode, payload.HistoryMode); err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.LeaveChat(payload.ChatID, c)
			delete(c.joinedChats, payload.ChatID)

			// Chỉ public leave mới broadcast toàn nhóm
			if strings.EqualFold(strings.TrimSpace(payload.LeaveMode), "public") {
				c.hub.Broadcast(payload.ChatID, dto.WsEvent{
					Type: "group:member:left",
					Payload: mustMarshal(map[string]any{
						"chat_id":      payload.ChatID,
						"user_id":      c.userID,
						"leave_mode":   payload.LeaveMode,
						"history_mode": payload.HistoryMode,
					}),
				})
			}

		case "group:member:add":
			var payload dto.GroupMemberActionPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu thêm thành viên không hợp lệ")
				continue
			}

			if err := c.groupService.AddMember(c.ctx, payload.ChatID, c.userID, payload.UserID); err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.Broadcast(payload.ChatID, dto.WsEvent{
				Type: "group:member:added",
				Payload: mustMarshal(map[string]any{
					"chat_id": payload.ChatID,
					"user_id": payload.UserID,
					"by":      c.userID,
				}),
			})

		case "group:member:ban":
			var payload dto.GroupBanPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu chặn thành viên không hợp lệ")
				continue
			}

			if err := c.groupService.BanMember(c.ctx, payload.ChatID, c.userID, payload.UserID); err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.Broadcast(payload.ChatID, dto.WsEvent{
				Type: "group:member:banned",
				Payload: mustMarshal(map[string]any{
					"chat_id": payload.ChatID,
					"user_id": payload.UserID,
					"by":      c.userID,
				}),
			})

		case "group:member:mute":
			var payload dto.GroupMutePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu mute không hợp lệ")
				continue
			}

			mute, err := c.groupService.MuteMember(
				c.ctx,
				payload.ChatID,
				c.userID,
				payload.UserID,
				payload.Reason,
				payload.DurationMins,
			)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.Broadcast(payload.ChatID, dto.WsEvent{
				Type:    "group:member:muted",
				Payload: mustMarshal(mute),
			})

		case "group:member:unmute":
			var payload dto.GroupUnmutePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu unmute không hợp lệ")
				continue
			}

			if err := c.groupService.UnmuteMember(c.ctx, payload.ChatID, c.userID, payload.UserID); err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.Broadcast(payload.ChatID, dto.WsEvent{
				Type: "group:member:unmuted",
				Payload: mustMarshal(map[string]any{
					"chat_id": payload.ChatID,
					"user_id": payload.UserID,
					"by":      c.userID,
				}),
			})

		case "group:admin:transfer":
			var payload dto.GroupTransferAdminPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu chuyển quyền không hợp lệ")
				continue
			}

			if err := c.groupService.TransferAdmin(c.ctx, payload.ChatID, c.userID, payload.TargetUserID); err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.Broadcast(payload.ChatID, dto.WsEvent{
				Type: "group:admin:transferred",
				Payload: mustMarshal(map[string]any{
					"chat_id":        payload.ChatID,
					"target_user_id": payload.TargetUserID,
					"by":             c.userID,
				}),
			})

		case "group:settings:update":
			var payload dto.GroupSettingsUpdatePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu settings không hợp lệ")
				continue
			}

			settings, err := c.groupService.UpdateSettings(c.ctx, payload.ChatID, c.userID, &dto.GroupChatSettingsDTO{
				NotificationsEnabled: payload.NotificationsEnabled,
				AllowMemberAdd:       payload.AllowMemberAdd,
				Name:                 payload.Name,
				AvatarURI:            payload.AvatarURI,
			})
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.Broadcast(payload.ChatID, dto.WsEvent{
				Type:    "group:settings:updated",
				Payload: mustMarshal(settings),
			})

		case "group:call:create":
			var payload dto.GroupCallInitiatePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tạo cuộc gọi không hợp lệ")
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, payload.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			memberIDs, err := c.messageService.ListGroupMemberIDs(c.ctx, c.userID, payload.ChatID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			// Chỉ hỗ trợ video: nếu payload không truyền thì mặc định video để giữ tương thích.
			callType := strings.TrimSpace(payload.CallType)
			if callType == "" {
				callType = "video"
			}
			if !strings.EqualFold(callType, "video") {
				c.sendError("group call hiện chỉ hỗ trợ video")
				continue
			}

			session, err := c.hub.CreateGroupCall(payload.ChatID, c.userID, callType, payload.ParticipantIDs, memberIDs)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			c.markCallActive(session.CallID, true)

			createdPayload := map[string]any{
				"call_id":      session.CallID,
				"chat_id":      session.ChatID,
				"caller_id":    session.CallerID,
				"participants": session.ParticipantIDs(),
				"is_video":     true,
			}

			c.sendEvent("group:call:created", createdPayload)
			c.groupChatHub.Broadcast(payload.ChatID, dto.WsEvent{
				Type:    "group:call:incoming",
				Payload: mustMarshal(createdPayload),
			})
			for participantID := range session.Participants {
				if participantID == c.userID {
					continue
				}
				c.hub.SendToUser(participantID, dto.WsEvent{
					Type:    "group:call:incoming",
					Payload: mustMarshal(createdPayload),
				})
			}
			c.sendGroupChatSystemMessage(payload.ChatID, "Cuộc gọi đã được bắt đầu")

		case "group:call:request-join":
			var payload dto.GroupCallJoinRequestPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu xin join không hợp lệ")
				continue
			}

			snapshot, err := c.hub.GetGroupCallSnapshot(payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, snapshot.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			result, err := c.hub.RequestJoinCall(c.userID, payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if result.JoinedNow {
				c.markCallActive(payload.CallID, false)
				c.sendEvent("group:call:joined_ack", map[string]any{
					"call_id": payload.CallID,
					"chat_id": result.ChatID,
				})
				c.hub.SendToUsers(result.JoinedIDs, dto.WsEvent{
					Type: "group:call:joined",
					Payload: mustMarshal(map[string]any{
						"call_id": payload.CallID,
						"chat_id": result.ChatID,
						"user_id": c.userID,
					}),
				})
			} else {
				c.hub.SendToUser(result.CallerID, dto.WsEvent{
					Type: "group:call:join-request",
					Payload: mustMarshal(map[string]any{
						"call_id": payload.CallID,
						"chat_id": result.ChatID,
						"user_id": c.userID,
					}),
				})
				c.sendEvent("group:call:join-request-sent", map[string]any{
					"call_id": payload.CallID,
				})
			}

		case "group:call:approve-join":
			var payload dto.GroupCallApprovePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu duyệt join không hợp lệ")
				continue
			}

			snapshot, err := c.hub.GetGroupCallSnapshot(payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, snapshot.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			result, err := c.hub.ApproveJoinCall(c.userID, payload.UserID, payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.SendToUser(payload.UserID, dto.WsEvent{
				Type: "group:call:joined",
				Payload: mustMarshal(map[string]any{
					"call_id": result.CallID,
					"chat_id": result.ChatID,
					"user_id": payload.UserID,
				}),
			})

			c.hub.SendToUsers(result.ParticipantIDs, dto.WsEvent{
				Type: "group:call:joined",
				Payload: mustMarshal(map[string]any{
					"call_id": result.CallID,
					"chat_id": result.ChatID,
					"user_id": payload.UserID,
				}),
			})

		case "group:call:reject-join":
			var payload dto.GroupCallApprovePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu từ chối join không hợp lệ")
				continue
			}

			snapshot, err := c.hub.GetGroupCallSnapshot(payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, snapshot.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.hub.RejectJoinCall(c.userID, payload.UserID, payload.CallID); err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.SendToUser(payload.UserID, dto.WsEvent{
				Type: "group:call:join-rejected",
				Payload: mustMarshal(map[string]any{
					"call_id": payload.CallID,
					"user_id": payload.UserID,
				}),
			})

		case "group:call:signal":
			var payload dto.GroupCallSignalPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu signal không hợp lệ")
				continue
			}

			chatID, recipients, err := c.hub.RelaySignalTargets(payload.CallID, c.userID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, chatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			for _, participantID := range recipients {
				c.hub.SendToUser(participantID, dto.WsEvent{
					Type: "group:call:signal",
					Payload: mustMarshal(map[string]any{
						"call_id":   payload.CallID,
						"sender_id": c.userID,
						"signal":    payload.Signal,
					}),
				})
			}

		case "group:call:end":
			var payload dto.GroupCallEndPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu kết thúc cuộc gọi không hợp lệ")
				continue
			}

			snapshot, err := c.hub.GetGroupCallSnapshot(payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, snapshot.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			result, err := c.hub.EndCallByUser(c.userID, payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if result.Ended {
				event := dto.WsEvent{
					Type: "group:call:ended",
					Payload: mustMarshal(map[string]any{
						"call_id": payload.CallID,
						"by":      c.userID,
					}),
				}
				c.hub.SendToUsers(result.NotifyUserIDs, event)
				if c.groupChatHub != nil {
					c.groupChatHub.Broadcast(result.ChatID, event)
				}
				c.sendGroupChatSystemMessage(result.ChatID, fmt.Sprintf("Cuộc gọi đã kết thúc bởi %s", c.userID))
			} else {
				event := dto.WsEvent{
					Type: "group:call:left",
					Payload: mustMarshal(map[string]any{
						"call_id": payload.CallID,
						"user_id": c.userID,
					}),
				}
				c.hub.SendToUsers(result.NotifyUserIDs, event)
			}
			c.clearActiveCall()

		case "group:call:toggle-mute":
			var payload dto.GroupCallToggleMutePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu tắt/mở mic không hợp lệ")
				continue
			}

			snapshot, err := c.hub.GetGroupCallSnapshot(payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, snapshot.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			changed, notifyIDs, err := c.hub.SetParticipantMuted(payload.CallID, c.userID, payload.TargetUserID, payload.Muted, true)
			if err != nil {
				c.sendError(err.Error())
				continue
			}
			if !changed {
				continue
			}

			c.hub.SendToUsers(notifyIDs, dto.WsEvent{
				Type: "group:call:mic",
				Payload: mustMarshal(map[string]any{
					"call_id":    payload.CallID,
					"user_id":    payload.TargetUserID,
					"muted":      payload.Muted,
					"changed_by": c.userID,
				}),
			})

		case "group:call:toggle-mic":
			var payload dto.GroupCallToggleMicPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu toggle mic không hợp lệ")
				continue
			}

			snapshot, err := c.hub.GetGroupCallSnapshot(payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, snapshot.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			changed, notifyIDs, err := c.hub.SetParticipantMuted(payload.CallID, c.userID, "", payload.Muted, false)
			if err != nil {
				c.sendError(err.Error())
				continue
			}
			if !changed {
				continue
			}

			c.hub.SendToUsers(notifyIDs, dto.WsEvent{
				Type: "group:call:mic",
				Payload: mustMarshal(map[string]any{
					"call_id":    payload.CallID,
					"user_id":    c.userID,
					"muted":      payload.Muted,
					"changed_by": c.userID,
				}),
			})

		case "group:call:toggle-video":
			var payload dto.GroupCallToggleVideoPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu video không hợp lệ")
				continue
			}

			snapshot, err := c.hub.GetGroupCallSnapshot(payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, snapshot.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			changed, notifyIDs, err := c.hub.SetParticipantVideo(payload.CallID, c.userID, payload.VideoEnabled)
			if err != nil {
				c.sendError(err.Error())
				continue
			}
			if !changed {
				continue
			}

			c.hub.SendToUsers(notifyIDs, dto.WsEvent{
				Type: "group:call:video",
				Payload: mustMarshal(map[string]any{
					"call_id":       payload.CallID,
					"user_id":       c.userID,
					"video_enabled": payload.VideoEnabled,
				}),
			})

		case "group:call:list-participants", "group:call:get-participants":
			var payload dto.GroupCallParticipantsPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu lấy danh sách participant không hợp lệ")
				continue
			}

			snapshot, err := c.hub.GetGroupCallSnapshot(payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if err := c.messageService.JoinRoom(c.ctx, c.userID, snapshot.ChatID); err != nil {
				c.sendError(err.Error())
				continue
			}

			participants, err := c.hub.GetParticipantsSnapshot(payload.CallID, c.userID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			c.sendEvent("group:call:participants", participants)

		default:
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
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
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

func (c *Client) sendEvent(eventType string, payload any) {
	c.send <- mustMarshal(dto.WsEvent{
		Type:    eventType,
		Payload: mustMarshal(payload),
	})
}

func (c *Client) sendError(text string) {
	c.sendEvent("error", map[string]string{"message": text})
}

func mustMarshal(v any) []byte {
	out, _ := json.Marshal(v)
	return out
}

func (c *Client) cleanupCallSession() {
	// Backward-compatible shim: cleanup hiện được xử lý theo toàn bộ session user đang tham gia.
	c.cleanupDisconnectedCallSessions()
}

func (c *Client) markCallActive(callID string, creator bool) {
	c.activeCallID = callID
	c.callCreator = creator
}

func (c *Client) clearActiveCall() {
	c.activeCallID = ""
	c.callCreator = false
}

func (c *Client) cleanupDisconnectedCallSessions() {
	if c.mode != "call" {
		return
	}

	for _, callID := range c.hub.ListCallIDsByUser(c.userID) {
		result, err := c.hub.EndCallByUser(c.userID, callID)
		if err != nil {
			continue
		}

		if result.Ended {
			event := dto.WsEvent{
				Type: "group:call:ended",
				Payload: mustMarshal(map[string]any{
					"call_id": callID,
					"by":      c.userID,
				}),
			}
			c.hub.SendToUsers(result.NotifyUserIDs, event)
			if c.groupChatHub != nil {
				c.groupChatHub.Broadcast(result.ChatID, event)
			}
			c.sendGroupChatSystemMessage(result.ChatID, fmt.Sprintf("Cuộc gọi đã kết thúc bởi %s", c.userID))
			continue
		}

		event := dto.WsEvent{
			Type: "group:call:left",
			Payload: mustMarshal(map[string]any{
				"call_id": callID,
				"user_id": c.userID,
			}),
		}
		c.hub.SendToUsers(result.NotifyUserIDs, event)
	}

	c.clearActiveCall()
}

func (c *Client) sendGroupChatSystemMessage(chatID, content string) {
	if c.messageService == nil {
		return
	}
	_, _ = c.messageService.CreateSystemMessage(c.ctx, chatID, content)
}
