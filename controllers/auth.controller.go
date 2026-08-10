package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	"linkup/services"
	"linkup/validations"
)

type AuthController struct {
	authService *services.AuthService
	validation  *validations.AuthValidation
}

func NewAuthController(authService *services.AuthService, validation *validations.AuthValidation) *AuthController {
	return &AuthController{authService: authService, validation: validation}
}

func (h *AuthController) Register(c *gin.Context) {
	var input dto.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	if err := h.validation.ValidateRegisterInput(input.DisplayName, input.Email, input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *AuthController) Login(c *gin.Context) {
	var input dto.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	if err := h.validation.ValidateLoginInput(input.Email, input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceName, ipAddress, userAgent := clientDeviceInfo(c)

	response, err := h.authService.Login(c.Request.Context(), input, deviceName, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthController) GoogleLogin(c *gin.Context) {
	var input dto.GoogleLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}
	if input.IDToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token không được để trống"})
		return
	}

	response, err := h.authService.GoogleLogin(c.Request.Context(), input.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthController) ChangePassword(c *gin.Context) {
	userID := c.GetString("userID") //From JWT middlware
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	var input dto.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ChangePasswordResponse{Message: "đổi mật khẩu thành công"})
}

func (h *AuthController) RefreshToken(c *gin.Context) {
	var input dto.RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu đầu vào không hợp lệ"})
		return
	}

	deviceName, ipAddress, userAgent := clientDeviceInfo(c)

	response, err := h.authService.RefreshToken(c.Request.Context(), input, deviceName, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// clientDeviceInfo extracts device metadata from the request for session tracking.
func clientDeviceInfo(c *gin.Context) (deviceName, ipAddress, userAgent string) {
	userAgent = c.GetHeader("User-Agent")
	ipAddress = c.ClientIP()
	deviceName = "Web"
	if userAgent != "" {
		if len(userAgent) > 64 {
			deviceName = userAgent[:64]
		} else {
			deviceName = userAgent
		}
	}
	return deviceName, ipAddress, userAgent
}

func (h *AuthController) Logout(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "đăng xuất thất bại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đăng xuất thành công"})
}
