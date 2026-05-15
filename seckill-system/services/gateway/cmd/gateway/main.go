package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/config"
	"seckill-system/pkg/database"
	"seckill-system/pkg/rabbitmq"
	"seckill-system/services/gateway/internal/api"
	"seckill-system/services/gateway/internal/middleware"
	"seckill-system/services/gateway/internal/service"
)

func main() {
	config.Init()
	cfg := config.C
	gin.SetMode(cfg.Server.Mode)

	if err := rabbitmq.InitRabbitMQ(cfg.RabbitMQ); err != nil {
		log.Fatalf("Failed to init rabbitmq: %v", err)
	}
	defer rabbitmq.MQ.Close()

	if err := database.InitRedis(cfg.Redis); err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}

	if err := middleware.InitUserGRPCClient(cfg.UserGRPC); err != nil {
		log.Fatalf("Failed to init user grpc client: %v", err)
	}
	if err := service.InitProductGRPCClient(cfg.ProductGRPC.Host, cfg.ProductGRPC.Port); err != nil {
		log.Fatalf("Failed to init product grpc client: %v", err)
	}
	if err := service.InitOrderGRPCClient(cfg.OrderGRPC.Host, cfg.OrderGRPC.Port); err != nil {
		log.Fatalf("Failed to init order grpc client: %v", err)
	}

	seckillService := service.NewSeckillService()
	seckillAPI := api.NewSeckillAPI(seckillService)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/ping", api.Ping)
	r.GET("/api/v1/captcha", api.Captcha)
	r.POST("/api/v1/register", api.Register)
	r.POST("/api/v1/login", api.Login)
	r.GET("/api/v1/products", api.GetProductList)
	r.GET("/api/v1/products/:id", api.GetProductDetail)
	r.GET("/api/v1/seckill/products", seckillAPI.GetSeckillProducts)

	authorized := r.Group("/api/v1")
	authorized.Use(middleware.JWTAuth())
	{
		authorized.POST("/seckill/do", seckillAPI.DoSeckill)
		authorized.GET("/orders", api.GetOrders)
		authorized.GET("/user/profile", api.GetUserProfile)
	}

	port := cfg.Server.Port
	if port == "" {
		port = ":8080"
	} else if port[0] != ':' {
		port = ":" + port
	}
	log.Printf("Gateway server starting on %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start gateway server: %v", err)
	}
}
