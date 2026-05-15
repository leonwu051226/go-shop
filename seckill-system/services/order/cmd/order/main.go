package main

import (
	"log"

	"seckill-system/pkg/common/config"
	"seckill-system/pkg/database"
	"seckill-system/pkg/rabbitmq"
	"seckill-system/services/order/internal/consumer"
	"seckill-system/services/order/internal/model"
	"seckill-system/services/order/internal/repository"
	"seckill-system/services/order/internal/service"
	"seckill-system/services/order/proto/server"
)

func main() {
	// 1. 加载配置
	config.Init()
	cfg := config.C

	// 2. 初始化 MySQL
	db, err := database.InitMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 3. 自动迁移（Order + SeckillProduct，供消费者落库使用）
	if err := db.AutoMigrate(
		&model.Order{},
		&model.SeckillProduct{},
	); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	log.Println("Auto migrate completed")

	// 4. 初始化 RabbitMQ
	if err := rabbitmq.InitRabbitMQ(cfg.RabbitMQ); err != nil {
		log.Fatalf("Failed to init rabbitmq: %v", err)
	}
	defer rabbitmq.MQ.Close()

	// 5. 初始化订单服务各层
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)

	// 6. 启动 gRPC Server
	go func() {
		grpcSrv := server.NewGRPCServer(orderService, ":50053")
		if err := grpcSrv.Run(); err != nil {
			log.Fatalf("Order gRPC server failed: %v", err)
		}
	}()

	// 7. 启动订单消费者
	consumer.StartOrderConsumer(db)

	// 8. 保持运行
	log.Println("Order service is running...")
	select {}
}
