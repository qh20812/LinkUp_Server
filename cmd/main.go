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

		postRepository := repository.NewPostRepository(gormDB)
		postService := services.NewPostService(postRepository)
		postController := controllers.NewPostController(postService)
		routes.RegisterPostRoutes(router, postController, env)

		profileService := services.NewProfileService(profileRepository)
		profileController := controllers.NewProfileController(profileService)
		routes.RegisterProfileRoutes(router, profileController, env)

		followRepository := repository.NewFollowRepository(gormDB)
		followService := services.NewFollowService(followRepository, authRepository)
		followController := controllers.NewFollowController(followService)
		routes.RegisterFollowRoutes(router, followController, env)

		mediaRepository := repository.NewMediaRepository(gormDB)
		mediaService := services.NewMediaService(mediaRepository, env.CloudinaryEnv)
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

		searchRepository := repository.NewSearchRepository(gormDB)
		searchValidation := validations.NewSearchValidation()
		searchService := services.NewSearchService(searchRepository, searchValidation)
		searchController := controllers.NewSearchController(searchService)
		routes.RegisterSearchRoutes(router, searchController)
	}

	addr := ":" + port
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
