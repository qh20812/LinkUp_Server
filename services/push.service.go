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

type ExpoPushMessage struct {
	To       string                 `json:"to"`
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Sound    string                 `json:"sound,omitempty"`
	Priority string                 `json:"priority,omitempty"`
}

type ExpoPushTicket struct {
	Status  string `json:"status"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

type PushService struct {
	client *http.Client
}

func NewPushService() *PushService {
	return &PushService{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *PushService) Send(token string, title, body string, data map[string]interface{}) error {
	msg := ExpoPushMessage{
		To:       token,
		Title:    title,
		Body:     body,
		Data:     data,
		Sound:    "default",
		Priority: "high",
	}

	payload, err := json.Marshal([]ExpoPushMessage{msg})
	if err != nil {
		return fmt.Errorf("marshal push message: %w", err)
	}

	resp, err := s.client.Post(expoPushAPI, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("expo push request failed: %w", err)
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
		return fmt.Errorf("decode expo response: %w", err)
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("expo push error: %s", result.Errors[0].Message)
	}

	for _, ticket := range result.Data {
		if ticket.Status == "error" && ticket.Message == "DeviceNotRegistered" {
			log.Printf("Push token stale: %s", token)
		}
	}

	return nil
}

func (s *PushService) SendBatch(token string, title, body string, data map[string]interface{}) {
	if err := s.Send(token, title, body, data); err != nil {
		log.Printf("Push send failed for token %s: %v", token, err)
	}
}
