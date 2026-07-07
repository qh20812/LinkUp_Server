package main

import (
	"fmt"
	"log"

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
	database, err := db.ConnectDb(env)
	if err != nil {
		log.Printf("DB connection: failed (%v)", err)
	} else {
		log.Println("DB connection: success")
		defer database.Close()
	}

	// 4. Khởi tạo Gin Router engine
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Endpoint kiểm tra sức khỏe hệ thống
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 5. Nếu kết nối cơ sở dữ liệu thành công, bắt đầu khởi tạo cấu trúc dự án qua GORM
	if database != nil {
		gormDB, err := gorm.Open(mysql.New(mysql.Config{Conn: database}), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to init gorm: %v", err)
		}

		// ====================================================================
		// TỰ ĐỘNG TẠO BẢNG STORY VIEWS & CẬP NHẬT LẠI CẤU TRÚC BẢNG INTERACT
		// ====================================================================
		log.Println("Running database auto-migration for Stories...")
		err = gormDB.AutoMigrate(&models.StoryView{}, &models.StoryInteract{})
		if err != nil {
			log.Printf("Warning: Migration failed: %v", err)
		}
		// ====================================================================

		// ===== KHỞI TẠO TẦNG AUTH & PROFILE =====
		authRepository := repository.NewAuthRepository(gormDB)
		profileRepository := repository.NewProfileRepository(gormDB)
		banRepository := repository.NewBanRepository(gormDB)
		authService := services.NewAuthService(authRepository, profileRepository, banRepository, env)
		authValidation := validations.NewAuthValidation()
		authController := controllers.NewAuthController(authService, authValidation)
		routes.RegisterAuthRoutes(router, authController, env)

		// ===== KHỞI TẠO TẦNG PASSWORD RESET =====
		resetRepository := repository.NewPasswordResetRepository(gormDB)
		passwordResetService := services.NewPasswordResetService(resetRepository, authRepository, authValidation, env)
		passwordResetController := controllers.NewPasswordResetController(passwordResetService, authValidation)
		routes.RegisterPasswordResetRoutes(router, passwordResetController)

		// ===== KHỞI TẠO TẦNG NOTIFICATION (HỖ TRỢ THÔNG BÁO TIN NHẮN/LIKE/COMMENT) =====
		notificationRepository := repository.NewNotificationRepository(gormDB)
		notificationPreferenceRepository := repository.NewNotificationPreferenceRepository(gormDB)
		notificationService := services.NewNotificationService(notificationRepository, notificationPreferenceRepository, hub)
		notificationController := controllers.NewNotificationController(notificationService)
		routes.RegisterNotificationRoutes(router, notificationController, env)

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
		routes.RegisterPostRoutes(router, postController, env)

		// ===== KHỞI TẠO TẦNG PROFILE USER =====
		profileService := services.NewProfileService(profileRepository)
		profileController := controllers.NewProfileController(profileService)
		routes.RegisterProfileRoutes(router, profileController, env)

		// ===== KHỞI TẠO TẦNG FOLLOW (THEO DÕI) =====
		followRepository := repository.NewFollowRepository(gormDB)
		followService := services.NewFollowService(followRepository, authRepository, notificationService)
		followController := controllers.NewFollowController(followService)
		routes.RegisterFollowRoutes(router, followController, env)

		// ===== KHỞI TẠO TẦNG MEDIA (HÌNH ẢNH/FILE TRÊN CLOUDINARY) =====
		mediaRepository := repository.NewMediaRepository(gormDB)
		mediaService := services.NewMediaService(*mediaRepository, env.CloudinaryEnv)
		mediaController := controllers.NewMediaController(mediaService)
		routes.RegisterMediaRoutes(router, mediaController, env)

		// ===== KHỞI TẠO TẦNG STORY (BẢN TIN HIỂN THỊ 24H) =====
		storyRepository := repository.NewStoryRepository(gormDB)
		storyService := services.NewStoryService(storyRepository)
		storyController := controllers.NewStoryController(storyService)
		routes.RegisterStoryRoutes(router, storyController, env)

		// ===== KHỞI TẠO TẦNG REPORT (BÁO CÁO VI PHẠM) =====
		reportRepository := repository.NewReportRepository(gormDB)
		reportValidation := validations.NewReportValidation()
		reportService := services.NewReportService(reportRepository, authRepository, postRepository, reportValidation)
		reportController := controllers.NewReportController(reportService)
		routes.RegisterReportRoutes(router, reportController, env)

		// ===== KHỞI TẠO TẦNG BLOCK (CHẶN USER) =====
		blockRepository := repository.NewBlockRepository(gormDB)
		blockValidation := validations.NewBlockValidation()
		blockService := services.NewBlockService(blockRepository, authRepository, blockValidation)
		blockController := controllers.NewBlockController(blockService)
		routes.RegisterBlockRoutes(router, blockController, env)

		// ===== KHỞI TẠO TẦNG FRIEND (BẠN BÈ) =====
		friendRepository := repository.NewFriendRepository(gormDB)
		friendValidation := validations.NewFriendValidation()
		friendService := services.NewFriendService(friendRepository, authRepository, profileRepository, friendValidation, notificationService)
		friendController := controllers.NewFriendController(friendService)
		routes.RegisterFriendRoutes(router, friendController, env)

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
		routes.RegisterChatRoutes(router, chatController, env)

		// ===== KHỞI TẠO GROUP CHAT (TIN NHẮN NHÓM, RỜI NHÓM, CHẶN QUAY LẠI) =====
		groupHub := groupws.NewHub()
		go groupHub.Run()
		groupChatRepository := repository.NewGroupChatRepository(gormDB)
		groupChatService := services.NewGroupChatService(groupChatRepository, chatRepository, notificationService, validations.NewGroupChatValidation())
		groupChatController := controllers.NewGroupChatController(groupChatService, chatService)
		routes.RegisterGroupChatRoutes(router, groupChatController, env)
		groupMessageService := services.NewGroupMessageService(chatRepository, groupChatRepository, mediaRepository, notificationService, chatValidation)
		routes.RegisterGroupChatWebSocketRoute(router, groupHub, groupMessageService, groupChatService, env)

		// ===== KHỞI TẠO COMMUNITY (NHÓM CỘNG ĐỒNG BÀI VIẾT) =====
		communityRepository := repository.NewCommunityRepository(gormDB)
		communityValidation := validations.NewCommunityValidation()
		communityService := services.NewCommunityService(communityRepository, communityValidation, authRepository, profileRepository, mediaService, notificationService)
		communityController := controllers.NewCommunityController(communityService, mediaService)
		routes.RegisterCommunityRoutes(router, communityController, env)

		// ===== KHỞI TẠO COMMUNITY RULE (QUY TẮC CỘNG ĐỒNG) =====
		communityRuleRepository := repository.NewCommunityRuleRepository(gormDB)
		communityRuleValidation := validations.NewCommunityRuleValidation()
		communityRuleService := services.NewCommunityRuleService(communityRuleRepository, communityRepository, communityRuleValidation)
		communityRuleController := controllers.NewCommunityRuleController(communityRuleService)
		routes.RegisterCommunityRuleRoutes(router, communityRuleController, env)

		// ===== KHỞI TẠO CONTRIBUTION POLICY (ĐIỂM ĐÓNG GÓP & CHALLENGE) =====
		contributionRepository := repository.NewContributionRepository(gormDB)
		contributionValidation := validations.NewContributionValidation()
		contributionService := services.NewContributionService(contributionRepository, communityRepository, profileRepository, notificationService, contributionValidation)
		postService.SetContributionService(contributionService)
		contributionController := controllers.NewContributionController(contributionService)
		routes.RegisterContributionRoutes(router, contributionController, env)

		// ===== KHỞI TẠO TẦNG ADVERTISEMENT (QUẢNG CÁO & PHÂN QUYỀN PARTNER) =====
		adRepository := repository.NewAdRepository(gormDB)
		adService := services.NewAdService(adRepository)
		adController := controllers.NewAdController(adService)
		routes.RegisterAdRoutes(router, adController, env, gormDB)

		// ===== KHỞI TẠO TẦNG ADMIN =====
		moderationRepository := repository.NewModerationRepository(gormDB)
		adminService := services.NewAdminService(authRepository, banRepository, postRepository, reportRepository, moderationRepository, notificationService)
		adminController := controllers.NewAdminController(adminService)
		routes.RegisterAdminRoutes(router, adminController, env)

		// ===== KHỞI TẠO VOICE/VIDEO CALL =====
		callRepository := repository.NewCallRepository(gormDB)
		callService := services.NewVoiceCallService(callRepository, friendRepository, hub)
		callController := controllers.NewVoiceCallController(hub, callService, env)
		routes.RegisterCallRoutes(router, callController, env)
	}

	// 6. Lắng nghe cổng kết nối WebSocket thời gian thực tổng
	router.GET("/ws", ws.ServeWS(hub, env))

	// 7. Khởi chạy toàn bộ hệ thống HTTP Server
	addr := ":" + port
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
