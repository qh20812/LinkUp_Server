package controllers

import (
	"fmt"
	"linkup/dto"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID là bắt buộc"})
		return
	}

	policy, err := ctrl.contributionService.GetPolicy(c.Request.Context(), communityID)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	adminID := fmt.Sprintf("%v", val)
	communityID := c.Param("communityID")

	var input dto.UpdatePolicyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	if err := ctrl.contributionService.UpdatePolicy(c.Request.Context(), adminID, communityID, input); err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật chính sách đóng góp thành công!"})
}

func (ctrl *ContributionController) GetMyContribution(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	userID := fmt.Sprintf("%v", val)
	communityID := c.Param("communityID")

	response, err := ctrl.contributionService.GetContributionResponse(c.Request.Context(), communityID, userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (ctrl *ContributionController) GetUserContribution(c *gin.Context) {
	communityID := c.Param("communityID")
	userID := c.Param("userID")
	if communityID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "communityID và userID là bắt buộc"})
		return
	}

	response, err := ctrl.contributionService.GetContributionResponse(c.Request.Context(), communityID, userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (ctrl *ContributionController) GetLeaderboard(c *gin.Context) {
	communityID := c.Param("communityID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	items, err := ctrl.contributionService.GetLeaderboard(c.Request.Context(), communityID, page, pageSize)
	if err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	adminID := fmt.Sprintf("%v", val)
	communityID := c.Param("communityID")

	var input dto.CreateChallengeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu đầu vào không hợp lệ"})
		return
	}

	if err := ctrl.contributionService.CreateChallenge(c.Request.Context(), adminID, communityID, input); err != nil {
		status := http.StatusBadRequest
		if err == validations.ErrCommunityNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
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
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": challenges})
}

func (ctrl *ContributionController) JoinChallenge(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin chứng thực người dùng"})
		return
	}
	userID := fmt.Sprintf("%v", val)
	challengeID := c.Param("challengeID")
	if challengeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "challengeID là bắt buộc"})
		return
	}

	if err := ctrl.contributionService.JoinChallenge(c.Request.Context(), userID, challengeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tham gia challenge thành công!"})
}

func (ctrl *ContributionController) GetChallengeParticipants(c *gin.Context) {
	challengeID := c.Param("challengeID")
	if challengeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "challengeID là bắt buộc"})
		return
	}

	participants, err := ctrl.contributionService.GetChallengeParticipants(c.Request.Context(), challengeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": participants})
}
