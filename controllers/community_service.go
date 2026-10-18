package controllers

import (
	"context"
	"mime/multipart"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/services"
)

// CommunityService abstracts *services.CommunityService so handlers are
// testable with fakes. The production implementation satisfies it; changing
// the concrete service does not require touching this list unless a handler
// calls a new method.
type CommunityService interface {
	CreateCommunity(ctx context.Context, creatorID, name, description, avatarURI, backgroundURI string, privacy models.CommunityPrivacy, autoApprove bool) (*services.CreateCommunityResult, error)
	SetCommunityBackground(ctx context.Context, userID, communityID string, file *multipart.FileHeader) error
	UpdateCommunity(ctx context.Context, userID, communityID string, input dto.UpdateCommunityInput) error
	RequestJoin(ctx context.Context, userID, communityID, inviteCode, invitationID string) (*dto.JoinResult, error)
	ListPendingRequests(ctx context.Context, adminID, communityID string) (dto.JoinRequestListResponse, error)
	ApproveJoinRequest(ctx context.Context, adminID, requestID string) error
	RejectJoinRequest(ctx context.Context, adminID, requestID string) error
	UpdateMemberRole(ctx context.Context, adminID, communityID, memberID, newRole string) error
	KickMember(ctx context.Context, adminID, communityID, memberID, reason string) error
	GetCommunityMembers(ctx context.Context, communityID string) (dto.CommunityMemberListResponse, error)
	LeaveCommunity(ctx context.Context, userID, communityID string, quiet bool) error
	TransferOwnership(ctx context.Context, requesterID, communityID, targetUserID string, keepAdmin bool) error
	CreateInviteCode(ctx context.Context, adminID, communityID string, maxUses int, expiresAt *time.Time) (*dto.InviteCodeResponse, error)
	ListInviteCodes(ctx context.Context, adminID, communityID string) ([]dto.InviteCodeResponse, error)
	DeactivateInviteCode(ctx context.Context, adminID, codeID string) error
	SendInvitation(ctx context.Context, inviterID, communityID, inviteeID string) (*dto.InvitationItem, error)
	ListMyInvitations(ctx context.Context, userID string) ([]dto.InvitationItem, error)
	RespondInvitation(ctx context.Context, userID, invitationID string, accept bool) error
	ListCommunities(ctx context.Context, keyword string, page, pageSize int) (dto.CommunityListResponse, error)
	ListJoinedCommunities(ctx context.Context, userID, keyword string, page, pageSize int) (dto.CommunityListResponse, error)
	ListCreatedCommunities(ctx context.Context, userID, keyword string, page, pageSize int) (dto.CommunityListResponse, error)
	GetCommunityDetail(ctx context.Context, communityID, userID string) (*dto.CommunityDetailResponse, error)
	GetCommunityPosts(ctx context.Context, communityID, viewerID string, cursorCreatedAt, cursorID *string, limit int) ([]models.Post, error)
}
