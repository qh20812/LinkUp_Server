package controllers

import (
	errorsapp "linkup/errors"
	"linkup/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FriendController struct {
	friendService *services.FriendService
}

func NewFriendController(friendService *services.FriendService) *FriendController {
	return &FriendController{
		friendService: friendService,
	}
}

func (ctrl *FriendController) ToggleFriendRequest(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := val.(string)

	response, err := ctrl.friendService.ToggleFriendRequest(c.Request.Context(), userID, targetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *FriendController) GetFriendRequests(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := val.(string)

	response, err := ctrl.friendService.GetFriendRequests(c.Request.Context(), userID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *FriendController) AcceptFriendRequest(c *gin.Context) {
	requestID := c.Param("id")
	if requestID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := val.(string)

	response, err := ctrl.friendService.AcceptFriendRequest(c.Request.Context(), userID, requestID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *FriendController) RejectFriendRequest(c *gin.Context) {
	requestID := c.Param("id")
	if requestID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := val.(string)

	response, err := ctrl.friendService.RejectFriendRequest(c.Request.Context(), userID, requestID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *FriendController) GetFriends(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := val.(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := ctrl.friendService.GetFriends(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *FriendController) Unfriend(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := val.(string)

	response, err := ctrl.friendService.Unfriend(c.Request.Context(), userID, targetUserID)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctrl *FriendController) GetFriendSuggestions(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUnauthorized))
		return
	}
	userID := val.(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := ctrl.friendService.GetFriendSuggestions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
