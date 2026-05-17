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

	OrderDelayExchange   = "order.delay.exchange"
	OrderDelayQueue      = "order.delay.queue"
	OrderDelayRoutingKey = "order.delay"
	OrderDLXExchange     = "order.dlx.exchange"
	OrderDLXQueue        = "order.dlx.queue"
	OrderDLXRoutingKey   = "order.timeout"
	OrderDelayTTL        = int32(15 * 60 * 1000)
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

type OrderDelayMessage struct {
	OrderID uint `json:"order_id"`
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

	if err := declareOrderDelayTopology(ch); err != nil {
		ch.Close()
		conn.Close()
		return err
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

func (c *Client) PublishOrderDelayMessage(msg *OrderDelayMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.publish(OrderDelayExchange, OrderDelayRoutingKey, fmt.Sprintf("%d", msg.OrderID), body)
}

func (c *Client) publish(exchange, routingKey, messageID string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    messageID,
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

func declareOrderDelayTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(OrderDelayExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare order delay exchange: %w", err)
	}
	if err := ch.ExchangeDeclare(OrderDLXExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare order dlx exchange: %w", err)
	}

	_, err := ch.QueueDeclare(
		OrderDelayQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    OrderDLXExchange,
			"x-dead-letter-routing-key": OrderDLXRoutingKey,
			"x-message-ttl":             OrderDelayTTL,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare order delay queue: %w", err)
	}
	if err := ch.QueueBind(OrderDelayQueue, OrderDelayRoutingKey, OrderDelayExchange, false, nil); err != nil {
		return fmt.Errorf("failed to bind order delay queue: %w", err)
	}

	_, err = ch.QueueDeclare(OrderDLXQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare order dlx queue: %w", err)
	}
	if err := ch.QueueBind(OrderDLXQueue, OrderDLXRoutingKey, OrderDLXExchange, false, nil); err != nil {
		return fmt.Errorf("failed to bind order dlx queue: %w", err)
	}
	return nil
}

func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
