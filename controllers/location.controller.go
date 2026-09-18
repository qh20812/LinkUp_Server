package controllers

import (
	"net/http"

	"linkup/internal/locations"
	errorsapp "linkup/errors"

	"github.com/gin-gonic/gin"
)

type LocationController struct {
	locationService *locations.Service
}

func NewLocationController(locationService *locations.Service) *LocationController {
	return &LocationController{locationService: locationService}
}

// ListProvinces GET /api/locations/provinces — no auth.
func (h *LocationController) ListProvinces(c *gin.Context) {
	provinces := h.locationService.ListProvinces()
	c.JSON(http.StatusOK, gin.H{"data": provinces})
}

// ListWards GET /api/locations/wards?province_id= — no auth.
func (h *LocationController) ListWards(c *gin.Context) {
	provinceID := c.Query("province_id")
	if provinceID == "" {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}
	wards, err := h.locationService.ListWards(provinceID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wards})
}

// ReverseGeocode POST /api/locations/reverse-geocode — auth.
func (h *LocationController) ReverseGeocode(c *gin.Context) {
	var input struct {
		Lat *float64 `json:"lat"`
		Lng *float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || input.Lat == nil || input.Lng == nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}

	result, err := h.locationService.ReverseGeocode(c.Request.Context(), *input.Lat, *input.Lng)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}