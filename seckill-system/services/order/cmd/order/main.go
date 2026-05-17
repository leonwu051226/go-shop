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
	config.Init()
	cfg := config.C

	db, err := database.InitMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Order{},
		&model.Product{},
		&model.SeckillProduct{},
	); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	log.Println("Auto migrate completed")

	if err := rabbitmq.InitRabbitMQ(cfg.RabbitMQ); err != nil {
		log.Fatalf("Failed to init rabbitmq: %v", err)
	}
	defer rabbitmq.MQ.Close()

	if err := consumer.InitProductGRPCClient(cfg.ProductGRPC.Host, cfg.ProductGRPC.Port); err != nil {
		log.Fatalf("Failed to init product grpc client: %v", err)
	}

	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)

	go func() {
		grpcSrv := server.NewGRPCServer(orderService, ":50053")
		if err := grpcSrv.Run(); err != nil {
			log.Fatalf("Order gRPC server failed: %v", err)
		}
	}()

	consumer.StartOrderConsumer(db)
	consumer.StartOrderTimeoutConsumer(db)

	log.Println("Order service is running...")
	select {}
}
