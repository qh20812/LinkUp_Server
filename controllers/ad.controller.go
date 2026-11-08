package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/internal/locations"
	"linkup/models"
	"linkup/repository"
	"linkup/services"
	"linkup/validations"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AdController struct {
	service       services.AdService
	profileRepo   *repository.ProfileRepository
	locationStore *locations.Store
}

func NewAdController(service services.AdService, profileRepo *repository.ProfileRepository) *AdController {
	return &AdController{service: service, profileRepo: profileRepo}
}

// SetLocationStore cho phép validate target_locations là province ID có thật.
func (ctrl *AdController) SetLocationStore(store *locations.Store) {
	ctrl.locationStore = store
}

// validateTargetLocations chuẩn hoá + kiểm tra danh sách tỉnh/thành.
// target_locations rỗng = chạy toàn quốc. Chỉ nhận province ID trong danh mục
// (34 tỉnh/thành) để khớp profiles.current_province.
func (ctrl *AdController) validateTargetLocations(list []string) error {
	if len(list) == 0 {
		return nil
	}
	if len(list) > 34 {
		return fmt.Errorf("target_locations tối đa 34 tỉnh/thành")
	}

	seen := make(map[string]struct{}, len(list))
	for _, raw := range list {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		if ctrl.locationStore == nil {
			continue
		}
		if _, ok := ctrl.locationStore.ProvinceByID(id); !ok {
			return fmt.Errorf("target_locations: %q không phải tỉnh/thành hợp lệ", id)
		}
	}
	return nil
}

// adMaxMultipartMemory giới hạn phần multipart giữ trong RAM; phần vượt sẽ
// được Go spill ra file tạm. Trước đây đặt 500 MB nên vài upload đồng thời có
// thể làm OOM VPS 2–4 GB.
const adMaxMultipartMemory = 32 << 20

// adTimeLayouts: input datetime-local của trình duyệt gửi "2006-01-02T15:04"
// (không kèm timezone) nên time.RFC3339 không parse được. Trước đây lỗi bị bỏ
// qua im lặng, khiến ngày bắt đầu/kết thúc người dùng chọn bị mất.
var adTimeLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// parseAdTime chuẩn hoá một giá trị thời gian từ client.
// Chuỗi rỗng trả (nil, nil) — trường không được gửi lên.
func parseAdTime(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	for _, layout := range adTimeLayouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("thời gian %q không hợp lệ (định dạng mong đợi: YYYY-MM-DDTHH:mm)", raw)
}

// adPayload là dạng trung gian cho mọi kiểu payload: tất cả giá trị đều phẳng
// và thời gian là string để parseAdTime xử lý thống nhất.
type adPayload struct {
	Title           string   `json:"title"`
	Content         string   `json:"content"`
	Format          string   `json:"format"`
	TargetURL       string   `json:"target_url"`
	Budget          float64  `json:"budget"`
	DailyBudget     float64  `json:"daily_budget"`
	CPMPrice        float64  `json:"cpm_price"`
	CPCPrice        float64  `json:"cpc_price"`
	MaxImpressions  int      `json:"max_impressions"`
	StartedAt       string   `json:"started_at"`
	ExpiresAt       string   `json:"expires_at"`
	TargetGender    string   `json:"target_gender"`
	TargetAgeMin    int      `json:"target_age_min"`
	TargetAgeMax    int      `json:"target_age_max"`
	TargetLocations []string `json:"target_locations"`
}

// decodeAdPayload nhận 3 kiểu payload đang tồn tại trên thực tế:
//  1. JSON thuần (application/json)
//  2. multipart/form-data với một field `data` = JSON (web hiện tại)
//  3. multipart/form-data với từng field phẳng (client cũ / mobile)
func decodeAdPayload(c *gin.Context) (adPayload, error) {
	var p adPayload

	if strings.Contains(c.ContentType(), "application/json") {
		if err := c.ShouldBindJSON(&p); err != nil {
			return p, err
		}
		return p, nil
	}

	// Body không phải multipart vẫn cho đi tiếp: PostForm sẽ trả chuỗi rỗng và
	// phần validate phía dưới báo lỗi trường thiếu, dễ hiểu hơn lỗi parse.
	if err := c.Request.ParseMultipartForm(adMaxMultipartMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return p, err
	}

	if raw := c.PostForm("data"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return p, fmt.Errorf("dữ liệu quảng cáo (data) không phải JSON hợp lệ: %w", err)
		}
		return p, nil
	}

	p.Title = c.PostForm("title")
	p.Content = c.PostForm("content")
	p.TargetURL = c.PostForm("target_url")
	p.Format = c.PostForm("format")
	p.Budget, _ = strconv.ParseFloat(c.PostForm("budget"), 64)
	p.DailyBudget, _ = strconv.ParseFloat(c.PostForm("daily_budget"), 64)
	p.CPMPrice, _ = strconv.ParseFloat(c.PostForm("cpm_price"), 64)
	p.CPCPrice, _ = strconv.ParseFloat(c.PostForm("cpc_price"), 64)
	p.MaxImpressions, _ = strconv.Atoi(c.PostForm("max_impressions"))
	p.StartedAt = c.PostForm("started_at")
	p.ExpiresAt = c.PostForm("expires_at")
	p.TargetGender = c.PostForm("target_gender")
	p.TargetAgeMin, _ = strconv.Atoi(c.PostForm("target_age_min"))
	p.TargetAgeMax, _ = strconv.Atoi(c.PostForm("target_age_max"))
	if locStr := c.PostForm("target_locations"); locStr != "" {
		if err := json.Unmarshal([]byte(locStr), &p.TargetLocations); err != nil {
			return p, fmt.Errorf("target_locations phải là mảng JSON: %w", err)
		}
	}

	return p, nil
}

func (p adPayload) toCreateInput() (dto.CreateAdInput, error) {
	startedAt, err := parseAdTime(p.StartedAt)
	if err != nil {
		return dto.CreateAdInput{}, err
	}
	expiresAt, err := parseAdTime(p.ExpiresAt)
	if err != nil {
		return dto.CreateAdInput{}, err
	}

	format := strings.TrimSpace(p.Format)
	if format == "" {
		format = "image"
	}

	return dto.CreateAdInput{
		Title:           p.Title,
		Content:         p.Content,
		TargetURL:       strings.TrimSpace(p.TargetURL),
		Format:          format,
		Budget:          p.Budget,
		DailyBudget:     p.DailyBudget,
		CPMPrice:        p.CPMPrice,
		CPCPrice:        p.CPCPrice,
		MaxImpressions:  p.MaxImpressions,
		StartedAt:       startedAt,
		ExpiresAt:       expiresAt,
		TargetGender:    p.TargetGender,
		TargetAgeMin:    p.TargetAgeMin,
		TargetAgeMax:    p.TargetAgeMax,
		TargetLocations: p.TargetLocations,
	}, nil
}

func (p adPayload) toUpdateInput() (dto.UpdateAdInput, error) {
	startedAt, err := parseAdTime(p.StartedAt)
	if err != nil {
		return dto.UpdateAdInput{}, err
	}
	expiresAt, err := parseAdTime(p.ExpiresAt)
	if err != nil {
		return dto.UpdateAdInput{}, err
	}

	return dto.UpdateAdInput{
		Title:           p.Title,
		Content:         p.Content,
		TargetURL:       strings.TrimSpace(p.TargetURL),
		Budget:          p.Budget,
		DailyBudget:     p.DailyBudget,
		CPMPrice:        p.CPMPrice,
		CPCPrice:        p.CPCPrice,
		MaxImpressions:  p.MaxImpressions,
		StartedAt:       startedAt,
		ExpiresAt:       expiresAt,
		TargetGender:    p.TargetGender,
		TargetAgeMin:    p.TargetAgeMin,
		TargetAgeMax:    p.TargetAgeMax,
		TargetLocations: p.TargetLocations,
	}, nil
}

// adMediaFiles trả danh sách file đính kèm; request JSON không có multipart
// form nên phải guard để tránh nil pointer.
func adMediaFiles(c *gin.Context) []*multipart.FileHeader {
	if c.Request.MultipartForm == nil {
		return nil
	}
	return c.Request.MultipartForm.File["media"]
}

func (ctrl *AdController) CreateAd(c *gin.Context) {
	payload, err := decodeAdPayload(c)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	input, err := payload.toCreateInput()
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	if err := validations.ValidateCreateAd(input.Title, input.Content, input.TargetURL, input.Format, input.Budget); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	if err := validations.ValidateTargeting(input.TargetGender, input.TargetAgeMin, input.TargetAgeMax); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	if err := ctrl.validateTargetLocations(input.TargetLocations); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	// Kiểm tra media trước khi service giữ slot: sai số lượng/loại/dung lượng
	// thì trả lỗi ngay, không upload lên Cloudinary.
	mediaFiles := adMediaFiles(c)
	if err := validations.ValidateAdMedia(mediaFiles, input.Format); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	currentUserID := c.GetString("userID")
	role := c.GetString("userRole")
	platformAdmin := role == string(models.RoleSuperAdmin) || role == string(models.RoleAdmin)

	ad, err := ctrl.service.CreateAdWithMedia(c, input, mediaFiles, currentUserID, platformAdmin)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}

	// Quảng cáo tạo xong là active ngay (bỏ bước duyệt pending).
	c.JSON(http.StatusCreated, gin.H{"message": "Tạo quảng cáo thành công", "data": ad})
}

func (ctrl *AdController) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	var input dto.UpdateAdStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	ad, err := ctrl.service.UpdateStatus(id, input.Status, partnerID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully", "data": ad})
}

// parseAdRange đọc query from/to (RFC3339 hoặc YYYY-MM-DD). Không truyền gì →
// nil, nil (analytics toàn đời). Chỉ truyền một đầu → mặc định 30 ngày/now.
func parseAdRange(c *gin.Context) (from, to *time.Time, err error) {
	fromRaw := strings.TrimSpace(c.Query("from"))
	toRaw := strings.TrimSpace(c.Query("to"))
	if fromRaw == "" && toRaw == "" {
		return nil, nil, nil
	}

	parse := func(raw string) (*time.Time, error) {
		for _, layout := range []string{time.RFC3339, "2006-01-02"} {
			if t, parseErr := time.ParseInLocation(layout, raw, time.Local); parseErr == nil {
				return &t, nil
			}
		}
		return nil, fmt.Errorf("thời gian %q không hợp lệ (YYYY-MM-DD hoặc RFC3339)", raw)
	}

	now := time.Now()
	if fromRaw != "" {
		if from, err = parse(fromRaw); err != nil {
			return nil, nil, err
		}
	} else {
		d := now.AddDate(0, 0, -30)
		from = &d
	}
	if toRaw != "" {
		if to, err = parse(toRaw); err != nil {
			return nil, nil, err
		}
	} else {
		to = &now
	}
	return from, to, nil
}

func (ctrl *AdController) GetAnalytics(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")
	role := c.GetString("userRole")
	isAdmin := role == string(models.RoleSuperAdmin) || role == string(models.RoleAdmin)

	from, to, err := parseAdRange(c)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	report, err := ctrl.service.GetAdPerformance(id, partnerID, isAdmin, from, to)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
		errorsapp.Respond(c, http.StatusForbidden, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}

func (ctrl *AdController) GetAdminList(c *gin.Context) {
	var input dto.AdListInput
	if err := c.ShouldBindQuery(&input); err != nil {
		errorsapp.RespondError(c, http.StatusBadRequest, errorsapp.New(errorsapp.ErrCodeInvalidInput))
		return
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	role := c.GetString("userRole")
	currentUserID := c.GetString("userID")

	if role == string(models.RoleSuperAdmin) || role == string(models.RoleAdmin) {
		// Một nguồn duy nhất cho admin là /api/admin/ads. Endpoint này giữ lại
		// cho partner; admin gọi vào đây sẽ bị hướng dẫn sang endpoint đúng
		// thay vì trả danh sách trùng shape.
		errorsapp.Respond(c, http.StatusBadRequest, fmt.Errorf("endpoint này dành cho partner; admin dùng GET /api/admin/ads"))
		return
	}

	ads, total, err := ctrl.service.GetPartnerAds(currentUserID, input.Page, input.PageSize, input.Keyword, input.Status)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      ads,
		"total":     total,
		"page":      input.Page,
		"page_size": input.PageSize,
	})
}

func (ctrl *AdController) GetUserFeed(c *gin.Context) {
	userID := c.GetString("userID")

	gender, age, location := "all", 0, ""
	if userID != "" && ctrl.profileRepo != nil {
		demo, err := ctrl.profileRepo.GetUserDemographics(userID)
		if err == nil && demo != nil {
			gender = demo.Gender
			age = demo.Age
			location = demo.Location
		}
	}

	ads, err := ctrl.service.GetAdsForUserFeed(userID, gender, age, location)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toFeedAdItems(ads)})
}

// toFeedAdItems map model nội bộ sang DTO công khai: chỉ giữ field cần render
// quảng cáo, không trả budget/total_spent/CPM/CPC/partner_id/package_id.
func toFeedAdItems(ads []models.Ad) []dto.FeedAdItem {
	items := make([]dto.FeedAdItem, 0, len(ads))
	for _, ad := range ads {
		item := dto.FeedAdItem{
			ID:        ad.ID,
			Title:     ad.Title,
			Content:   ad.Content,
			Format:    string(ad.Format),
			TargetURL: ad.TargetURL,
		}
		for _, media := range ad.MediaList {
			item.MediaList = append(item.MediaList, dto.FeedAdMedia{
				URL:       media.URL,
				MediaType: media.MediaType,
			})
		}
		items = append(items, item)
	}
	return items
}

func (ctrl *AdController) TrackAction(c *gin.Context) {
	adID := c.Param("id")
	var input dto.TrackActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	var userIDPtr *string
	if uid, exists := c.Get("userID"); exists {
		if strUID, ok := uid.(string); ok {
			userIDPtr = &strUID
		}
	}

	ipAddress := c.ClientIP()

	// Rate limit theo user (fallback IP): 60 event/phút. Trước đây endpoint mở
	// hoàn toàn nên có thể spam impression/click để bòn ngân sách quảng cáo.
	limitKey := ipAddress
	if userIDPtr != nil && *userIDPtr != "" {
		limitKey = *userIDPtr
	}
	if !adTrackLimiter.allow(limitKey, time.Now()) {
		log.Printf("ad track: rate limited %s (action=%s ad=%s)", limitKey, input.ActionType, adID)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"status":   "rate_limited",
			"recorded": false,
		})
		return
	}

	recorded, err := ctrl.service.TrackUserAction(adID, userIDPtr, input.ActionType, ipAddress)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
		errorsapp.Respond(c, http.StatusNotFound, err)
		}
		return
	}

	status := "skipped"
	if recorded {
		status = "recorded"
	}
	c.JSON(http.StatusOK, gin.H{"status": status, "recorded": recorded})
}

func (ctrl *AdController) UpdateAd(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	payload, err := decodeAdPayload(c)
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	input, err := payload.toUpdateInput()
	if err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}

	if err := validations.ValidateUpdateAd(input.Title, input.Content, input.TargetURL, input.Budget); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	if err := validations.ValidateTargeting(input.TargetGender, input.TargetAgeMin, input.TargetAgeMax); err != nil {
		errorsapp.Respond(c, http.StatusBadRequest, err)
		return
	}
	if input.TargetLocations != nil {
		if err := ctrl.validateTargetLocations(input.TargetLocations); err != nil {
			errorsapp.Respond(c, http.StatusBadRequest, err)
			return
		}
	}

	ad, err := ctrl.service.UpdateAd(id, partnerID, input)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật quảng cáo thành công", "data": ad})
}

// SubmitAd đưa quảng cáo bị từ chối trở lại hàng đợi duyệt.
func (ctrl *AdController) SubmitAd(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	ad, err := ctrl.service.SubmitForReview(id, partnerID)
	if err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi quảng cáo để duyệt lại", "data": ad})
}

func (ctrl *AdController) DeleteAd(c *gin.Context) {
	id := c.Param("id")
	partnerID := c.GetString("userID")

	if err := ctrl.service.DeleteAd(id, partnerID); err != nil {
		if appErr, ok := errorsapp.IsAppError(err); ok {
			errorsapp.Respond(c, errorsapp.StatusCode(appErr.Code), appErr)
		} else {
			errorsapp.Respond(c, http.StatusBadRequest, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Xóa quảng cáo thành công"})
}

func (ctrl *AdController) GetOverview(c *gin.Context) {
	partnerID := c.GetString("userID")

	overview, err := ctrl.service.GetOverview(partnerID)
	if err != nil {
		errorsapp.Respond(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": overview})
}
