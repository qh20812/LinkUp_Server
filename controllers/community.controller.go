package controllers

import (
	"linkup/dto"
	"linkup/services"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	avatarURI := ""
	file, err := c.FormFile("avatar")
	if err == nil && file != nil {
		media, err := ctrl.mediaService.UploadMedia(c.Request.Context(), userID.(string), file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tải ảnh đại diện thất bại"})
			return
		}
		avatarURI = media.FileURI
	}

	community, err := ctrl.communityService.CreateCommunity(c.Request.Context(), userID.(string), name, description, avatarURI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Tạo cộng đồng thành công!",
		"community_id": community.ID,
	})
}

func (ctrl *CommunityController) SetCommunityBackground(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}

	communityID := c.Param("communityID")

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	file, err := c.FormFile("background")
	if err != nil || file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng chọn ảnh background"})
		return
	}

	if err := ctrl.communityService.SetCommunityBackground(c.Request.Context(), userID.(string), communityID, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật background cộng đồng thành công!",
	})
}

func (ctrl *CommunityController) RequestJoin(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	userID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID là bắt buộc"})
		return
	}

	requestID, err := ctrl.communityService.RequestJoin(c.Request.Context(), userID, communityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Gửi yêu cầu tham gia cộng đồng thành công!",
		"request_id": requestID,
	})
}

func (ctrl *CommunityController) ListPendingJoinRequests(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	adminID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID là bắt buộc"})
		return
	}

	response, err := ctrl.communityService.ListPendingRequests(c.Request.Context(), adminID, communityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *CommunityController) ApproveJoinRequest(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	adminID := val.(string)

	requestID := c.Param("requestID")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requestID là bắt buộc"})
		return
	}

	if err := ctrl.communityService.ApproveJoinRequest(c.Request.Context(), adminID, requestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chấp nhận yêu cầu tham gia thành công!"})
}

func (ctrl *CommunityController) RejectJoinRequest(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	adminID := val.(string)

	requestID := c.Param("requestID")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requestID là bắt buộc"})
		return
	}

	if err := ctrl.communityService.RejectJoinRequest(c.Request.Context(), adminID, requestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Từ chối yêu cầu tham gia thành công!"})
}

func (ctrl *CommunityController) UpdateMemberRole(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	adminID := val.(string)

	communityID := c.Param("communityID")
	memberID := c.Param("memberID")

	var input dto.UpdateMemberRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	if err := ctrl.communityService.UpdateMemberRole(c.Request.Context(), adminID, communityID, memberID, input.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật vai trò thành công!"})
}

func (ctrl *CommunityController) GetCommunityMembers(c *gin.Context) {
	communityID := c.Param("communityID")

	response, err := ctrl.communityService.GetCommunityMembers(c.Request.Context(), communityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *CommunityController) LeaveCommunity(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	userID := val.(string)

	communityID := c.Param("communityID")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID là bắt buộc"})
		return
	}

	var input dto.LeaveCommunityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	if err := ctrl.communityService.LeaveCommunity(c.Request.Context(), userID, communityID, input.Quiet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rời cộng đồng thành công!"})
}
