package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"seckill-system/pkg/common/config"
)

const (
	OrderQueueName  = "seckill.order.queue"
	OrderExchange   = "seckill.order.exchange"
	OrderRoutingKey = "seckill.order"
)

type Client struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	confirms <-chan amqp.Confirmation
	mu       sync.Mutex
}

var MQ *Client

type OrderMessage struct {
	OrderID          string  `json:"order_id"`
	UserID           uint    `json:"user_id"`
	ProductID        uint    `json:"product_id"`
	SeckillProductID uint    `json:"seckill_product_id"`
	Quantity         int     `json:"quantity"`
	TotalPrice       float64 `json:"total_price"`
	CreatedAt        string  `json:"created_at"`
}

func InitRabbitMQ(cfg config.RabbitMQConfig) error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.VHost)

	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		OrderExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		OrderQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		OrderQueueName,
		OrderRoutingKey,
		OrderExchange,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	MQ = &Client{
		conn:     conn,
		channel:  ch,
		confirms: ch.NotifyPublish(make(chan amqp.Confirmation, 1)),
	}

	log.Println("RabbitMQ connected successfully")
	return nil
}

func (c *Client) PublishOrderMessage(msg *OrderMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.channel.PublishWithContext(
		ctx,
		OrderExchange,
		OrderRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    msg.OrderID,
			Timestamp:    time.Now(),
			Body:         body,
		},
	); err != nil {
		return err
	}

	select {
	case confirm := <-c.confirms:
		if !confirm.Ack {
			return fmt.Errorf("rabbitmq publish not acknowledged")
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("rabbitmq publish confirm timeout: %w", ctx.Err())
	}
}

func (c *Client) NewChannel() (*amqp.Channel, error) {
	return c.conn.Channel()
}

func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
