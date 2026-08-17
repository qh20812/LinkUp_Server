package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"linkup/cmd/migrate"
	"linkup/config"
	"linkup/controllers"
	"linkup/db"
	"linkup/groupws"
	"linkup/middlewares"
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

		gormDB, err = gorm.Open(mysql.New(mysql.Config{Conn: database}), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			log.Fatalf("failed to init gorm: %v", err)
		}
	}

	// 4. Khởi tạo Gin Router engine
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middlewares.PrometheusMiddleware())

	// Endpoint kiểm tra sức khỏe hệ thống
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "Server đang chạy"})
	})

	// Metrics endpoint cho Prometheus
	router.GET("/metrics", middlewares.MetricsHandler())

	// 5. Nếu kết nối cơ sở dữ liệu thành công, bắt đầu khởi tạo cấu trúc dự án qua GORM
	if gormDB != nil {
		// Schema migration + seed dữ liệu mặc định
		migrate.Run(gormDB)
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

		// ===== KHỞI TẠO TẦNG USER SESSION (QUẢN LÝ PHIÊN ĐĂNG NHẬP) =====
		userSessionRepository := repository.NewUserSessionRepository(gormDB)
		authService.SetSessionRepository(userSessionRepository)

		// ===== KHỞI TẠO TẦNG EMAIL VERIFICATION =====
		emailVerifRepository := repository.NewEmailVerificationRepository(gormDB)
		emailVerifService := services.NewEmailVerificationService(emailVerifRepository, authRepository, adminSettingsRepository, env)
		emailVerifController := controllers.NewEmailVerificationController(emailVerifService, authValidation)
		authService.SetEmailVerificationService(emailVerifService)

		// ===== KHỞI TẠO GOOGLE SIGN-IN (ID-token verification) =====
		if googleVerifier, gerr := services.NewGoogleIDTokenVerifier(context.Background(), env.GoogleClientIDs); gerr != nil {
			log.Printf("Google auth: init failed (%v)", gerr)
		} else {
			authService.SetGoogleIDTokenVerifier(googleVerifier)
		}

		routes.RegisterAuthRoutes(router, authController, emailVerifController, env, gormDB)

		adminSettingsService.SetAuthRepository(authRepository)

		// ===== KHỞI TẠO TẦNG PASSWORD RESET =====
		resetRepository := repository.NewPasswordResetRepository(gormDB)
		passwordResetService := services.NewPasswordResetService(resetRepository, authRepository, adminSettingsRepository, authValidation, env)
		passwordResetController := controllers.NewPasswordResetController(passwordResetService, authValidation)
		routes.RegisterPasswordResetRoutes(router, passwordResetController)

		// ===== KHỞI TẠO TẦNG USER SETTINGS (CÀI ĐẶT & QUYỀN RIÊNG TƯ) =====
		userSettingsRepository := repository.NewUserSettingsRepository(gormDB)
		userSettingsService := services.NewUserSettingsService(userSettingsRepository, userSessionRepository, authRepository, env)
		userSettingsController := controllers.NewUserSettingsController(userSettingsService)
		routes.RegisterSettingsRoutes(router, userSettingsController, env, gormDB)

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
		storyService := services.NewStoryService(storyRepository, profileRepository, mediaService, notificationService)
		storyController := controllers.NewStoryController(storyService)
		routes.RegisterStoryRoutes(router, storyController, env, gormDB)

		// Wire sau khi cả media + story repo đã khởi tạo
		mediaService.SetStoryRepo(storyRepository)
		profileService.SetMediaRepo(mediaRepository)

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
		chatService := services.NewChatService(chatRepository, friendRepository, inviteRepository, mediaRepository, userSettingsRepository, profileRepository, notificationService, chatValidation)
		chatHub := ws.NewHub()
		go chatHub.Run()
		chatController := controllers.NewChatController(chatHub, chatService, env)
		routes.RegisterChatRoutes(router, chatController, env, gormDB)

		// ===== KHỞI TẠO TẦNG PRESENCE (ONLINE/OFFLINE STATUS) =====
		presenceRepository := repository.NewPresenceRepository(gormDB)
		presenceService := services.NewPresenceService(presenceRepository, userSettingsRepository, friendRepository, chatRepository)
		presenceController := controllers.NewPresenceController(presenceService)
		routes.RegisterPresenceRoutes(router, presenceController, env, gormDB)
		hub.SetPresenceService(presenceService)

		// ===== KHỞI TẠO TẦNG E2E (MÃ HÓA ĐẦU CUỐI CHO TIN NHẮN TRỰC TIẾP) =====
		e2eRepository := repository.NewE2ERepository(gormDB)
		e2eService := services.NewE2EService(e2eRepository, chatRepository)
		e2eController := controllers.NewE2EController(e2eService)
		routes.RegisterE2ERoutes(router, e2eController, env, gormDB)

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
		callService := services.NewVoiceCallService(callRepository, friendRepository, profileRepository, notificationService, hub)
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
