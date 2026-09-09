package controllers

import (
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/services"
	"linkup/validations"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContributionController struct {
	contributionService *services.ContributionService
}

func NewContributionController(contributionService *services.ContributionService) *ContributionController {
	return &ContributionController{contributionService: contributionService}
}

func (ctrl *ContributionController) GetPolicy(c *gin.Context) {
	communityID := c.Param("communityID")
	if communityID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", val)

	policy, err := ctrl.contributionService.GetPolicy(c.Request.Context(), communityID, userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		} else if err == validations.ErrNotCommunityMember {
			status = http.StatusForbidden
		}
		errorsapp.Respond(c, status, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.PolicyResponse{
		PostWeight:                  policy.PostWeight,
		CommentWeight:               policy.CommentWeight,
		ReactionWeight:              policy.ReactionWeight,
		EventWeight:                 policy.EventWeight,
		TopContributorThreshold:     policy.TopContributorThreshold,
		ModeratorPromotionThreshold: policy.ModeratorPromotionThreshold,
		AutoPromoteEnabled:          policy.AutoPromoteEnabled,
		BadgeEnabled:                policy.BadgeEnabled,
	}})
}

func (ctrl *ContributionController) UpdatePolicy(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	adminID := fmt.Sprintf("%v", val)
	communityID := c.Param("communityID")

	var input dto.UpdatePolicyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.contributionService.UpdatePolicy(c.Request.Context(), adminID, communityID, input); err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		errorsapp.Respond(c, status, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật chính sách đóng góp thành công!"})
}

func (ctrl *ContributionController) GetMyContribution(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", val)
	communityID := c.Param("communityID")

	if err := ctrl.contributionService.RequireMember(c.Request.Context(), communityID, userID); err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		errorsapp.Respond(c, status, err)
		return
	}

	response, err := ctrl.contributionService.GetContributionResponse(c.Request.Context(), communityID, userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		errorsapp.Respond(c, status, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (ctrl *ContributionController) GetUserContribution(c *gin.Context) {
	communityID := c.Param("communityID")
	userID := c.Param("userID")
	if communityID == "" || userID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	response, err := ctrl.contributionService.GetContributionResponse(c.Request.Context(), communityID, userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		errorsapp.Respond(c, status, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (ctrl *ContributionController) GetLeaderboard(c *gin.Context) {
	communityID := c.Param("communityID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 || pageSize < 1 {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	items, err := ctrl.contributionService.GetLeaderboard(c.Request.Context(), communityID, page, pageSize)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		errorsapp.Respond(c, status, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"data":      items,
	})
}

func (ctrl *ContributionController) GetCommunityMembers(c *gin.Context) {
	communityID := c.Param("communityID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 || pageSize < 1 {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	items, err := ctrl.contributionService.GetCommunityMembers(c.Request.Context(), communityID, page, pageSize)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		errorsapp.Respond(c, status, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"data":      items,
	})
}

func (ctrl *ContributionController) CreateChallenge(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	adminID := fmt.Sprintf("%v", val)
	communityID := c.Param("communityID")

	var input dto.CreateChallengeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.contributionService.CreateChallenge(c.Request.Context(), adminID, communityID, input); err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		errorsapp.Respond(c, status, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tạo challenge thành công!"})
}

func (ctrl *ContributionController) GetActiveChallenges(c *gin.Context) {
	communityID := c.Param("communityID")

	challenges, err := ctrl.contributionService.GetActiveChallengeResponses(c.Request.Context(), communityID)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		errorsapp.Respond(c, status, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": challenges})
}

func (ctrl *ContributionController) JoinChallenge(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := fmt.Sprintf("%v", val)
	challengeID := c.Param("challengeID")
	if challengeID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	if err := ctrl.contributionService.JoinChallenge(c.Request.Context(), userID, challengeID); err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			status := errorsapp.StatusCode(appErr.Code)
			if appErr.Code == errorsapp.ErrCodeContribAlreadyJoined || appErr.Code == errorsapp.ErrCodeContribParticipantLimitHit {
				status = http.StatusConflict
			}
			errorsapp.Respond(c, status, appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tham gia challenge thành công!"})
}

func (ctrl *ContributionController) GetChallengeParticipants(c *gin.Context) {
	challengeID := c.Param("challengeID")
	if challengeID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	participants, err := ctrl.contributionService.GetChallengeParticipants(c.Request.Context(), challengeID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": participants})
}
