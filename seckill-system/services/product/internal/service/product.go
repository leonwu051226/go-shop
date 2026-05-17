package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"seckill-system/pkg/database"
	"seckill-system/services/product/internal/model"
	"seckill-system/services/product/internal/repository"
)

const (
	StockKeyPrefix = "seckill:stock:%d"
	UserKeyPrefix  = "seckill:users:%d"
)

var luaDeductionScript = redis.NewScript(`
	local stockKey = KEYS[1]
	local userKey = KEYS[2]
	local userID = ARGV[1]

	if redis.call("SISMEMBER", userKey, userID) == 1 then
		return -1
	end

	local stock = redis.call("GET", stockKey)
	if stock == false or tonumber(stock) <= 0 then
		return 0
	end

	redis.call("DECR", stockKey)
	redis.call("SADD", userKey, userID)

	return 1
`)

type ProductService struct {
	seckillRepo *repository.SeckillProductRepository
	productRepo *repository.ProductRepository
	searchRepo  *repository.ProductSearchRepository
}

type SeckillReservation struct {
	Result        int
	ReservationID string
	ProductID     uint
	SeckillPrice  float64
}

func NewProductService(seckillRepo *repository.SeckillProductRepository, productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{
		seckillRepo: seckillRepo,
		productRepo: productRepo,
	}
}

func (s *ProductService) SetSearchRepository(searchRepo *repository.ProductSearchRepository) {
	s.searchRepo = searchRepo
}

func (s *ProductService) GetSeckillProducts() ([]model.SeckillProduct, error) {
	return s.seckillRepo.List()
}

// PreheatStock loads seckill product stock from MySQL to Redis
func (s *ProductService) PreheatStock() error {
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

// DoSeckill atomically checks duplicate purchase and reserves Redis stock.
// Result: 1=success, 0=out of stock, -1=already purchased, -2=error
func (s *ProductService) DoSeckill(userID, seckillProductID uint, requestID string) (*SeckillReservation, error) {
	if userID == 0 {
		return &SeckillReservation{Result: -2}, fmt.Errorf("user id is required")
	}
	if requestID == "" {
		return &SeckillReservation{Result: -2}, fmt.Errorf("request id is required")
	}
	product, err := s.seckillRepo.GetByID(seckillProductID)
	if err != nil {
		return &SeckillReservation{Result: -2}, err
	}
	now := time.Now()
	if product.Status != 1 || now.Before(product.StartTime) || now.After(product.EndTime) {
		return &SeckillReservation{Result: -2}, fmt.Errorf("seckill product is not active")
	}

	stockKey := fmt.Sprintf(StockKeyPrefix, seckillProductID)
	userKey := fmt.Sprintf(UserKeyPrefix, seckillProductID)

	result, err := luaDeductionScript.Run(
		context.Background(),
		database.RDB,
		[]string{stockKey, userKey},
		fmt.Sprintf("%d", userID),
	).Int()

	if err != nil {
		return &SeckillReservation{Result: -2}, err
	}
	return &SeckillReservation{
		Result:        result,
		ReservationID: requestID,
		ProductID:     product.ProductID,
		SeckillPrice:  product.SeckillPrice,
	}, nil
}

func (s *ProductService) CheckStock(seckillProductID uint) (int, error) {
	stockKey := fmt.Sprintf(StockKeyPrefix, seckillProductID)
	ctx := context.Background()
	stock, err := database.RDB.Get(ctx, stockKey).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return stock, nil
}

func (s *ProductService) GetByID(id uint) (*model.SeckillProduct, error) {
	return s.seckillRepo.GetByID(id)
}

func (s *ProductService) GetProductList(keyword string, minPrice, maxPrice float64, limit, offset int) ([]model.Product, int64, error) {
	if s.searchRepo != nil && (strings.TrimSpace(keyword) != "" || minPrice > 0 || maxPrice > 0) {
		result, err := s.searchRepo.SearchProducts(context.Background(), keyword, minPrice, maxPrice, limit, offset)
		if err == nil {
			return result.Products, result.Total, nil
		}
	}

	products, err := s.productRepo.List(limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.productRepo.Count()
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (s *ProductService) GetProductDetail(id uint) (*model.Product, error) {
	return s.productRepo.GetByID(id)
}

func (s *ProductService) CreateProduct(name, description string, price float64, stock int) (*model.Product, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if price <= 0 {
		return nil, fmt.Errorf("product price must be greater than 0")
	}
	if stock < 0 {
		return nil, fmt.Errorf("stock cannot be negative")
	}

	product := &model.Product{
		Name:        name,
		Description: strings.TrimSpace(description),
		Price:       price,
		Stock:       stock,
	}
	if err := s.productRepo.CreateProduct(product); err != nil {
		return nil, err
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.IndexProduct(context.Background(), product); err != nil {
			return nil, err
		}
	}
	return product, nil
}

func (s *ProductService) ReleaseSeckillStock(userID, seckillProductID uint, quantity int) error {
	if userID == 0 {
		return fmt.Errorf("user id is required")
	}
	if seckillProductID == 0 {
		return fmt.Errorf("seckill product id is required")
	}
	if quantity <= 0 {
		quantity = 1
	}

	ctx := context.Background()
	stockKey := fmt.Sprintf(StockKeyPrefix, seckillProductID)
	userKey := fmt.Sprintf(UserKeyPrefix, seckillProductID)
	removed, err := database.RDB.SRem(ctx, userKey, userID).Result()
	if err != nil {
		return err
	}
	if removed > 0 {
		if err := database.RDB.IncrBy(ctx, stockKey, int64(quantity)).Err(); err != nil {
			return err
		}
	}
	return nil
}
