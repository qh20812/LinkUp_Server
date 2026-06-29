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
	"linkup/repository"
	"linkup/routes"
	"linkup/services"
	"linkup/validations"
	"linkup/ws"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	if _, err := config.LoadCloudinaryEnv(); err != nil {
		log.Printf("warning: cloudinary config incomplete: %v", err)
	}

	env := config.GetEnv()
	port := env.Port

	hub := ws.NewHub()
	go hub.Run()

	database, err := db.ConnectDb(env)
	if err != nil {
		log.Printf("DB connection: failed (%v)", err)
	} else {
		log.Println("DB connection: success")
		defer database.Close()
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if database != nil {
		gormDB, err := gorm.Open(mysql.New(mysql.Config{Conn: database}), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to init gorm: %v", err)
		}
		authRepository := repository.NewAuthRepository(gormDB)
		profileRepository := repository.NewProfileRepository(gormDB)
		authService := services.NewAuthService(authRepository, profileRepository, env)
		authValidation := validations.NewAuthValidation()
		authController := controllers.NewAuthController(authService, authValidation)
		routes.RegisterAuthRoutes(router, authController, env)

		resetRepository := repository.NewPasswordResetRepository(gormDB)
		passwordResetService := services.NewPasswordResetService(resetRepository, authRepository, authValidation, env)
		passwordResetController := controllers.NewPasswordResetController(passwordResetService, authValidation)
		routes.RegisterPasswordResetRoutes(router, passwordResetController)

		notificationRepository := repository.NewNotificationRepository(gormDB)
		notificationPreferenceRepository := repository.NewNotificationPreferenceRepository(gormDB)
		notificationService := services.NewNotificationService(notificationRepository, notificationPreferenceRepository, hub)
		notificationController := controllers.NewNotificationController(notificationService)
		routes.RegisterNotificationRoutes(router, notificationController, env)

		// 🌟 ĐÃ CẬP NHẬT: Khởi tạo cụm Tag tại đây
		tagRepository := repository.NewTagRepository(gormDB)
		tagService := services.NewTagService(tagRepository)
		tagController := controllers.NewTagController(tagService)
		routes.RegisterTagRoutes(router, tagController)

		postRepository := repository.NewPostRepository(gormDB)
		// 🌟 ĐÃ CẬP NHẬT: Truyền thêm tham số tagService vào hàm khởi tạo
		postService := services.NewPostService(postRepository, notificationService, tagService)
		postController := controllers.NewPostController(postService)
		routes.RegisterPostRoutes(router, postController, env)

		profileService := services.NewProfileService(profileRepository)
		profileController := controllers.NewProfileController(profileService)
		routes.RegisterProfileRoutes(router, profileController, env)

		followRepository := repository.NewFollowRepository(gormDB)
		followService := services.NewFollowService(followRepository, authRepository, notificationService)
		followController := controllers.NewFollowController(followService)
		routes.RegisterFollowRoutes(router, followController, env)

		mediaRepository := repository.NewMediaRepository(gormDB)
		mediaService := services.NewMediaService(*mediaRepository, env.CloudinaryEnv)
		mediaController := controllers.NewMediaController(mediaService)
		routes.RegisterMediaRoutes(router, mediaController, env)

		reportRepository := repository.NewReportRepository(gormDB)
		reportValidation := validations.NewReportValidation()
		reportService := services.NewReportService(reportRepository, authRepository, postRepository, reportValidation)
		reportController := controllers.NewReportController(reportService)
		routes.RegisterReportRoutes(router, reportController, env)

		blockRepository := repository.NewBlockRepository(gormDB)
		blockValidation := validations.NewBlockValidation()
		blockService := services.NewBlockService(blockRepository, authRepository, blockValidation)
		blockController := controllers.NewBlockController(blockService)
		routes.RegisterBlockRoutes(router, blockController, env)

		friendRepository := repository.NewFriendRepository(gormDB)
		friendValidation := validations.NewFriendValidation()
		friendService := services.NewFriendService(friendRepository, authRepository, profileRepository, friendValidation, notificationService)
		friendController := controllers.NewFriendController(friendService)
		routes.RegisterFriendRoutes(router, friendController, env)

		searchRepository := repository.NewSearchRepository(gormDB)
		searchValidation := validations.NewSearchValidation()
		searchService := services.NewSearchService(searchRepository, searchValidation)
		searchController := controllers.NewSearchController(searchService)
		routes.RegisterSearchRoutes(router, searchController)

		chatRepository := repository.NewChatRepository(gormDB)
		friendRepository = repository.NewFriendRepository(gormDB)
		inviteRepository := repository.NewChatInvitationRepository(gormDB)
		chatService := services.NewChatService(chatRepository, friendRepository, inviteRepository)
		chatHub := ws.NewHub()
		go chatHub.Run()
		chatController := controllers.NewChatController(chatHub, chatService, env)
		routes.RegisterChatRoutes(router, chatController, env)

		groupChatRepository := repository.NewGroupChatRepository(gormDB)
		groupChatService := services.NewGroupChatService(groupChatRepository)
		groupChatController := controllers.NewGroupChatController(groupChatService)
		routes.RegisterGroupChatRoutes(router, groupChatController, env)
	}

	router.GET("/ws", ws.ServeWS(hub, env))

	addr := ":" + port
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
