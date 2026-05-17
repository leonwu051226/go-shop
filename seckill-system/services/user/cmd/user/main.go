package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"seckill-system/pkg/common/config"
	"seckill-system/pkg/common/utils"
	"seckill-system/pkg/database"
	"seckill-system/services/user/api"
	"seckill-system/services/user/internal/model"
	"seckill-system/services/user/internal/repository"
	"seckill-system/services/user/internal/service"
	"seckill-system/services/user/proto/server"
)

func main() {
	config.Init()
	cfg := config.C
	gin.SetMode(cfg.Server.Mode)

	db, err := database.InitMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	log.Println("Auto migrate completed")

	if err := initTestData(db); err != nil {
		log.Printf("Failed to init test data: %v", err)
	}
	if err := database.InitRedis(cfg.Redis); err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userAPI := api.NewUserAPI(userService)

	go func() {
		grpcSrv := server.NewGRPCServer(userService, ":50051")
		if err := grpcSrv.Run(); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	r := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	r.GET("/api/v1/captcha", userAPI.Captcha)
	r.POST("/api/v1/register", userAPI.Register)
	r.POST("/api/v1/login", userAPI.Login)

	port := ":8081"
	log.Printf("User HTTP server starting on %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initTestData(db *gorm.DB) error {
	if err := ensureTestUser(db, "admin", "Admin123456", 0); err != nil {
		return err
	}
	if err := ensureTestUser(db, "merchant01", "Merchant123456", 1); err != nil {
		return err
	}

	log.Println("Test users ready: admin / Admin123456, merchant01 / Merchant123456")
	return nil
}

func ensureTestUser(db *gorm.DB, username, password string, role int8) error {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	var user model.User
	err = db.Where("username = ?", username).First(&user).Error
	if err == nil {
		return db.Model(&user).Updates(map[string]any{
			"password_hash": hash,
			"role":          role,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	return db.Create(&model.User{
		Username:     username,
		PasswordHash: hash,
		Role:         role,
	}).Error
}
