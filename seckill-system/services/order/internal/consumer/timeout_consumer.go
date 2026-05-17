package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
	"seckill-system/pkg/rabbitmq"
	"seckill-system/services/order/internal/repository"
	productpb "seckill-system/services/product/proto/gen"
)

var ProductClient productpb.ProductServiceClient

func InitProductGRPCClient(host string, port int) error {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	ProductClient = productpb.NewProductServiceClient(conn)
	return nil
}

func StartOrderTimeoutConsumer(db *gorm.DB) {
	ch, err := rabbitmq.MQ.NewChannel()
	if err != nil {
		log.Fatalf("Failed to create timeout consumer channel: %v", err)
	}

	msgs, err := ch.Consume(
		rabbitmq.OrderDLXQueue,
		"order-timeout-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to start timeout consuming: %v", err)
	}

	log.Println("Order timeout consumer started, listening on queue:", rabbitmq.OrderDLXQueue)

	go func() {
		repo := repository.NewOrderRepository(db)
		for msg := range msgs {
			var delayMsg rabbitmq.OrderDelayMessage
			if err := json.Unmarshal(msg.Body, &delayMsg); err != nil {
				log.Printf("[TimeoutConsumer] Failed to unmarshal message: %v, body: %s", err, string(msg.Body))
				msg.Nack(false, false)
				continue
			}

			order, cancelled, err := repo.CancelUnpaidOrder(delayMsg.OrderID)
			if err != nil {
				log.Printf("[TimeoutConsumer] Cancel order failed, will retry. orderID=%d, err=%v", delayMsg.OrderID, err)
				msg.Nack(false, true)
				continue
			}
			needsRedisRelease := order != nil && order.SeckillProductID != nil && (cancelled || order.Status == 3)
			if needsRedisRelease {
				if err := releaseSeckillRedisStock(order.UserID, *order.SeckillProductID, order.Quantity); err != nil {
					log.Printf("[TimeoutConsumer] Release seckill redis stock failed, will retry. orderID=%d, err=%v", delayMsg.OrderID, err)
					msg.Nack(false, true)
					continue
				}
			}

			if cancelled {
				log.Printf("[TimeoutConsumer] Order cancelled. orderID=%d, orderNo=%s", order.ID, order.OrderNo)
			}
			if err := msg.Ack(false); err != nil {
				log.Printf("[TimeoutConsumer] Failed to ack timeout message: %v", err)
			}
		}
	}()
}

func releaseSeckillRedisStock(userID, seckillProductID uint, quantity int) error {
	if ProductClient == nil {
		return fmt.Errorf("product grpc client is not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := ProductClient.ReleaseSeckillStock(ctx, &productpb.ReleaseSeckillStockRequest{
		UserId:           uint64(userID),
		SeckillProductId: uint64(seckillProductID),
		Quantity:         int32(quantity),
	})
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("%s", resp.Message)
	}
	return nil
}
