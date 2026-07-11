package models

import "time"

type GroupChatMemberRequestStatus string

const (
    GroupChatMemberRequestPending  GroupChatMemberRequestStatus = "pending"
    GroupChatMemberRequestApproved GroupChatMemberRequestStatus = "approved"
    GroupChatMemberRequestRejected GroupChatMemberRequestStatus = "rejected"
)

type GroupChatMemberRequest struct {
    ID          string                        `json:"id" gorm:"primaryKey"`
    ChatID      string                        `json:"chat_id" gorm:"index"`
    RequesterID string                        `json:"requester_id" gorm:"index"`
    TargetUserID string                       `json:"target_user_id" gorm:"index"`
    Status      GroupChatMemberRequestStatus   `json:"status" gorm:"index"`
    CreatedAt   time.Time                     `json:"created_at"`
    RespondedAt *time.Time                    `json:"responded_at,omitempty"`
}

func (GroupChatMemberRequest) TableName() string {
    return "group_chat_member_requests"
}