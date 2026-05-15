package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/config"
	"seckill-system/pkg/database"
	"seckill-system/services/product/api"
	"seckill-system/services/product/internal/model"
	"seckill-system/services/product/internal/repository"
	"seckill-system/services/product/internal/service"
	"seckill-system/services/product/proto/server"
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

	// 4. 自动迁移（Product + SeckillProduct）
	if err := db.AutoMigrate(
		&model.Product{},
		&model.SeckillProduct{},
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

	// 7. 初始化各层
	seckillRepo := repository.NewSeckillProductRepository(db)
	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(seckillRepo, productRepo)
	productAPI := api.NewProductAPI(productService)

	// 8. 库存预热
	if err := productService.PreheatStock(); err != nil {
		log.Fatalf("Failed to preheat stock: %v", err)
	}
	log.Println("Stock preheated successfully")

	// 9. 启动 gRPC Server
	go func() {
		grpcSrv := server.NewGRPCServer(productService, ":50052")
		if err := grpcSrv.Run(); err != nil {
			log.Fatalf("Product gRPC server failed: %v", err)
		}
	}()

	// 10. 启动 HTTP Server
	r := gin.Default()

	// 配置 CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	// 秒杀商品列表
	r.GET("/api/v1/seckill/products", productAPI.GetSeckillProducts)

	// 11. 启动服务
	port := ":8082"
	log.Printf("Product HTTP server starting on %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initTestData(db *gorm.DB) error {
	// 检查是否已有商品数据
	var count int64
	if err := db.Model(&model.Product{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
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
