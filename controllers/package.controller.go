package controllers

import (
	"context"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TokenIssuer cấp lại JWT khi claim đổi giữa phiên (vd: role sau subscribe).
// *services.AuthService satisfy interface này.
type TokenIssuer interface {
	ReissueTokensForUser(ctx context.Context, userID, sessionID string) (*dto.TokenResponse, error)
}

type PackageController struct {
	service services.PackageService
	issuer  TokenIssuer
}

func NewPackageController(service services.PackageService, issuer TokenIssuer) *PackageController {
	return &PackageController{service: service, issuer: issuer}
}

func (ctrl *PackageController) GetPackages(c *gin.Context) {
	list, err := ctrl.service.GetPackages()
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (ctrl *PackageController) Subscribe(c *gin.Context) {
	var input dto.SubscribePackageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	userID := c.GetString("userID")
	sub, err := ctrl.service.SubscribePackage(userID, input.PackageID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	// Role DB vừa nâng lên PARTNER — cấp lại token để JWT claim khớp ngay.
	// Best-effort: reissue fail không được làm fail subscribe (client còn fallback refresh).
	var tokens *dto.TokenResponse
	if ctrl.issuer != nil {
		sessionID := c.GetString("sessionID")
		t, terr := ctrl.issuer.ReissueTokensForUser(c.Request.Context(), userID, sessionID)
		if terr != nil {
			log.Printf("[Package] reissue tokens after subscribe failed for user %s: %v", userID, terr)
		} else {
			tokens = t
		}
	}

	body := gin.H{"message": "Đăng ký gói quảng cáo thành công", "data": sub}
	if tokens != nil {
		body["tokens"] = tokens
	}
	c.JSON(http.StatusOK, body)
}

func (ctrl *PackageController) GetMySubscription(c *gin.Context) {
	userID := c.GetString("userID")
	res, err := ctrl.service.GetUserSubscription(userID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusNotFound, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}
