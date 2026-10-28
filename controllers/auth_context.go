package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	errorsapp "linkup/errors"
)

// authUserID lấy userID do AuthMiddleware gắn vào Gin context.
//
// Luôn dùng hàm này thay cho `userID.(string)` trực tiếp: type assertion trần
// sẽ panic khi middleware không gắn giá trị (route thiếu auth, wiring sai), và
// panic trong handler bị gin.Recovery() biến thành 500 với body rỗng — client
// chỉ thấy "HTTP 500" mà không có manh mối nào để debug.
func authUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	if !ok || id == "" {
		return "", false
	}
	return id, true
}

// requireUserID như authUserID nhưng tự trả lỗi JSON khi thiếu userID.
// Trả về ok=false thì handler phải return ngay.
func requireUserID(c *gin.Context) (string, bool) {
	id, ok := authUserID(c)
	if !ok {
		// Thiếu userID ở đây là lỗi wiring phía server (middleware auth phải
		// chạy trước), nên trả 500 — nhưng là 500 JSON đọc được.
		errorsapp.RespondError(c, http.StatusInternalServerError, errorsapp.New(errorsapp.ErrCodeInternal))
		return "", false
	}
	return id, true
}
