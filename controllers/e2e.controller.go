package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
)

type E2EController struct {
	service *services.E2EService
}

func NewE2EController(service *services.E2EService) *E2EController {
	return &E2EController{service: service}
}

// RegisterUserKey stores the user's ECDH public key for end-to-end encrypted
// direct chats. The private key never leaves the client.
func (h *E2EController) RegisterUserKey(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.RegisterE2EKeyRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := h.service.RegisterUserKey(c.Request.Context(), userID, input.PublicKey); err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "khóa E2E đã được đăng ký"})
}

// GetUserKey returns another user's public key so the client can derive the
// shared secret and wrap the chat key for them.
func (h *E2EController) GetUserKey(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	targetID := c.Param("userID")
	if targetID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	key, err := h.service.GetUserKey(c.Request.Context(), targetID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	if key == nil {
		errorsapp.RespondError(c, http.StatusNotFound, errorsapp.New(errorsapp.ErrCodeUserNotFound))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     key.UserID,
		"public_key":  key.PublicKey,
		"key_version": key.KeyVersion,
	})
}

// RekeyChat ghi đè khóa bọc của chính user sau khi đối phương đổi identity.
func (h *E2EController) RekeyChat(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	chatID := c.Param("chatID")
	if chatID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	var input dto.RekeyChatKeyRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := h.service.RekeyChat(c.Request.Context(), userID, chatID, input.Nonce, input.WrappedKey); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "khóa chat đã được cập nhật"})
}

// StoreChatKeys persists the wrapped chat keys for each participant of the
// listed chats.
func (h *E2EController) StoreChatKeys(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.ChatE2EKeyBatchRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := h.service.StoreChatKeys(c.Request.Context(), userID, input.Keys); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "khóa chat đã được lưu"})
}

// GetChatKey returns the caller's wrapped copy of a chat's symmetric key.
func (h *E2EController) GetChatKey(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	chatID := c.Param("chatID")
	if chatID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	key, err := h.service.GetChatKey(c.Request.Context(), userID, chatID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	if key == nil {
		errorsapp.RespondError(c, http.StatusNotFound, errorsapp.New(errorsapp.ErrCodeChatKeyNotFound))
		return
	}

	c.JSON(http.StatusOK, key)
}

// PutRecovery lưu backup khóa chat (blob mã hóa + salt + hash check) để khôi
// phục lịch sử tin nhắn trên thiết bị mới.
func (h *E2EController) PutRecovery(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.PutRecoveryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := h.service.PutRecovery(c.Request.Context(), userID, input); err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã lưu dữ liệu khôi phục"})
}

// GetRecoveryMeta trả về trạng thái backup khôi phục của user (đã có blob chưa
// + salt cần để dẫn khóa + thời điểm cập nhật cuối).
func (h *E2EController) GetRecoveryMeta(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	meta, err := h.service.GetRecoveryMeta(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, meta)
}

// UnlockRecovery xác thực giá trị check (PIN hoặc recovery key) và trả blob
// backup cho client tự giải mã. Nhập sai nhiều lần → khóa tạm thời.
func (h *E2EController) UnlockRecovery(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.UnlockRecoveryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	blob, err := h.service.UnlockRecovery(c.Request.Context(), userID, input.Check)
	if err != nil {
		errorsapp.Respond(c, http.StatusForbidden, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"blob": blob})
}

// DeleteRecovery xóa backup khôi phục (user tắt tính năng khôi phục).
func (h *E2EController) DeleteRecovery(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	if err := h.service.DeleteRecovery(c.Request.Context(), userID); err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã xóa dữ liệu khôi phục"})
}
