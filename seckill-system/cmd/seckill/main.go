package main

import (
	"log"
	"time"

	"seckill-system/internal/api"
	"seckill-system/internal/middleware"
	"seckill-system/internal/model"
	"seckill-system/internal/repository"
	"seckill-system/internal/service"
	"seckill-system/pkg/config"
	"seckill-system/pkg/database"
	"seckill-system/pkg/rabbitmq"
	"seckill-system/pkg/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	config.Init()
	cfg := config.C

	// 2. 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 3. 初始化 MySQL
	db, err := database.InitMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 4. 自动迁移
	if err := db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.SeckillProduct{},
		&model.Order{},
	); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	log.Println("Auto migrate completed")

	// 5. 初始化测试数据
	if err := initTestData(db); err != nil {
		log.Printf("Failed to init test data: %v", err)
	}

	// 6. 初始化 Redis
	if err := database.InitRedis(cfg.Redis); err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}

	// 7. 初始化 RabbitMQ
	if err := rabbitmq.InitRabbitMQ(cfg.RabbitMQ); err != nil {
		log.Fatalf("Failed to init rabbitmq: %v", err)
	}
	defer rabbitmq.MQ.Close()

	// 8. 初始化各层
	userRepo := repository.NewUserRepository(db)
	seckillRepo := repository.NewSeckillProductRepository(db)

	userService := service.NewUserService(userRepo)
	seckillService := service.NewSeckillService(seckillRepo)

	userAPI := api.NewUserAPI(userService)
	seckillAPI := api.NewSeckillAPI(seckillService)

	// 9. 库存预热
	if err := seckillService.PreheatStock(); err != nil {
		log.Fatalf("Failed to preheat stock: %v", err)
	}
	log.Println("Stock preheated successfully")

	// 10. 启动 RabbitMQ 订单消费者
	go rabbitmq.StartOrderConsumer(db)

	// 11. 注册路由
	r := gin.Default()

	// 配置 CORS 中间件，必须在注册任何路由之前
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/ping", api.Ping)

	// 用户相关（公开）
	r.POST("/api/v1/register", userAPI.Register)
	r.POST("/api/v1/login", userAPI.Login)

	// 秒杀相关（公开列表，抢购需登录）
	r.GET("/api/v1/seckill/products", seckillAPI.GetSeckillProducts)

	authorized := r.Group("/api/v1")
	authorized.Use(middleware.JWTAuth())
	{
		authorized.POST("/seckill/do", seckillAPI.DoSeckill)
	}

	// 11. 启动服务
	port := cfg.Server.Port
	if port == "" {
		port = ":8080"
	} else if port[0] != ':' {
		port = ":" + port
	}
	log.Printf("Server starting on %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initTestData(db *gorm.DB) error {
	// 检查是否已有默认用户
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// 创建默认用户
	hash, _ := utils.HashPassword("123456")
	admin := &model.User{Username: "admin", PasswordHash: hash}
	if err := db.Create(admin).Error; err != nil {
		return err
	}

	// 创建普通商品
	product1 := &model.Product{Name: "iPhone 16", Description: "Apple iPhone 16", Price: 5999.00, Stock: 1000}
	product2 := &model.Product{Name: "MacBook Pro", Description: "Apple MacBook Pro 14", Price: 14999.00, Stock: 500}
	product3 := &model.Product{Name: "AirPods Pro", Description: "Apple AirPods Pro 2", Price: 1899.00, Stock: 2000}

	if err := db.Create(product1).Error; err != nil {
		return err
	}
	if err := db.Create(product2).Error; err != nil {
		return err
	}
	if err := db.Create(product3).Error; err != nil {
		return err
	}

	// 创建秒杀商品
	now := time.Now()
	seckillProducts := []model.SeckillProduct{
		{
			ProductID:    product1.ID,
			SeckillPrice: 4999.00,
			Stock:        10,
			StartTime:    now.Add(-1 * time.Hour),
			EndTime:      now.Add(24 * time.Hour),
			Status:       1,
		},
		{
			ProductID:    product2.ID,
			SeckillPrice: 12999.00,
			Stock:        5,
			StartTime:    now.Add(-1 * time.Hour),
			EndTime:      now.Add(24 * time.Hour),
			Status:       1,
		},
		{
			ProductID:    product3.ID,
			SeckillPrice: 1499.00,
			Stock:        50,
			StartTime:    now.Add(-1 * time.Hour),
			EndTime:      now.Add(24 * time.Hour),
			Status:       1,
		},
	}

	for _, sp := range seckillProducts {
		if err := db.Create(&sp).Error; err != nil {
			return err
		}
	}

	log.Println("Test data initialized")
	return nil
}
