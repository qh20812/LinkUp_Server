# ADS Plan — Refactor & Feature Completion

---

## Mục lục

1. [Tổng quan issues](#1-tổng-quan-issues)
2. [Thiết kế gói quảng cáo (Ad Package)](#2-thiết-kế-gói-quảng-cáo)
3. [Hình thức & định dạng quảng cáo](#3-hình-thức-quảng-cáo)
4. [Multi media / Upload từ máy](#4-multi-media--upload)
5. [Analytics có nghĩa](#5-analytics-có-nghĩa)
6. [Bảo mật: CheckAdOwnership](#6-bảo-mật-checkadownership)
7. [Kế hoạch triển khai theo phase](#7-kế-hoạch-triển-khai)

---

## 1. Tổng quan issues

### 1.1 Thiếu gói quảng cáo (Package)
- `Ad` chỉ có `Title` + `TargetURL` — không có khái niệm "gói"
- Partner không biết mua gì, không có link thanh toán/nâng cấp
- Không tracking subscription, không hạn mức

### 1.2 Thiếu định dạng & target
- `Ad` không có trường `format` — không phân biệt image/video/carousel
- `StartedAt`/`ExpiresAt` thuần datetime — không có duration type (ngày/tuần/tháng)
- Không có reach target, audience targeting

### 1.3 Single media, không upload
- `MediaID *string` — chỉ 1 media tùy chọn
- `CreateAd` dùng `ShouldBindJSON` — không multipart, không upload file
- Không cho phép chọn từ media library đã upload

### 1.4 Analytics vô nghĩa
- `action_type` = `impression` | `click` | `interact` — không rõ interact là gì
- Chỉ trả raw count, thiếu context: time range, daily breakdown
- Thiếu cost metrics: CPC, CPM, total_spent, remaining_budget
- Không phân biệt unique vs total

### 1.5 Lỗi bảo mật
- `CheckAdOwnership` định nghĩa ở `rbac.middleware.go:100` nhưng **không wire** vào route nào
- Partner hiện tại có thể `PATCH /ads-management/:id/status` ad của partner khác

---

## 2. Thiết kế gói quảng cáo

### 2.1 Mô hình: Hybrid — Monthly Subscription + Slot-based

```
Partner mua gói tháng → được N slot quảng cáo.
Mỗi slot = 1 chiến dịch (Ad) có budget riêng.
Hết slot → không tạo ad mới được → prompt "Nâng cấp gói".
```

**Lý do chọn hybrid:**
- Partner chỉ trả tiền khi có nhu cầu (không thu cố định vô lý)
- Có recurring revenue cho nền tảng
- Demo được multi-tier pricing — điểm cộng đồ án
- Tương thích với Ad.Budget hiện tại

### 2.2 Bảng `ad_packages`

```sql
CREATE TABLE ad_packages (
    id            VARCHAR(36) PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,          -- "Cơ bản" | "Chuyên nghiệp" | "Doanh nghiệp"
    description   TEXT,
    price_monthly DOUBLE NOT NULL,                -- VNĐ / tháng
    max_slots     INT NOT NULL,                   -- số slot quảng cáo đồng thời
    max_duration_days INT NOT NULL,               -- 30 | 90 | 0 (không giới hạn)
    supports_video BOOLEAN DEFAULT FALSE,
    supports_carousel BOOLEAN DEFAULT FALSE,
    has_advanced_analytics BOOLEAN DEFAULT FALSE,
    sort_order    INT DEFAULT 0,
    created_at    DATETIME NOT NULL
);
```

### 2.3 Bảng `partner_subscriptions`

```sql
CREATE TABLE partner_subscriptions (
    id            VARCHAR(36) PRIMARY KEY,
    user_id       VARCHAR(36) NOT NULL,           -- FK → users(id)
    package_id    VARCHAR(36) NOT NULL,           -- FK → ad_packages(id)
    slots_used    INT DEFAULT 0,                  -- số slot đang dùng
    started_at    DATETIME NOT NULL,
    expires_at    DATETIME NOT NULL,
    status        VARCHAR(20) DEFAULT 'active',   -- active | expired | cancelled
    auto_renew    BOOLEAN DEFAULT TRUE,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (package_id) REFERENCES ad_packages(id)
);
```

### 2.4 Mở rộng `ads` table

Thêm cột:

```sql
ALTER TABLE ads ADD COLUMN slot_id VARCHAR(36) NULL;
ALTER TABLE ads ADD COLUMN package_id VARCHAR(36) NULL;
ALTER TABLE ads ADD COLUMN format VARCHAR(20) DEFAULT 'image';
                                    -- image | video | carousel
ALTER TABLE ads ADD COLUMN max_impressions INT DEFAULT 0;
                                    -- 0 = không giới hạn
ALTER TABLE ads ADD COLUMN daily_budget DOUBLE DEFAULT 0;
                                    -- 0 = dùng budget tổng
```

### 2.5 API Package

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| `GET` | `/api/ads/packages` | No | Danh sách gói công khai |
| `POST` | `/api/ads/subscribe` | Auth+Partner | Đăng ký/mua gói (mô phỏng) |
| `GET` | `/ads-management/subscription` | Auth+Partner | Xem gói hiện tại, slots |
| `POST` | `/ads-management/upgrade` | Auth+Partner | Nâng cấp gói |

### 2.6 Check slot trước khi tạo ad

Thêm logic trong `CreateAd`:
```
1. Lấy subscription của user
2. Kiểm tra status = active, expires_at > now
3. Kiểm tra slots_used < max_slots
4. Kiểm tra format được support (supports_video ...)
5. Nếu OK → increment slots_used, tạo ad
6. Nếu không → trả lỗi + link "/api/ads/packages"
```

---

## 3. Hình thức quảng cáo

### 3.1 Ad Format enum

```go
type AdFormat string

const (
    AdFormatImage    AdFormat = "image"
    AdFormatVideo    AdFormat = "video"
    AdFormatCarousel AdFormat = "carousel"
)
```

- `image`: 1 ảnh tĩnh, hiển thị trong feed/ sidebar
- `video`: video ngắn (< 60s), tự động phát
- `carousel`: 3-5 ảnh, vuốt ngang

### 3.2 Duration options

Thay vì datetime thuần, cho phép partner chọn duration type:

```go
type AdDurationType string

const (
    AdDurationDays  AdDurationType = "days"
    AdDurationWeek  AdDurationType = "week"
    AdDurationMonth AdDurationType = "month"
)
```

Hệ thống tự tính `ExpiresAt = StartedAt + duration`.

### 3.3 Reach & Targeting (cho sau MVP)

Giữ đơn giản: dùng `max_impressions` (trong bảng ads).
Targeting nâng cao (gender, location, interests) có thể thêm sau dưới dạng JSON column `targeting`.

---

## 4. Multi media / Upload

### 4.1 Relation media 1-n

```
ad_media:
  id       VARCHAR(36) PK
  ad_id    VARCHAR(36) FK → ads(id)
  media_id VARCHAR(36) FK → media(id)
  sort_order INT DEFAULT 0
```

Hoặc dùng JSON column `media_ids` trong `ads` nếu muốn đơn giản hơn.

### 4.2 Controller nhận multipart

```go
func (ctrl *AdController) CreateAd(c *gin.Context) {
    title   := c.PostForm("title")
    content := c.PostForm("content")
    ...
    files := c.Request.MultipartForm.File["media"] // []*multipart.FileHeader
}
```

Luồng:
1. Upload từng file lên Cloudinary → lưu media records
2. Tạo ad_media records
3. Tạo Ad record

---

## 5. Analytics có nghĩa

### 5.1 Định nghĩa action types rõ ràng

```go
const (
    ActionImpression = "impression"  // ad được hiển thị 1 lần
    ActionView       = "view"        // user nhìn thấy >= 50% ad trong 1s (viewable impression)
    ActionClick      = "click"       // user click vào ad
    ActionSwipe      = "swipe"       // carousel: vuốt sang slide tiếp
    ActionVideoStart = "video_start" // video: bắt đầu play
    ActionVideoEnd   = "video_end"   // video: xem hết
)
```

Bỏ `interact` mơ hồ — thay bằng các action cụ thể.

### 5.2 Metrics có ngữ cảnh

```go
type AdAnalyticsResponse struct {
    // Tổng quan
    AdID            string  `json:"ad_id"`
    Title           string  `json:"title"`
    Status          string  `json:"status"`
    Format          string  `json:"format"`

    // Budget
    Budget          float64 `json:"budget"`
    DailyBudget     float64 `json:"daily_budget"`
    TotalSpent      float64 `json:"total_spent"`       // tính từ impressions * CPM tham chiếu
    RemainingBudget float64 `json:"remaining_budget"`

    // Volume
    Impressions     int64   `json:"impressions"`       // total
    UniqueReach     int64   `json:"unique_reach"`      // distinct users
    Clicks          int64   `json:"clicks"`
    CTR             float64 `json:"ctr"`                // clicks / impressions * 100

    // Cost
    CPC             float64 `json:"cpc"`                // total_spent / clicks
    CPM             float64 `json:"cpm"`                // total_spent / impressions * 1000

    // Video (if format=video)
    VideoStarts     int64   `json:"video_starts,omitempty"`
    VideoCompletions int64  `json:"video_completions,omitempty"`
    CompletionRate  float64 `json:"completion_rate,omitempty"`

    // Trending (24h / 7d so sánh)
    ImpressionsToday int64 `json:"impressions_today"`
    ClicksToday      int64 `json:"clicks_today"`

    // Time range
    StartedAt    time.Time `json:"started_at"`
    ExpiresAt    *time.Time `json:"expires_at,omitempty"`
    DaysElapsed  int       `json:"days_elapsed"`
    DaysRemaining int      `json:"days_remaining"`
}
```

### 5.3 API analytics mở rộng

| Method | Path | Mô tả |
|--------|------|-------|
| `GET` | `/ads-management/:id/analytics` | Như cũ, response mở rộng |
| `GET` | `/ads-management/:id/analytics/daily?from=&to=` | Daily breakdown (impressions, clicks, cost theo ngày) |
| `GET` | `/ads-management/:id/analytics/summary` | Tóm tắt 1 câu: "Ad A đạt 1.2K impressions, CTR 3.2%, còn 5 ngày" |

### 5.4 Repository optimization

Gộp 3 query `GetCountsByAction` thành 1:

```sql
SELECT action_type, COUNT(*) AS count
FROM ad_analytics
WHERE ad_id = ?
GROUP BY action_type
```

---

## 6. Bảo mật: CheckAdOwnership

### 6.1 Wire middleware vào route

File `routes/ad.routes.go`:

```go
adsManagement.PATCH("/:id/status",
    middlewares.AuthMiddleware(env),
    middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner),
    middlewares.CheckAdOwnership(db, adService),    // ← THÊM DÒNG NÀY
    ctrl.UpdateStatus,
)

adsManagement.GET("/:id/analytics",
    middlewares.AuthMiddleware(env),
    middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin, models.RolePartner),
    middlewares.CheckAdOwnership(db, adService),    // ← THÊM DÒNG NÀY
    ctrl.GetAnalytics,
)
```

### 6.2 Import

`RegisterAdRoutes` nhận thêm `services.AdService` param (đã có sẵn trong `cmd/main.go`).

---

## 7. Kế hoạch triển khai

### Phase 1 — Fix lỗi bảo mật + analytics (nhanh)

| Task | File(s) | Thời gian |
|------|---------|-----------|
| Wire `CheckAdOwnership` vào routes | `routes/ad.routes.go` | 15 phút |
| Gộp 3 COUNT queries → GROUP BY | `repository/ad.repository.go` | 10 phút |
| Định nghĩa action types rõ ràng (bỏ interact) | `dto/ad.dto.go`, `models/ad_analytics.model.go` | 20 phút |
| Mở rộng `AdPerformanceResponse` (CPC, CPM, remaining_budget, daily breakdown) | `dto/ad.dto.go`, `services/ad.service.go`, `repository/ad.repository.go` | 1h |

### Phase 2 — Package model (trọng tâm)

| Task | File(s) | Thời gian |
|------|---------|-----------|
| Tạo `models/ad_package.model.go` | `models/ad_package.model.go` | 20 phút |
| Tạo `models/partner_subscription.model.go` | `models/partner_subscription.model.go` | 20 phút |
| Tạo `repository/package.repository.go` | `repository/package.repository.go` | 30 phút |
| Tạo `services/package.service.go` | `services/package.service.go` | 30 phút |
| Tạo `controllers/package.controller.go` | `controllers/package.controller.go` | 20 phút |
| Tạo `routes/package.routes.go` | `routes/package.routes.go` | 15 phút |
| Wire vào `cmd/main.go` | `cmd/main.go` | 10 phút |
| Check slot trong `CreateAd` | `services/ad.service.go` | 20 phút |
| Thêm seed data packages | `cmd/seed/extended/main.go` | 15 phút |
| Thêm schema CREATE TABLE | `cmd/seed/schema/main.go` | 15 phút |

### Phase 3 — Format + Multi media

| Task | File(s) | Thời gian |
|------|---------|-----------|
| Thêm `format`, `max_impressions`, `daily_budget` vào model | `models/ad.model.go` | 15 phút |
| Tạo `ad_media` model | `models/ad_media.model.go` | 10 phút |
| Đổi `CreateAd` sang multipart | `controllers/ad.controller.go` | 30 phút |
| Upload media files trong `CreateAd` | `services/ad.service.go` | 30 phút |
| Thêm columns vào seed schema | `cmd/seed/schema/main.go` | 10 phút |

### Phase 4 — Analytics nâng cao

| Task | File(s) | Thời gian |
|------|---------|-----------|
| API daily breakdown | `repository/ad.repository.go`, `services/ad.service.go` | 30 phút |
| Migration action types mới | seed + idempotent alter | 15 phút |

### Tổng thời gian ước tính: ~7-8 giờ làm việc

---

## Notes kỹ thuật bổ sung

1. **UUID**: Dùng `utils.GenerateUUID()` (RFC 9562) cho tất cả model mới — không dùng `uuidGenerate()` custom.
2. **GORM tags**: Thêm gorm tags cho tất cả model mới (primaryKey, column, foreignKey, index).
3. **idempotent DDL**: Dùng `addColumnIfMissing` pattern từ schema hiện tại.
4. **Ngôn ngữ lỗi**: Package service trả về Tiếng Việt (consistent với services khác).
5. **Kiểm tra test**: Sau mỗi phase, `go build ./... && go vet ./...` để verify.
