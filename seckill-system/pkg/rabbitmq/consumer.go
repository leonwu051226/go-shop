package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"seckill-system/internal/model"
)

func StartOrderConsumer(db *gorm.DB) {
	ch, err := MQ.NewChannel()
	if err != nil {
		log.Fatalf("Failed to create consumer channel: %v", err)
	}

	err = ch.Qos(100, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := ch.Consume(
		OrderQueueName,
		"seckill-consumer",
		false, // autoAck = false
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Println("Order consumer started, listening on queue:", OrderQueueName)

	go func() {
		for msg := range msgs {
			var orderMsg OrderMessage
			if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
				log.Printf("[Consumer] Failed to unmarshal message: %v, body: %s", err, string(msg.Body))
				msg.Nack(false, false)
				continue
			}

			if err := processOrder(db, &orderMsg); err != nil {
				if strings.Contains(err.Error(), "insufficient stock") {
					log.Printf("[Consumer] Insufficient stock, discarding. orderNo=%s, seckillProductID=%d", orderMsg.OrderID, orderMsg.SeckillProductID)
					msg.Nack(false, false)
					continue
				}

				log.Printf("[Consumer] Process order failed, will retry. orderNo=%s, err=%v", orderMsg.OrderID, err)
				msg.Nack(false, true)
				continue
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("[Consumer] Failed to ack message: %v", err)
			} else {
				log.Printf("[Consumer] Order processed successfully. orderNo=%s", orderMsg.OrderID)
			}
		}
	}()
}

func processOrder(db *gorm.DB, msg *OrderMessage) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 获取秒杀商品对应的 ProductID
		var seckillProduct model.SeckillProduct
		if err := tx.Select("product_id").First(&seckillProduct, msg.SeckillProductID).Error; err != nil {
			return fmt.Errorf("failed to find seckill product: %w", err)
		}

		// 2. 乐观锁扣减库存
		result := tx.Model(&model.SeckillProduct{}).
			Where("id = ? AND stock > 0", msg.SeckillProductID).
			Update("stock", gorm.Expr("stock - ?", msg.Quantity))
		if result.Error != nil {
			return fmt.Errorf("failed to deduct stock: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("insufficient stock for seckill product %d", msg.SeckillProductID)
		}

		// 3. 插入订单（利用唯一索引保证幂等性）
		order := &model.Order{
			OrderNo:          msg.OrderID,
			UserID:           msg.UserID,
			ProductID:        seckillProduct.ProductID,
			SeckillProductID: msg.SeckillProductID,
			Quantity:         msg.Quantity,
			TotalPrice:       msg.TotalPrice,
			Status:           1, // 已支付/已创建
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		return nil
	})
}
