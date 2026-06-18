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
	router.GET("/", func(c *gin.Context) {
		c.String(200, "LinkUp server is running")
	})

	if database != nil {
		gormDB, err := gorm.Open(mysql.New(mysql.Config{Conn: database}), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to init gorm: %v", err)
		}
		authRepository := repository.NewAuthRepository(gormDB)
		authService := services.NewAuthService(authRepository, env)
		authValidation := validations.NewAuthValidation()
		authController := controllers.NewAuthController(authService, authValidation)
		routes.RegisterAuthRoutes(router, authController)
	}

	addr := ":" + port
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
