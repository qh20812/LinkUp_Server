package controllers

import (
	"context"
	errorsapp "linkup/errors"
	"linkup/dto"
	"linkup/models"
	"linkup/services"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxCreateCommunityBodySize = 45 << 20

func isRequestBodyTooLarge(err error) bool {
	return err != nil && strings.Contains(err.Error(), "request body too large")
}

// respondServiceError maps service-layer errors to HTTP status codes:
// AppError → errorsapp.StatusCode mapping (400 validation, 403 forbidden,
// 404 not found, 409 conflict, 500 internal); plain errors → 500 without
// leaking internal details.
func respondServiceError(c *gin.Context, err error) {
	if appErr, ok := errorsapp.IsAppError(err); ok {
		errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		return
	}
	errorsapp.RespondError(c, http.StatusInternalServerError, errorsapp.New(errorsapp.ErrCodeInternal))
}

// respondUploadError preserves media validation AppErrors (media.* → 400 via
// StatusCode) instead of masking them as community upload failures.
func respondUploadError(c *gin.Context, err error) {
	if appErr, ok := errorsapp.IsAppError(err); ok {
		errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		return
	}
	errorsapp.RespondError(c, http.StatusInternalServerError, errorsapp.New(errorsapp.ErrCodeInternal))
}

// cleanupUploadedMedia deletes already-uploaded media best-effort when a later
// step of community creation fails, so files do not linger on Cloudinary and
// consume storage quota. Failures are logged, never fatal.
func cleanupUploadedMedia(ctx context.Context, mediaService services.MediaService, userID string, mediaIDs []string) {
	if len(mediaIDs) == 0 {
		return
	}
	// Decouple from request cancellation: the client may already be gone.
	ctx = context.WithoutCancel(ctx)
	for _, mediaID := range mediaIDs {
		if err := mediaService.DeleteMedia(ctx, userID, mediaID); err != nil {
			log.Printf("[community] failed to cleanup orphan media %s after failed create: %v", mediaID, err)
		}
	}
}

type CommunityController struct {
	communityService CommunityService
	mediaService     services.MediaService
}

func NewCommunityController(communityService CommunityService, mediaService services.MediaService) *CommunityController {
	return &CommunityController{communityService: communityService, mediaService: mediaService}
}

func (ctrl *CommunityController) CreateCommunity(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCreateCommunityBodySize)
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		if isRequestBodyTooLarge(err) {
			errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeMediaFileTooLarge))
			return
		}
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	description := strings.TrimSpace(c.PostForm("description"))
	autoApprove := c.PostForm("auto_approve") == "true"
	// No silent fallback: unknown values are rejected by service validation
	// (community.PRIVACY_INVALID). Empty means public (model default).
	privacy := models.CommunityPrivacy(strings.ToLower(strings.TrimSpace(c.PostForm("privacy"))))

	// Ảnh cộng đồng đi theo đường auto-approve: không chạy AI moderation,
	// không chờ admin duyệt. Nội dung vi phạm được xử lý hậu kiểm qua report
	// system (xem Phase report cộng đồng). Uploaded media IDs are tracked so a
	// later failure can clean them up.
	var uploadedMediaIDs []string
	avatarURI, avatarMediaID := "", ""
	file, err := c.FormFile("avatar")
	if err == nil && file != nil {
		media, err := ctrl.mediaService.AutoApproveUpload(c.Request.Context(), userID.(string), file)
		if err != nil {
			respondUploadError(c, err)
			return
		}
		avatarURI = media.FileURI
		avatarMediaID = media.ID
		uploadedMediaIDs = append(uploadedMediaIDs, media.ID)
	}

	backgroundURI, backgroundMediaID := "", ""
	bgFile, err := c.FormFile("background")
	if err == nil && bgFile != nil {
		bgMedia, err := ctrl.mediaService.AutoApproveUpload(c.Request.Context(), userID.(string), bgFile)
		if err != nil {
			cleanupUploadedMedia(c.Request.Context(), ctrl.mediaService, userID.(string), uploadedMediaIDs)
			respondUploadError(c, err)
			return
		}
		backgroundURI = bgMedia.FileURI
		backgroundMediaID = bgMedia.ID
		uploadedMediaIDs = append(uploadedMediaIDs, bgMedia.ID)
	}

	result, err := ctrl.communityService.CreateCommunity(c.Request.Context(), userID.(string), name, description, avatarURI, backgroundURI, avatarMediaID, backgroundMediaID, privacy, autoApprove)
	if err != nil {
		cleanupUploadedMedia(c.Request.Context(), ctrl.mediaService, userID.(string), uploadedMediaIDs)
		respondServiceError(c, err)
		return
	}

	resp := gin.H{
		"message":      "Tạo cộng đồng thành công!",
		"community_id": result.Community.ID,
		"community":    result.Community,
		"auto_approve": result.Community.AutoApprove,
		"default_group_chat": gin.H{
			"id":   result.Chat.ID,
			"name": result.Chat.Name,
		},
	}
	if result.InviteCode != "" {
		resp["invite_code"] = result.InviteCode
	}
	c.JSON(http.StatusCreated, resp)
}

func (ctrl *CommunityController) SetCommunityBackground(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}

	communityID := c.Param("communityID")

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCreateCommunityBodySize)
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		if isRequestBodyTooLarge(err) {
			errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeMediaFileTooLarge))
			return
		}
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	file, err := c.FormFile("background")
	if err != nil || file == nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.Newf(errorsapp.ErrCodeCommunityInvalidFormat, map[string]any{
			"detail": "Vui lòng chọn ảnh background",
		}))
		return
	}

	if err := ctrl.communityService.SetCommunityBackground(c.Request.Context(), userID.(string), communityID, file); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật background cộng đồng thành công!",
	})
}

func (ctrl *CommunityController) UpdateCommunity(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}

	communityID := c.Param("communityID")

	var input dto.UpdateCommunityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	if err := ctrl.communityService.UpdateCommunity(c.Request.Context(), userID.(string), communityID, input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật cộng đồng thành công!",
	})
}

func (ctrl *CommunityController) RequestJoin(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityNotFound))
		return
	}

	var input dto.JoinCommunityInput
	c.ShouldBindJSON(&input)

	result, err := ctrl.communityService.RequestJoin(c.Request.Context(), userID, communityID, input.InviteCode, input.InvitationID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	if result.AutoApproved {
		c.JSON(http.StatusOK, gin.H{
			"message": "Tham gia cộng đồng thành công!",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Gửi yêu cầu tham gia cộng đồng thành công!",
			"request_id": result.RequestID,
		})
	}
}

func (ctrl *CommunityController) ListPendingJoinRequests(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	adminID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityNotFound))
		return
	}

	response, err := ctrl.communityService.ListPendingRequests(c.Request.Context(), adminID, communityID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *CommunityController) ApproveJoinRequest(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	adminID := val.(string)

	requestID := c.Param("requestID")
	if requestID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeJoinRequestNotFound))
		return
	}

	if err := ctrl.communityService.ApproveJoinRequest(c.Request.Context(), adminID, requestID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chấp nhận yêu cầu tham gia thành công!"})
}

func (ctrl *CommunityController) RejectJoinRequest(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	adminID := val.(string)

	requestID := c.Param("requestID")
	if requestID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeJoinRequestNotFound))
		return
	}

	if err := ctrl.communityService.RejectJoinRequest(c.Request.Context(), adminID, requestID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Từ chối yêu cầu tham gia thành công!"})
}

func (ctrl *CommunityController) UpdateMemberRole(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	adminID := val.(string)

	communityID := c.Param("communityID")
	memberID := c.Param("memberID")

	var input dto.UpdateMemberRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	if err := ctrl.communityService.UpdateMemberRole(c.Request.Context(), adminID, communityID, memberID, input.Role); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật vai trò thành công!"})
}

func (ctrl *CommunityController) KickMember(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	adminID := val.(string)

	communityID := c.Param("communityID")
	memberID := c.Param("memberID")

	var input dto.KickMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	if err := ctrl.communityService.KickMember(c.Request.Context(), adminID, communityID, memberID, input.Reason); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đuổi thành viên thành công!"})
}

func (ctrl *CommunityController) GetCommunityMembers(c *gin.Context) {
	communityID := c.Param("communityID")

	response, err := ctrl.communityService.GetCommunityMembers(c.Request.Context(), communityID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *CommunityController) LeaveCommunity(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityNotFound))
		return
	}

	var input dto.LeaveCommunityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	if err := ctrl.communityService.LeaveCommunity(c.Request.Context(), userID, communityID, input.Quiet); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rời cộng đồng thành công!"})
}

func (ctrl *CommunityController) TransferOwnership(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	requesterID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityNotFound))
		return
	}

	var input dto.CommunityTransferOwnershipInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	if err := ctrl.communityService.TransferOwnership(c.Request.Context(), requesterID, communityID, input.TargetUserID, input.KeepAdmin); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chuyển quyền sở hữu cộng đồng thành công!"})
}

func (ctrl *CommunityController) CreateInviteCode(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityNotFound))
		return
	}

	var input dto.CreateInviteCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	result, err := ctrl.communityService.CreateInviteCode(c.Request.Context(), userID, communityID, input.MaxUses, input.ExpiresAt)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CommunityController) ListInviteCodes(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityNotFound))
		return
	}

	result, err := ctrl.communityService.ListInviteCodes(c.Request.Context(), userID, communityID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CommunityController) DeactivateInviteCode(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	codeID := c.Param("codeID")
	if codeID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInviteCodeNotFound))
		return
	}

	if err := ctrl.communityService.DeactivateInviteCode(c.Request.Context(), userID, codeID); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vô hiệu hóa mã mời thành công!"})
}

func (ctrl *CommunityController) SendInvitation(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityNotFound))
		return
	}

	var input dto.SendInvitationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	result, err := ctrl.communityService.SendInvitation(c.Request.Context(), userID, communityID, input.InviteeID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CommunityController) ListMyInvitations(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	result, err := ctrl.communityService.ListMyInvitations(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CommunityController) RespondInvitation(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	invitationID := c.Param("invitationID")
	if invitationID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvitationNotFound))
		return
	}

	var input dto.RespondInvitationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	if err := ctrl.communityService.RespondInvitation(c.Request.Context(), userID, invitationID, input.Accept); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Phản hồi lời mời thành công!"})
}

// ── User-facing Community List/Detail ──

func (ctrl *CommunityController) ListCommunities(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	result, err := ctrl.communityService.ListCommunities(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	userID, _ := c.Get("userID")
	if uid, ok := userID.(string); ok && uid != "" {
		for i := range result.Communities {
			result.Communities[i].IsCreator = result.Communities[i].CreatorID == uid
		}
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CommunityController) ListJoinedCommunities(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	result, err := ctrl.communityService.ListJoinedCommunities(c.Request.Context(), userID, keyword, page, pageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CommunityController) ListCreatedCommunities(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}
	userID := val.(string)

	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	result, err := ctrl.communityService.ListCreatedCommunities(c.Request.Context(), userID, keyword, page, pageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CommunityController) GetCommunityDetail(c *gin.Context) {
	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityNotFound))
		return
	}

	// Optional auth — userID may not exist
	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	result, err := ctrl.communityService.GetCommunityDetail(c.Request.Context(), communityID, userIDStr)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CommunityController) GetCommunityPosts(c *gin.Context) {
	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityNotFound))
		return
	}

	// Optional auth
	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	cursorStr := c.Query("cursor")
	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var cursorCreatedAt, cursorID *string
	if cursorStr != "" {
		parts := splitCursor(cursorStr)
		if len(parts) == 2 {
			cursorCreatedAt = &parts[0]
			cursorID = &parts[1]
		} else {
			// Bare UUID (no comma) — look up the post's created_at
			cursorID = &cursorStr
		}
	}

	cursorTime, cursorIDVal := parseCursor(cursorCreatedAt, cursorID)

	posts, err := ctrl.communityService.GetCommunityPosts(c.Request.Context(), communityID, userIDStr, cursorTime, cursorIDVal, limit)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func splitCursor(cursor string) []string {
	// cursor format: "timestamp_nano,post_id"
	for i := len(cursor) - 1; i >= 0; i-- {
		if cursor[i] == ',' {
			return []string{cursor[:i], cursor[i+1:]}
		}
	}
	return nil
}

func parseCursor(createdAtStr, idStr *string) (*string, *string) {
	if createdAtStr == nil || idStr == nil {
		return nil, nil
	}
	return createdAtStr, idStr
}
