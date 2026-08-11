package controllers

import (
	"linkup/dto"
	errorsapp "linkup/errors"
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
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	profile, err := h.profileService.ViewProfile(c.Request.Context(), userID.(string))
	if err != nil {
		errorsapp.Respond(c, http.StatusNotFound, err)
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

func (h *ProfileController) ViewProfileByID(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	viewerID := ""
	if val, exists := c.Get("userID"); exists {
		viewerID = val.(string)
	}

	profile, err := h.profileService.ViewProfileByID(c.Request.Context(), viewerID, targetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusNotFound, err)
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
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}

	var input dto.EditProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if input.DisplayName == nil && input.PhoneNumber == nil &&
		input.DateOfBirth == nil && input.AvatarURI == nil &&
		input.Bio == nil && input.IsPrivateProfile == nil &&
		input.IsPrivatePosts == nil && input.AllowStrangerFriendRequest == nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if input.DisplayName != nil && *input.DisplayName == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	updatedProfile, err := h.profileService.EditProfile(c.Request.Context(), userID.(string), input)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	response := dto.EditProfileResponse{
		Message: "Cập nhật hồ sơ thành công",
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
