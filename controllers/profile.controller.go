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

func (h *ProfileController) ViewProfileByID(c *gin.Context) {
    targetUserID := c.Param("userID")
    if targetUserID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
        return
    }

    viewerID := ""
    if val, exists := c.Get("userID"); exists {
        viewerID = val.(string)
    }

    profile, err := h.profileService.ViewProfileByID(c.Request.Context(), viewerID, targetUserID)
    if err != nil {
        if err.Error() == "this profile is private" {
            c.JSON(http.StatusForbidden, gin.H{"error": "this profile is private"})
            return
        }
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
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }

    var input dto.EditProfileInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    if input.DisplayName == nil && input.PhoneNumber == nil &&
        input.DateOfBirth == nil && input.AvatarURI == nil &&
        input.Bio == nil && input.IsPrivateProfile == nil &&
        input.IsPrivatePosts == nil && input.AllowStrangerFriendRequest == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field is required"})
        return
    }

    if input.DisplayName != nil && *input.DisplayName == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "display_name cannot be empty"})
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