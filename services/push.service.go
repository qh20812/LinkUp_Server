package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const expoPushAPI = "https://exp.host/--/api/v2/push/send"

// ExpoPushMessage — payload theo Expo Push API (docs.expo.dev/push-notifications/sending-notifications).
// Badge/RichContent/ThreadID là các field "rich": badge iOS, ảnh đại diện
// (Android hiện out-of-box, iOS cần Notification Service Extension), grouping
// theo hội thoại trên iOS (aps.thread-id).
type ExpoPushMessage struct {
	To          string                 `json:"to"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Sound       string                 `json:"sound,omitempty"`
	Priority    string                 `json:"priority,omitempty"`
	Badge       *int                   `json:"badge,omitempty"`
	RichContent map[string]string      `json:"richContent,omitempty"`
	ThreadID    string                 `json:"threadId,omitempty"`
}

type ExpoPushTicket struct {
	Status  string                 `json:"status"`
	ID      string                 `json:"id,omitempty"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// PushOptions — nội dung push đã được biên soạn (title = tên người gửi,
// body = chi tiết, badge = số chưa đọc, image = avatar).
type PushOptions struct {
	Title    string
	Body     string
	Data     map[string]interface{}
	Badge    *int
	ImageURL string
	ThreadID string
}

type PushService struct {
	client *http.Client
}

func NewPushService() *PushService {
	return &PushService{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send trả về stale=true khi Expo báo token DeviceNotRegistered
// (details.error trên ticket) — caller nên xóa token khỏi DB.
func (s *PushService) Send(token string, opts PushOptions) (bool, error) {
	msg := ExpoPushMessage{
		To:       token,
		Title:    opts.Title,
		Body:     opts.Body,
		Data:     opts.Data,
		Sound:    "default",
		Priority: "high",
		Badge:    opts.Badge,
		ThreadID: opts.ThreadID,
	}
	if opts.ImageURL != "" {
		msg.RichContent = map[string]string{"image": opts.ImageURL}
	}

	payload, err := json.Marshal([]ExpoPushMessage{msg})
	if err != nil {
		return false, fmt.Errorf("marshal push message: %w", err)
	}

	resp, err := s.client.Post(expoPushAPI, "application/json", bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("expo push request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data   []ExpoPushTicket `json:"data"`
		Errors []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode expo response: %w", err)
	}

	if len(result.Errors) > 0 {
		return false, fmt.Errorf("expo push error: %s", result.Errors[0].Message)
	}

	stale := false
	for _, ticket := range result.Data {
		if ticket.Status != "error" {
			continue
		}
		errCode, _ := ticket.Details["error"].(string)
		if errCode == "DeviceNotRegistered" {
			stale = true
			log.Printf("Push token stale: %s", token)
			continue
		}
		log.Printf("Push ticket error for token %s: %s (%s)", token, ticket.Message, errCode)
	}

	return stale, nil
}

// SendBatch gửi 1 token; trả về true nếu token đã chết (DeviceNotRegistered).
func (s *PushService) SendBatch(token string, opts PushOptions) bool {
	stale, err := s.Send(token, opts)
	if err != nil {
		log.Printf("Push send failed for token %s: %v", token, err)
	}
	return stale
}
