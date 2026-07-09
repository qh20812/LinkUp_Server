package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	adminService *services.AdminService
}

func NewAdminController(adminService *services.AdminService) *AdminController {
	return &AdminController{adminService: adminService}
}

func (ctrl *AdminController) ListUsers(c *gin.Context) {
	var input dto.AdminUserFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tham số truy vấn không hợp lệ"})
		return
	}

	result, err := ctrl.adminService.ListUsers(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminController) UpdateUserStatus(c *gin.Context) {
	superAdminID := c.GetString("userID")
	if superAdminID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	targetUserID := c.Param("userID")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID không hợp lệ"})
		return
	}

	var input dto.AdminUserUpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	if err := ctrl.adminService.UpdateUserStatus(c.Request.Context(), superAdminID, targetUserID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cập nhật trạng thái người dùng thành công"})
}

func (ctrl *AdminController) BanUser(c *gin.Context) {
	superAdminID := c.GetString("userID")
	if superAdminID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	targetUserID := c.Param("userID")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID không hợp lệ"})
		return
	}

	var input dto.AdminUserBanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu vào không hợp lệ"})
		return
	}

	if err := ctrl.adminService.BanUser(c.Request.Context(), superAdminID, targetUserID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ban user thành công"})
}

func (ctrl *AdminController) ListPosts(c *gin.Context) {
	var input dto.AdminPostFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tham số truy vấn không hợp lệ"})
		return
	}

	superAdminID := c.GetString("userID")
	if superAdminID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	result, err := ctrl.adminService.ListPosts(c.Request.Context(), superAdminID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminController) HidePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "postID không hợp lệ"})
		return
	}

	var input dto.AdminHidePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	superAdminID := c.GetString("userID")
	if superAdminID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	if err := ctrl.adminService.HidePost(c.Request.Context(), superAdminID, postID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ẩn bài viết thành công"})
}

func (ctrl *AdminController) ChangePostStatus(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "postID không hợp lệ"})
		return
	}

	var input dto.AdminUpdatePostStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	superAdminID := c.GetString("userID")
	if superAdminID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	if err := ctrl.adminService.ChangePostStatus(c.Request.Context(), superAdminID, postID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật trạng thái bài viết thành công"})
}

func (ctrl *AdminController) ListReports(c *gin.Context) {
	var input dto.AdminReportFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tham số truy vấn không hợp lệ"})
		return
	}

	superAdminID := c.GetString("userID")
	result, err := ctrl.adminService.ListReports(c.Request.Context(), superAdminID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminController) GetReportDetail(c *gin.Context) {
	reportID := c.Param("reportID")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reportID không hợp lệ"})
		return
	}

	superAdminID := c.GetString("userID")
	result, err := ctrl.adminService.GetReportDetail(c.Request.Context(), superAdminID, reportID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ── Group Chat: Admin handlers ──────────────────────────────────────────────

func (ctrl *AdminController) ListGroups(c *gin.Context) {
	var input dto.AdminGroupFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tham số truy vấn không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	result, err := ctrl.adminService.ListGroups(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminController) GetGroupDetail(c *gin.Context) {
	chatID := c.Param("chatID")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chatID không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	result, err := ctrl.adminService.GetGroupDetail(c.Request.Context(), userID, chatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminController) ListGroupMembers(c *gin.Context) {
	chatID := c.Param("chatID")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chatID không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	result, err := ctrl.adminService.ListGroupMembers(c.Request.Context(), userID, chatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": result})
}

func (ctrl *AdminController) GetGroupModerationLogs(c *gin.Context) {
	chatID := c.Param("chatID")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chatID không hợp lệ"})
		return
	}

	var input dto.AdminGroupFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tham số truy vấn không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	result, err := ctrl.adminService.GetGroupModerationLogs(c.Request.Context(), userID, chatID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminController) HideGroup(c *gin.Context) {
	chatID := c.Param("chatID")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chatID không hợp lệ"})
		return
	}

	var input dto.AdminModerateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp lý do"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.HideGroup(c.Request.Context(), userID, chatID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ẩn group thành công"})
}

func (ctrl *AdminController) UnhideGroup(c *gin.Context) {
	chatID := c.Param("chatID")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chatID không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.UnhideGroup(c.Request.Context(), userID, chatID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bỏ ẩn group thành công"})
}

func (ctrl *AdminController) ArchiveGroup(c *gin.Context) {
	chatID := c.Param("chatID")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chatID không hợp lệ"})
		return
	}

	var input dto.AdminModerateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp lý do"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.ArchiveGroup(c.Request.Context(), userID, chatID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đình chỉ group thành công"})
}

func (ctrl *AdminController) WarnGroup(c *gin.Context) {
	chatID := c.Param("chatID")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chatID không hợp lệ"})
		return
	}

	var input dto.AdminWarnInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp reason và message"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.WarnGroup(c.Request.Context(), userID, chatID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cảnh báo group thành công"})
}

// ── Community: Admin handlers ───────────────────────────────────────────────

func (ctrl *AdminController) ListCommunities(c *gin.Context) {
	var input dto.AdminCommunityFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tham số truy vấn không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	result, err := ctrl.adminService.ListCommunities(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminController) GetCommunityDetail(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	result, err := ctrl.adminService.GetCommunityDetail(c.Request.Context(), userID, communityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminController) ListCommunityMembers(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	result, err := ctrl.adminService.ListCommunityMembers(c.Request.Context(), userID, communityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": result})
}

func (ctrl *AdminController) GetCommunityModerationLogs(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID không hợp lệ"})
		return
	}

	var input dto.AdminGroupFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tham số truy vấn không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	result, err := ctrl.adminService.GetCommunityModerationLogs(c.Request.Context(), userID, communityID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminController) HideCommunity(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID không hợp lệ"})
		return
	}

	var input dto.AdminModerateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp lý do"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.HideCommunity(c.Request.Context(), userID, communityID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ẩn cộng đồng thành công"})
}

func (ctrl *AdminController) UnhideCommunity(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID không hợp lệ"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.UnhideCommunity(c.Request.Context(), userID, communityID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bỏ ẩn cộng đồng thành công"})
}

func (ctrl *AdminController) ArchiveCommunity(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID không hợp lệ"})
		return
	}

	var input dto.AdminModerateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp lý do"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.ArchiveCommunity(c.Request.Context(), userID, communityID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đình chỉ cộng đồng thành công"})
}

func (ctrl *AdminController) WarnCommunity(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID không hợp lệ"})
		return
	}

	var input dto.AdminWarnInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp reason và message"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.WarnCommunity(c.Request.Context(), userID, communityID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cảnh báo cộng đồng thành công"})
}

func (ctrl *AdminController) DeleteGroup(c *gin.Context) {
	chatID := c.Param("chatID")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chatID không hợp lệ"})
		return
	}

	var input dto.AdminModerateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp lý do"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.DeleteGroup(c.Request.Context(), userID, chatID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa group chat thành công"})
}

func (ctrl *AdminController) DeleteCommunity(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID không hợp lệ"})
		return
	}

	var input dto.AdminModerateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp lý do"})
		return
	}

	userID := c.GetString("userID")
	if err := ctrl.adminService.DeleteCommunity(c.Request.Context(), userID, communityID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa cộng đồng thành công"})
}

func (ctrl *AdminController) ReviewReport(c *gin.Context) {
	reportID := c.Param("reportID")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reportID không hợp lệ"})
		return
	}

	var input dto.AdminReportReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	superAdminID := c.GetString("userID")
	if err := ctrl.adminService.ReviewReport(c.Request.Context(), superAdminID, reportID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "xử lý báo cáo thành công"})
}
