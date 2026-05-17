package service

import (
	"seckill-system/pkg/rabbitmq"
	"seckill-system/services/order/internal/model"
	"seckill-system/services/order/internal/repository"
)

type OrderService struct {
	orderRepo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{orderRepo: repo}
}

func (s *OrderService) CreateOrder(userID, productID uint, quantity int) (*model.Order, error) {
	if userID == 0 {
		return nil, errInvalid("user id is required")
	}
	if productID == 0 {
		return nil, errInvalid("product id is required")
	}
	if quantity <= 0 {
		quantity = 1
	}
	order, err := s.orderRepo.CreateProductOrder(userID, productID, quantity)
	if err != nil {
		return nil, err
	}
	if err := rabbitmq.MQ.PublishOrderDelayMessage(&rabbitmq.OrderDelayMessage{OrderID: order.ID}); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) GetOrderListByUserID(userID uint) ([]model.Order, error) {
	return s.orderRepo.ListByUserID(userID)
}

func (s *OrderService) GetOrderDetail(orderID uint) (*model.Order, error) {
	return s.orderRepo.GetByID(orderID)
}

func (s *OrderService) PayOrder(userID, orderID uint) (*model.Order, error) {
	return s.orderRepo.PayOrder(userID, orderID)
}

func (s *OrderService) GetAllOrders(limit, offset int) ([]model.Order, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.orderRepo.ListAll(limit, offset)
}

func (s *OrderService) ShipOrder(orderID uint, trackingNo string) (*model.Order, error) {
	return s.orderRepo.ShipOrder(orderID, trackingNo)
}

func (s *OrderService) GetProductSalesStats() ([]repository.ProductSalesStat, error) {
	return s.orderRepo.ProductSalesStats()
}

type errInvalid string

func (e errInvalid) Error() string {
	return string(e)
}
