package groupws

import (
	"context"
	"encoding/json"
	"fmt"
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
			case "group:call:create", "group:call:request-join", "group:call:approve-join", "group:call:reject-join", "group:call:signal", "group:call:end":
			default:
				c.sendError("event này chỉ dùng trên endpoint call riêng")
				continue
			}
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

			msgs := make([]dto.MessagePayload, 0, len(history))
			for _, m := range history {
				msgs = append(msgs, dto.MessagePayload{
					ID:               m.ID,
					ChatID:           m.ChatID,
					SenderID:         m.SenderID,
					Content:          m.Content,
					EmojiID:          m.EmojiID,
					MediaID:          m.MediaID,
					ReplyToMessageID: m.ReplyToMessageID,
					IsAnonymized:     m.IsAnonymized,
					AnonymousName:    m.AnonymousName,
					CreatedAt:        m.CreatedAt,
				})
			}

			c.sendEvent("group:history", dto.GroupHistoryPayload{
				ChatID:   payload.ChatID,
				Messages: msgs,
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
				payload.ReplyToMessageID,
			)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.Broadcast(payload.ChatID, dto.WsEvent{
				Type: "group:message:new",
				Payload: mustMarshal(dto.MessagePayload{
					ID:               msg.ID,
					ChatID:           msg.ChatID,
					SenderID:         msg.SenderID,
					Content:          msg.Content,
					EmojiID:          msg.EmojiID,
					MediaID:          msg.MediaID,
					ReplyToMessageID: msg.ReplyToMessageID,
					CreatedAt:        msg.CreatedAt,
				}),
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

			memberIDs, err := c.messageService.ListGroupMemberIDs(c.ctx, c.userID, payload.ChatID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			session, err := c.hub.CreateGroupCall(payload.ChatID, c.userID, "video", payload.ParticipantIDs, memberIDs)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			createdPayload := map[string]any{
				"call_id":      session.CallID,
				"chat_id":      session.ChatID,
				"caller_id":    session.CallerID,
				"participants": session.ParticipantIDs(),
				"is_video":     true,
			}

			c.sendEvent("group:call:created", createdPayload)

			// Phát thông báo vào group chat như tin nhắn mới
			if c.groupChatHub != nil {
				c.groupChatHub.Broadcast(payload.ChatID, dto.WsEvent{
					Type: "group:call:incoming",
					Payload: mustMarshal(map[string]any{
						"call_id":      session.CallID,
						"chat_id":      session.ChatID,
						"caller_id":    session.CallerID,
						"participants": session.ParticipantIDs(),
						"is_video":     true,
					}),
				})
			}

			// Gửi tới các người được mời
			for participantID := range session.Participants {
				if participantID == c.userID {
					continue
				}
				c.hub.SendToUser(participantID, dto.WsEvent{
					Type: "group:call:incoming",
					Payload: mustMarshal(map[string]any{
						"call_id":      session.CallID,
						"chat_id":      session.ChatID,
						"caller_id":    session.CallerID,
						"participants": session.ParticipantIDs(),
						"is_video":     true,
					}),
				})
			}

		case "group:call:request-join":
			var payload dto.GroupCallJoinRequestPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu xin join không hợp lệ")
				continue
			}

			session, err := c.hub.RequestJoinCall(c.userID, payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			// Kiểm tra nếu user đã được join (nếu họ là invited thì hub đã thêm vào Joined)
			if _, joined := session.Joined[c.userID]; joined {
				// thông báo cho tất cả participants rằng user đã join
				c.hub.SendToUsers(session.ParticipantIDs(), dto.WsEvent{
					Type: "group:call:joined",
					Payload: mustMarshal(map[string]any{
						"call_id": session.CallID,
						"chat_id": session.ChatID,
						"user_id": c.userID,
					}),
				})
				// ack cho user
				c.sendEvent("group:call:joined_ack", map[string]any{
					"call_id": session.CallID,
					"chat_id": session.ChatID,
				})
			} else {
				// là request -> notify creator
				c.hub.SendToUser(session.CallerID, dto.WsEvent{
					Type: "group:call:join-request",
					Payload: mustMarshal(map[string]any{
						"call_id": session.CallID,
						"chat_id": session.ChatID,
						"user_id": c.userID,
					}),
				})
				c.sendEvent("group:call:join-request-sent", map[string]any{
					"call_id": session.CallID,
				})
			}

		case "group:call:approve-join":
			var payload dto.GroupCallApprovePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu duyệt join không hợp lệ")
				continue
			}

			session, err := c.hub.ApproveJoinCall(c.userID, payload.UserID, payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			c.hub.SendToUser(payload.UserID, dto.WsEvent{
				Type: "group:call:joined",
				Payload: mustMarshal(map[string]any{
					"call_id": session.CallID,
					"chat_id": session.ChatID,
					"user_id": payload.UserID,
				}),
			})

			c.hub.SendToUsers(session.ParticipantIDs(), dto.WsEvent{
				Type: "group:call:joined",
				Payload: mustMarshal(map[string]any{
					"call_id": session.CallID,
					"chat_id": session.ChatID,
					"user_id": payload.UserID,
				}),
			})

		case "group:call:reject-join":
			var payload dto.GroupCallApprovePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.sendError("dữ liệu từ chối join không hợp lệ")
				continue
			}

			_, err := c.hub.RejectJoinCall(c.userID, payload.UserID, payload.CallID)
			if err != nil {
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

			session, err := c.hub.GetGroupCall(payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			for participantID := range session.Participants {
				if participantID == c.userID {
					continue
				}
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

			ended, session, err := c.hub.EndCallByUser(c.userID, payload.CallID)
			if err != nil {
				c.sendError(err.Error())
				continue
			}

			if ended {
				// creator đã kết thúc -> notify tất cả participants cuộc gọi kết thúc
				c.hub.SendToUsers(session.ParticipantIDs(), dto.WsEvent{
					Type: "group:call:ended",
					Payload: mustMarshal(map[string]any{
						"call_id": payload.CallID,
						"by":      c.userID,
					}),
				})
			} else {
				// một user rời cuộc gọi -> notify remaining participants
				c.hub.SendToUsers(session.ParticipantIDs(), dto.WsEvent{
					Type: "group:call:left",
					Payload: mustMarshal(map[string]any{
						"call_id": payload.CallID,
						"user_id": c.userID,
					}),
				})
			}

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
