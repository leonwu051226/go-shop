package service

import (
	"seckill-system/services/order/internal/model"
	"seckill-system/services/order/internal/repository"
)

type OrderService struct {
	orderRepo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{orderRepo: repo}
}

func (s *OrderService) GetOrderListByUserID(userID uint) ([]model.Order, error) {
	return s.orderRepo.ListByUserID(userID)
}

func (s *OrderService) GetOrderDetail(orderID uint) (*model.Order, error) {
	return s.orderRepo.GetByID(orderID)
}
