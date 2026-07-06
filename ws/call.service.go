package ws

import (
	"context"
	"encoding/json"
	"linkup/dto"
	"linkup/models"
)

type CallService interface {
	InitiateCall(ctx context.Context, callerID string, payload dto.CallInitiatePayload) (*models.Call, error)
	AcceptCall(ctx context.Context, userID string, callID string) error
	RejectCall(ctx context.Context, userID string, callID string) error
	EndCall(ctx context.Context, userID string, callID string) error
	HandleSignal(ctx context.Context, senderID string, callID string, signal json.RawMessage) error
}
