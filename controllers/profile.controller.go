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

	response := dto.ViewProfileResponse{
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

	c.JSON(http.StatusOK, response)
}

func (h *ProfileController) EditProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	var input dto.EditProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if input.DisplayName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_name is required"})
		return
	}

	updatedProfile, err := h.profileService.EditProfile(c.Request.Context(), userID.(string), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := dto.EditProfileResponse{
		Message: "Profile updated successfully",
		Data: dto.ViewProfileResponse{
			DisplayName:                updatedProfile.DisplayName,
			PhoneNumber:                updatedProfile.PhoneNumber,
			DateOfBirth:                updatedProfile.DateOfBirth,
			AvatarURI:                  updatedProfile.AvatarURI,
			Bio:                        updatedProfile.Bio,
			IsPrivateProfile:           updatedProfile.IsPrivateProfile,
			IsPrivatePosts:             updatedProfile.IsPrivatePosts,
			AllowStrangerFriendRequest: updatedProfile.AllowStrangerFriendRequest,
			UpdatedAt:                  updatedProfile.UpdatedAt,
		},
	}

	c.JSON(http.StatusOK, response)
}
