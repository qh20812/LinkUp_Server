package controllers

import (
	errorsapp "linkup/errors"
	"linkup/dto"
	"linkup/models"
	"linkup/services"
	"linkup/validations"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommunityController struct {
	communityService *services.CommunityService
	mediaService     services.MediaService
}

func NewCommunityController(communityService *services.CommunityService, mediaService services.MediaService) *CommunityController {
	return &CommunityController{communityService: communityService, mediaService: mediaService}
}

func (ctrl *CommunityController) CreateCommunity(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	autoApprove := c.PostForm("auto_approve") == "true"

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeCommunityInvalidFormat))
		return
	}

	avatarURI := ""
	file, err := c.FormFile("avatar")
	if err == nil && file != nil {
		src, err := file.Open()
		if err == nil {
			if _, _, err := validations.ValidateImageDimensions(src, validations.DimensionConstraint{
				MinWidth: 200, MinHeight: 200,
				MaxWidth: 2048, MaxHeight: 2048,
				AspectRatio: "1:1",
			}); err != nil {
				src.Close()
				errorsapp.Respond(c, http.StatusBadRequest, err)
				return
			}
			src.Close()
		}

		media, err := ctrl.mediaService.UploadMedia(c.Request.Context(), userID.(string), file)
		if err != nil {
			errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeBackgroundUploadFailed))
			return
		}
		if media.Status == models.MediaStatusRejected {
			errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeBackgroundRejected))
			return
		}
		avatarURI = media.FileURI
	}

	community, groupChat, err := ctrl.communityService.CreateCommunity(c.Request.Context(), userID.(string), name, description, avatarURI, autoApprove)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Tạo cộng đồng thành công!",
		"community_id": community.ID,
		"auto_approve": community.AutoApprove,
		"default_group_chat": gin.H{
			"id":   groupChat.ID,
			"name": groupChat.Name,
		},
	})
}

func (ctrl *CommunityController) SetCommunityBackground(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
		return
	}

	communityID := c.Param("communityID")

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
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
