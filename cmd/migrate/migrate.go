package migrate

import (
	"log"

	"gorm.io/gorm"

	"linkup/models"
)

// Run thực hiện toàn bộ schema migration khi server khởi động.
// Bao gồm: AutoMigrate, thêm cột, FK, đồng bộ collation, seed dữ liệu mặc định.
// Idempotent — chạy lại không gây lỗi hay mất dữ liệu.
func Run(db *gorm.DB) {
	// Bảng do GORM tạo (ad_packages, partner_subscriptions, story_views, ...) mặc định
	// thừa hưởng collation của DB (utf8mb4_0900_ai_ci) trong khi bảng seed dùng
	// utf8mb4_unicode_ci → khi GORM tự tạo FK lệch collation sẽ báo Error 3780.
	// Vì vậy: tắt FK trong AutoMigrate, tự đồng bộ collation toàn bảng, rồi tạo lại FK.
	log.Println("Running database auto-migration...")
	err := db.AutoMigrate(
		&models.User{},
		&models.StoryView{},
		&models.StoryInteract{},
		&models.AdPackage{},
		&models.PartnerSubscription{},
		&models.Ad{},
		&models.AdMedia{},
		&models.AdAnalytics{},
		&models.SystemConfig{},
		&models.UserSetting{},
		&models.UserSession{},
		&models.UserE2EKey{},
		&models.ChatE2EKey{},
	)
	if err != nil {
		log.Printf("Warning: Migration failed: %v", err)
	}

	// Thêm cột messages.e2e_version (0 = legacy server-encrypted, 1 = E2E)
	// idempotent, không làm mất dữ liệu khi chạy lại.
	ensureColumn(db, "messages", "e2e_version", "INT NOT NULL DEFAULT 0")

	// Thêm cột profiles.cover_uri cho tính năng đổi ảnh bìa
	ensureColumn(db, "profiles", "cover_uri", "VARCHAR(500) NOT NULL DEFAULT ''")

	// Thêm các cột profile mới cho About tab
	ensureColumn(db, "profiles", "location", "VARCHAR(255) NOT NULL DEFAULT ''")
	ensureColumn(db, "profiles", "work", "VARCHAR(255) NOT NULL DEFAULT ''")
	ensureColumn(db, "profiles", "education", "VARCHAR(255) NOT NULL DEFAULT ''")
	ensureColumn(db, "profiles", "website", "VARCHAR(255) NOT NULL DEFAULT ''")

	// Thêm cột last_seen vào users cho online/offline presence
	ensureColumn(db, "users", "last_seen", "DATETIME NULL")
	ensureIndex(db, "users", "idx_users_last_seen", "last_seen")

	// Thêm các cột presence vào user_settings
	ensureColumn(db, "user_settings", "activity_status_enabled", "BOOLEAN NOT NULL DEFAULT TRUE")
	ensureColumn(db, "user_settings", "last_seen_visibility", "VARCHAR(20) NOT NULL DEFAULT 'all_friends'")

	// Đồng bộ collation toàn bảng GORM về utf8mb4_unicode_ci (idempotent).
	// "ads" do seed tạo đã là unicode_ci nhưng chạy lại cũng an toàn.
	for _, table := range []string{"ads", "ad_packages", "partner_subscriptions", "ad_media", "ad_analytics", "story_views", "story_interacts"} {
		if db.Migrator().HasTable(table) {
			_ = db.Exec("ALTER TABLE " + table + " CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
		}
	}

	// Tạo lại các FK liên quan ads sau khi collation đã đồng nhất (idempotent).
	ensureForeignKey(db, "ads", "fk_ads_package", "package_id", "ad_packages", "id")
	ensureForeignKey(db, "partner_subscriptions", "fk_partner_subscriptions_package", "package_id", "ad_packages", "id")

	// Tự động kiểm tra và đồng bộ lại tất cả các cột của struct Ad vào MySQL
	// (Giải quyết trường hợp DB đã tạo bảng cũ nhưng thiếu cột mới)
	if db.Migrator().HasTable(&models.Ad{}) {
		_ = db.Migrator().AutoMigrate(&models.Ad{})
	}

	// Tạo Index giải quyết cảnh báo SLOW SQL cho Partner Subscriptions
	if db.Migrator().HasTable(&models.PartnerSubscription{}) {
		if !db.Migrator().HasIndex(&models.PartnerSubscription{}, "idx_user_status_expires") {
			_ = db.Exec("ALTER TABLE partner_subscriptions ADD INDEX idx_user_status_expires (user_id, status, expires_at)")
		}
	}

	// Seed dữ liệu mặc định cho các Gói Quảng Cáo
	seedAdPackages(db)
}

// seedAdPackages hỗ trợ khởi tạo sẵn 3 gói cước mẫu cho hệ thống
func seedAdPackages(db *gorm.DB) {
	var count int64
	db.Model(&models.AdPackage{}).Count(&count)
	if count > 0 {
		return
	}

	packages := []models.AdPackage{
		{
			ID:                   "pkg_basic",
			Name:                 "Gói Cơ Bản (Basic)",
			Description:          "Phù hợp cho cá nhân kinh doanh nhỏ, hỗ trợ 3 chiến dịch ảnh tĩnh.",
			PriceMonthly:         500000,
			MaxSlots:             3,
			MaxDurationDays:      30,
			SupportsVideo:        false,
			SupportsCarousel:     false,
			HasAdvancedAnalytics: false,
			SortOrder:            1,
		},
		{
			ID:                   "pkg_standard",
			Name:                 "Gói Tiêu Chuẩn (Standard)",
			Description:          "Dành cho doanh nghiệp vừa, hỗ trợ tối đa 10 chiến dịch và định dạng Video.",
			PriceMonthly:         1500000,
			MaxSlots:             10,
			MaxDurationDays:      30,
			SupportsVideo:        true,
			SupportsCarousel:     true,
			HasAdvancedAnalytics: true,
			SortOrder:            2,
		},
		{
			ID:                   "pkg_vip",
			Name:                 "Gói VIP Pro",
			Description:          "Không giới hạn sáng tạo, full tính năng Carousel, Video & Analytics nâng cao.",
			PriceMonthly:         3500000,
			MaxSlots:             30,
			MaxDurationDays:      60,
			SupportsVideo:        true,
			SupportsCarousel:     true,
			HasAdvancedAnalytics: true,
			SortOrder:            3,
		},
	}

	for _, pkg := range packages {
		db.Create(&pkg)
	}
	log.Println("[Seed] Đã tạo thành công 3 gói quảng cáo mẫu!")
}

// ensureForeignKey tạo FK nếu chưa tồn tại (idempotent). Kiểm tra
// information_schema trước, tránh lỗi duplicate khi chạy lại server.
func ensureForeignKey(db *gorm.DB, table, fkName, column, refTable, refColumn string) {
	type row struct {
		Count int
	}
	var r row
	err := db.Raw(`SELECT COUNT(*) AS count FROM information_schema.REFERENTIAL_CONSTRAINTS
		WHERE CONSTRAINT_SCHEMA = DATABASE() AND CONSTRAINT_NAME = ? AND TABLE_NAME = ?`,
		fkName, table).Scan(&r).Error
	if err != nil || r.Count > 0 {
		return
	}
	if err := db.Exec("ALTER TABLE "+table+" ADD CONSTRAINT "+fkName+
		" FOREIGN KEY ("+column+") REFERENCES "+refTable+"("+refColumn+")").Error; err != nil {
		log.Printf("Warning: ensure FK %s: %v", fkName, err)
	}
}

// ensureColumn thêm cột nếu chưa tồn tại (idempotent). Kiểm tra
// information_schema trước, tránh lỗi duplicate khi chạy lại server.
func ensureColumn(db *gorm.DB, table, column, definition string) {
	if !db.Migrator().HasTable(table) {
		return
	}
	if db.Migrator().HasColumn(table, column) {
		return
	}
	if err := db.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + definition).Error; err != nil {
		log.Printf("Warning: ensure column %s.%s: %v", table, column, err)
	}
}

// ensureIndex thêm index nếu chưa tồn tại (idempotent).
func ensureIndex(db *gorm.DB, table, indexName, columns string) {
	if !db.Migrator().HasTable(table) {
		return
	}
	if db.Migrator().HasIndex(table, indexName) {
		return
	}
	if err := db.Exec("CREATE INDEX " + indexName + " ON " + table + " (" + columns + ")").Error; err != nil {
		log.Printf("Warning: ensure index %s: %v", indexName, err)
	}
}
