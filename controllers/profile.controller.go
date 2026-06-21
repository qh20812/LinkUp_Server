package controllers

import (
	"linkup/dto"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	profileService *services.ProfileService
}

func NewProfileController(profileService *services.ProfileService) *ProfileController {
	return &ProfileController{
		profileService: profileService,
	}
}

func (h *ProfileController) ViewProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	profile, err := h.profileService.ViewProfile(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	respone := dto.ViewProfileResponse{
		DisplayName:                profile.DisplayName,
		PhoneNumber:                profile.PhoneNumber,
		DateOfBirth:                profile.DateOfBirth,
		AvatarURI:                  profile.AvatarURI,
		Bio:                        profile.Bio,
		IsPrivateProfile:           profile.IsPrivateProfile,
		IsPrivatePosts:             profile.IsPrivatePosts,
		AllowStrangerFriendRequest: profile.AllowStrangerFriendRequest,
		UpdatedAt:                  profile.UpdatedAt,
	}

	c.JSON(http.StatusOK, respone)
}
