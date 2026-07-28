package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"linkup/config"
	"linkup/controllers"
	"linkup/db"
	"linkup/groupws"
	"linkup/models"
	"linkup/repository"
	"linkup/routes"
	"linkup/services"
	"linkup/validations"
	"linkup/ws"
)

func main() {
	// 1. Tải các cấu hình môi trường hệ thống
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	if _, err := config.LoadCloudinaryEnv(); err != nil {
		log.Printf("warning: cloudinary config incomplete: %v", err)
	}

	env := config.GetEnv()
	port := env.Port

	// 2. Khởi chạy WebSocket Hub tổng của hệ thống
	hub := ws.NewHub()
	go hub.Run()

	// 3. Kết nối Cơ sở dữ liệu MySQL thông thường
	var gormDB *gorm.DB
	database, err := db.ConnectDb(env)
	if err != nil {
		log.Printf("DB connection: failed (%v)", err)
	} else {
		log.Println("DB connection: success")
		defer database.Close()

		gormDB, err = gorm.Open(mysql.New(mysql.Config{Conn: database}), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to init gorm: %v", err)
		}
	}

	// 4. Khởi tạo Gin Router engine
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Endpoint kiểm tra sức khỏe hệ thống
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 5. Nếu kết nối cơ sở dữ liệu thành công, bắt đầu khởi tạo cấu trúc dự án qua GORM
	if gormDB != nil {
		// ====================================================================
		// TỰ ĐỘNG MIGRATION CHO STORIES & ADVERTISEMENT
		// ====================================================================

		if gormDB.Migrator().HasTable(&models.PartnerSubscription{}) {
			_ = gormDB.Exec("ALTER TABLE partner_subscriptions DROP FOREIGN KEY fk_partner_subscriptions_package")
		}

		if gormDB.Migrator().HasTable(&models.AdPackage{}) {
			_ = gormDB.Exec("ALTER TABLE ad_packages MODIFY COLUMN id VARCHAR(36) NOT NULL")
		}

		if gormDB.Migrator().HasTable(&models.PartnerSubscription{}) {
			_ = gormDB.Exec("ALTER TABLE partner_subscriptions MODIFY COLUMN package_id VARCHAR(36) NOT NULL")
		}

		log.Println("Running database auto-migration...")
		err = gormDB.AutoMigrate(
			&models.User{},
			&models.StoryView{},
			&models.StoryInteract{},
			&models.AdPackage{},
			&models.PartnerSubscription{},
			&models.Ad{},
			&models.AdMedia{},
			&models.AdAnalytics{},
			&models.SystemConfig{},
		)
		if err != nil {
			log.Printf("Warning: Migration failed: %v", err)
		}

		// Tự động kiểm tra và đồng bộ lại tất cả các cột của struct Ad vào MySQL
		// (Giải quyết trường hợp DB đã tạo bảng cũ nhưng thiếu cột mới)
		if gormDB.Migrator().HasTable(&models.Ad{}) {
			_ = gormDB.Migrator().AutoMigrate(&models.Ad{})
		}

		// Tạo Index giải quyết cảnh báo SLOW SQL cho Partner Subscriptions
		if gormDB.Migrator().HasTable(&models.PartnerSubscription{}) {
			if !gormDB.Migrator().HasIndex(&models.PartnerSubscription{}, "idx_user_status_expires") {
				_ = gormDB.Exec("ALTER TABLE partner_subscriptions ADD INDEX idx_user_status_expires (user_id, status, expires_at)")
			}
		}

		// Seed dữ liệu mặc định cho các Gói Quảng Cáo
		seedAdPackages(gormDB)
		// ====================================================================

		// ===== KHỞI TẠO TẦNG ADMIN SETTINGS =====
		adminSettingsRepository := repository.NewAdminSettingsRepository(gormDB)
		adminSettingsService := services.NewAdminSettingsService(adminSettingsRepository, env)
		adminSettingsController := controllers.NewAdminSettingsController(adminSettingsService)

		// ===== KHỞI TẠO TẦNG AUTH & PROFILE =====
		authRepository := repository.NewAuthRepository(gormDB)
		profileRepository := repository.NewProfileRepository(gormDB)
		banRepository := repository.NewBanRepository(gormDB)
		authService := services.NewAuthService(authRepository, profileRepository, banRepository, adminSettingsRepository, env)
		authValidation := validations.NewAuthValidation()
		authController := controllers.NewAuthController(authService, authValidation)
		routes.RegisterAuthRoutes(router, authController, env, gormDB)

		adminSettingsService.SetAuthRepository(authRepository)

		// ===== KHỞI TẠO TẦNG PASSWORD RESET =====
		resetRepository := repository.NewPasswordResetRepository(gormDB)
		passwordResetService := services.NewPasswordResetService(resetRepository, authRepository, authValidation, env)
		passwordResetController := controllers.NewPasswordResetController(passwordResetService, authValidation)
		routes.RegisterPasswordResetRoutes(router, passwordResetController)

		// ===== KHỞI TẠO TẦNG NOTIFICATION (HỖ TRỢ THÔNG BÁO TIN NHẮN/LIKE/COMMENT) =====
		notificationRepository := repository.NewNotificationRepository(gormDB)
		notificationPreferenceRepository := repository.NewNotificationPreferenceRepository(gormDB)
		notificationService := services.NewNotificationService(notificationRepository, notificationPreferenceRepository, profileRepository, hub)
		notificationController := controllers.NewNotificationController(notificationService)
		routes.RegisterNotificationRoutes(router, notificationController, env, gormDB)

		// ===== KHỞI TẠO TẦNG TAG (HASHTAG & MENTION) =====
		tagRepository := repository.NewTagRepository(gormDB)
		tagService := services.NewTagService(tagRepository)
		tagController := controllers.NewTagController(tagService)
		routes.RegisterTagRoutes(router, tagController)

		// ===== KHỞI TẠO TẦNG POST (BÀI VIẾT VÀ BÌNH LUẬN) =====
		postRepository := repository.NewPostRepository(gormDB)
		postValidation := validations.NewPostValidation()
		postService := services.NewPostService(postRepository, notificationService, tagService, postValidation)
		postController := controllers.NewPostController(postService)
		routes.RegisterPostRoutes(router, postController, env, gormDB)

		// ===== KHỞI TẠO TẦNG PROFILE USER =====
		profileService := services.NewProfileService(profileRepository) // profileRepository đã được khởi tạo ở notification layer
		profileController := controllers.NewProfileController(profileService)
		routes.RegisterProfileRoutes(router, profileController, env, gormDB)

		// ===== KHỞI TẠO TẦNG FOLLOW (THEO DÕI) =====
		followRepository := repository.NewFollowRepository(gormDB)
		followService := services.NewFollowService(followRepository, authRepository, notificationService)
		followController := controllers.NewFollowController(followService)
		routes.RegisterFollowRoutes(router, followController, env, gormDB)

		// ===== KHỞI TẠO TẦNG MEDIA (HÌNH ÁNH/FILE TRÊN CLOUDINARY) =====
		mediaRepository := repository.NewMediaRepository(gormDB)
		cldForMedia, _ := cloudinary.NewFromURL(env.CloudinaryEnv)
		aiModerationService := services.NewCloudinaryModerationService(cldForMedia)
		mediaService := services.NewMediaService(*mediaRepository, env.CloudinaryEnv, aiModerationService, notificationService)
		postService.SetMediaService(mediaService)
		mediaService.SetModerationRepo(repository.NewModerationRepository(gormDB))
		mediaController := controllers.NewMediaController(mediaService)
		routes.RegisterMediaRoutes(router, mediaController, env, gormDB)

		// ===== KHỞI TẠO TẦNG STORY (BẢN TIN HIỂN THỊ 24H) =====
		storyRepository := repository.NewStoryRepository(gormDB)
		storyService := services.NewStoryService(storyRepository, mediaService)
		storyController := controllers.NewStoryController(storyService)
		routes.RegisterStoryRoutes(router, storyController, env, gormDB)

		// ===== KHỞI TẠO TẦNG REPORT (BÁO CÁO VI PHẠM) =====
		reportRepository := repository.NewReportRepository(gormDB)
		reportValidation := validations.NewReportValidation()
		reportService := services.NewReportService(reportRepository, authRepository, postRepository, reportValidation)
		reportController := controllers.NewReportController(reportService)
		routes.RegisterReportRoutes(router, reportController, env, gormDB)

		// ===== KHỞI TẠO TẦNG BLOCK (CHẶN USER) =====
		blockRepository := repository.NewBlockRepository(gormDB)
		blockValidation := validations.NewBlockValidation()
		blockService := services.NewBlockService(blockRepository, authRepository, blockValidation)
		blockController := controllers.NewBlockController(blockService)
		routes.RegisterBlockRoutes(router, blockController, env, gormDB)

		// ===== KHỞI TẠO TẦNG FRIEND (BẠN BÈ) =====
		friendRepository := repository.NewFriendRepository(gormDB)
		friendValidation := validations.NewFriendValidation()
		friendService := services.NewFriendService(friendRepository, authRepository, profileRepository, friendValidation, notificationService)
		friendController := controllers.NewFriendController(friendService)
		routes.RegisterFriendRoutes(router, friendController, env, gormDB)

		// ===== KHỞI TẠO TẦNG SEARCH (TÌM KIẾM CHUNG) =====
		searchRepository := repository.NewSearchRepository(gormDB)
		searchValidation := validations.NewSearchValidation()
		searchService := services.NewSearchService(searchRepository, searchValidation)
		searchController := controllers.NewSearchController(searchService)
		routes.RegisterSearchRoutes(router, searchController)

		// ===== KHỞI TẠO CHAT DIRECT (TIN NHẮN TRỰC TIẾP) =====
		chatRepository := repository.NewChatRepository(gormDB)
		friendRepository = repository.NewFriendRepository(gormDB)
		inviteRepository := repository.NewChatInvitationRepository(gormDB)
		chatValidation := validations.NewChatValidation()
		chatService := services.NewChatService(chatRepository, friendRepository, inviteRepository, mediaRepository, notificationService, chatValidation)
		chatHub := ws.NewHub()
		go chatHub.Run()
		chatController := controllers.NewChatController(chatHub, chatService, env)
		routes.RegisterChatRoutes(router, chatController, env, gormDB)

		// ===== KHỞI TẠO GROUP CHAT (TIN NHẮN NHÓM, RỜI NHÓM, CHẶN QUAY LẠI) =====
		groupHub := groupws.NewHub()
		go groupHub.Run()
		groupChatRepository := repository.NewGroupChatRepository(gormDB)
		groupChatService := services.NewGroupChatService(groupChatRepository, chatRepository, notificationService, validations.NewGroupChatValidation())
		groupChatController := controllers.NewGroupChatController(groupChatService, chatService)
		routes.RegisterGroupChatRoutes(router, groupChatController, env, gormDB)
		groupMessageService := services.NewGroupMessageService(chatRepository, groupChatRepository, mediaRepository, notificationService, chatValidation)
		routes.RegisterGroupChatWebSocketRoute(router, groupHub, groupMessageService, groupChatService, env, gormDB)

		// ===== KHỞI TẠO COMMUNITY (NHÓM CỘNG ĐỒNG BÀI VIẾT) =====
		communityRepository := repository.NewCommunityRepository(gormDB)
		communityValidation := validations.NewCommunityValidation()
		communityService := services.NewCommunityService(communityRepository, communityValidation, authRepository, profileRepository, mediaService, notificationService)
		communityController := controllers.NewCommunityController(communityService, mediaService)
		routes.RegisterCommunityRoutes(router, communityController, env, gormDB)

		// ===== KHỞI TẠO COMMUNITY RULE (QUY TẮC CỘNG ĐỒNG) =====
		communityRuleRepository := repository.NewCommunityRuleRepository(gormDB)
		communityRuleValidation := validations.NewCommunityRuleValidation()
		communityRuleService := services.NewCommunityRuleService(communityRuleRepository, communityRepository, communityRuleValidation)
		communityRuleController := controllers.NewCommunityRuleController(communityRuleService)
		routes.RegisterCommunityRuleRoutes(router, communityRuleController, env, gormDB)

		// ===== KHỞI TẠO CONTRIBUTION POLICY (ĐIỂM ĐÓNG GÓP & CHALLENGE) =====
		contributionRepository := repository.NewContributionRepository(gormDB)
		contributionValidation := validations.NewContributionValidation()
		contributionService := services.NewContributionService(contributionRepository, communityRepository, profileRepository, notificationService, contributionValidation)
		postService.SetContributionService(contributionService)
		contributionController := controllers.NewContributionController(contributionService)
		routes.RegisterContributionRoutes(router, contributionController, env, gormDB)

		// ===== KHỞI TẠO TẦNG ADVERTISEMENT & PACKAGES (QUẢNG CÁO & GÓI ĐĂNG KÝ) =====
		adRepository := repository.NewAdRepository(gormDB)
		packageRepository := repository.NewPackageRepository(gormDB)

		adService := services.NewAdService(adRepository, packageRepository, mediaService)
		packageService := services.NewPackageService(packageRepository)

		adController := controllers.NewAdController(adService)
		packageController := controllers.NewPackageController(packageService)

		// Đăng ký routes cho cả Quảng cáo và Gói Đăng ký
		routes.RegisterAdRoutes(router, adController, adService, env, gormDB)
		routes.RegisterPackageRoutes(router, packageController, env, gormDB)

		// ===== KHỞI TẠO TẦNG ADMIN =====
		moderationRepository := repository.NewModerationRepository(gormDB)
		adminRepository := repository.NewAdminRepository(gormDB)
		adminService := services.NewAdminService(authRepository, banRepository, postRepository, reportRepository, moderationRepository, chatRepository, communityRepository, profileRepository, groupChatRepository, adminRepository, mediaRepository, adRepository, notificationService)
		adminService.SetCloudinary(cldForMedia)
		adminController := controllers.NewAdminController(adminService)
		routes.RegisterAdminRoutes(router, adminController, adminSettingsController, env, gormDB)

		// ===== KHỞI TẠO VOICE/VIDEO CALL =====
		callRepository := repository.NewCallRepository(gormDB)
		callService := services.NewVoiceCallService(callRepository, friendRepository, profileRepository, hub)
		callController := controllers.NewVoiceCallController(hub, callService, env)
		routes.RegisterCallRoutes(router, callController, env, gormDB)

		// ===== GROUP CALL =====
		groupCallHub := groupws.NewHub()
		go groupCallHub.Run()
		groupCallHub.SetGroupChatHub(groupHub)
		if mongoClient, err := db.ConnectMongoDB(env.MongoURI); err != nil {
			log.Printf("MongoDB connection: failed (%v) — group calls will not be persisted", err)
		} else {
			log.Println("MongoDB connection: success")
			defer mongoClient.Disconnect(context.Background())
			groupCallRepo := repository.NewGroupCallRepository(mongoClient.Database(env.MongoDBName))
			groupCallHub.SetCallStore(groupCallRepo)
			groupMessageService.SetGroupCallRepository(groupCallRepo)
		}
		groupCallHub.SetMessageService(groupMessageService)
		routes.RegisterGroupCallRoutes(router, groupCallHub, groupMessageService, groupChatService, groupHub, env, gormDB)
	}

	// 6. Lắng nghe cổng kết nối WebSocket thời gian thực tổng
	router.GET("/api/ws", ws.ServeWS(hub, env, gormDB))

	// 7. Khởi chạy toàn bộ hệ thống HTTP Server
	addr := ":" + port
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
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
