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

func (ctrl *AdminController) GetDashboardAnalytics(c *gin.Context) {
	// Lấy ID Admin từ Middleware Auth
	superAdminID := c.GetString("userID")
	if superAdminID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	var input dto.AdminAnalyticsFilterInput
	// Bind query params
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tham số truy vấn bộ lọc không hợp lệ"})
		return
	}

	result, err := ctrl.adminService.GetDashboardAnalytics(c.Request.Context(), superAdminID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
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
