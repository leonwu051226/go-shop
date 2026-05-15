package consumer

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"seckill-system/pkg/rabbitmq"
	"seckill-system/services/order/internal/model"
)

var errDuplicateOrder = errors.New("duplicate order")

func StartOrderConsumer(db *gorm.DB) {
	ch, err := rabbitmq.MQ.NewChannel()
	if err != nil {
		log.Fatalf("Failed to create consumer channel: %v", err)
	}

	if err := ch.Qos(100, 0, false); err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := ch.Consume(
		rabbitmq.OrderQueueName,
		"seckill-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Println("Order consumer started, listening on queue:", rabbitmq.OrderQueueName)

	go func() {
		for msg := range msgs {
			var orderMsg rabbitmq.OrderMessage
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

func processOrder(db *gorm.DB, msg *rabbitmq.OrderMessage) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		var existing model.Order
		err := tx.Where("order_no = ? OR (user_id = ? AND seckill_product_id = ?)",
			msg.OrderID, msg.UserID, msg.SeckillProductID).
			First(&existing).Error
		if err == nil {
			return errDuplicateOrder
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check duplicate order: %w", err)
		}

		var seckillProduct model.SeckillProduct
		if err := tx.Select("product_id", "seckill_price").First(&seckillProduct, msg.SeckillProductID).Error; err != nil {
			return fmt.Errorf("failed to find seckill product: %w", err)
		}

		result := tx.Model(&model.SeckillProduct{}).
			Where("id = ? AND stock >= ?", msg.SeckillProductID, msg.Quantity).
			Update("stock", gorm.Expr("stock - ?", msg.Quantity))
		if result.Error != nil {
			return fmt.Errorf("failed to deduct stock: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("insufficient stock for seckill product %d", msg.SeckillProductID)
		}

		totalPrice := msg.TotalPrice
		if totalPrice <= 0 {
			totalPrice = seckillProduct.SeckillPrice * float64(msg.Quantity)
		}

		order := &model.Order{
			OrderNo:          msg.OrderID,
			UserID:           msg.UserID,
			ProductID:        seckillProduct.ProductID,
			SeckillProductID: msg.SeckillProductID,
			Quantity:         msg.Quantity,
			TotalPrice:       totalPrice,
			Status:           1,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := tx.Create(order).Error; err != nil {
			if isDuplicateErr(err) {
				return errDuplicateOrder
			}
			return fmt.Errorf("failed to create order: %w", err)
		}

		return nil
	})
	if errors.Is(err, errDuplicateOrder) {
		log.Printf("[Consumer] Duplicate order ignored. orderNo=%s, userID=%d, seckillProductID=%d",
			msg.OrderID, msg.UserID, msg.SeckillProductID)
		return nil
	}
	return err
}

func isDuplicateErr(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "1062")
}
