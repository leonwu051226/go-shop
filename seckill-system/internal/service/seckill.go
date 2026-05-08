package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"seckill-system/internal/model"
	"seckill-system/internal/repository"
	"seckill-system/pkg/database"
	"seckill-system/pkg/rabbitmq"
)

const (
	StockKeyPrefix = "seckill:stock:%d"
	UserSetPrefix  = "seckill:users:%d"
)

var luaDeductionScript = redis.NewScript(`
	local stockKey = KEYS[1]

	-- Check stock
	local stock = redis.call("GET", stockKey)
	if stock == false or tonumber(stock) <= 0 then
		return 0
	end

	-- Deduct stock
	redis.call("DECR", stockKey)

	return 1
`)

type SeckillService struct {
	seckillRepo *repository.SeckillProductRepository
}

func NewSeckillService(repo *repository.SeckillProductRepository) *SeckillService {
	return &SeckillService{seckillRepo: repo}
}

func (s *SeckillService) GetSeckillProducts() ([]model.SeckillProduct, error) {
	return s.seckillRepo.List()
}

// PreheatStock loads seckill product stock from MySQL to Redis
func (s *SeckillService) PreheatStock() error {
	products, err := s.seckillRepo.List()
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, p := range products {
		stockKey := fmt.Sprintf(StockKeyPrefix, p.ID)
		if err := database.RDB.Set(ctx, stockKey, p.Stock, 0).Err(); err != nil {
			return fmt.Errorf("failed to preheat stock for product %d: %w", p.ID, err)
		}
	}
	return nil
}

// DoSeckill executes the Lua script to atomically deduct stock
// Returns: 1=success, 0=out of stock, -2=error
func (s *SeckillService) DoSeckill(userID, seckillProductID uint) (int, error) {
	stockKey := fmt.Sprintf(StockKeyPrefix, seckillProductID)

	result, err := luaDeductionScript.Run(
		context.Background(),
		database.RDB,
		[]string{stockKey},
	).Int()

	if err != nil {
		return -2, err
	}
	return result, nil
}

// SendOrderMessage sends an order message to RabbitMQ
func (s *SeckillService) SendOrderMessage(userID, seckillProductID uint, orderID string, totalPrice float64) error {
	msg := &rabbitmq.OrderMessage{
		OrderID:          orderID,
		UserID:           userID,
		SeckillProductID: seckillProductID,
		Quantity:         1,
		TotalPrice:       totalPrice,
		CreatedAt:        time.Now().Format(time.RFC3339),
	}
	return rabbitmq.MQ.PublishOrderMessage(msg)
}
