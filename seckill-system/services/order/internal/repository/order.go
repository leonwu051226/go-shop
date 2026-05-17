package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"seckill-system/services/order/internal/model"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateProductOrder(userID, productID uint, quantity int) (*model.Order, error) {
	var order *model.Order
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var product model.Product
		if err := tx.First(&product, productID).Error; err != nil {
			return fmt.Errorf("failed to find product: %w", err)
		}
		if product.Stock < quantity {
			return fmt.Errorf("insufficient stock")
		}

		result := tx.Model(&model.Product{}).
			Where("id = ? AND stock >= ?", productID, quantity).
			Update("stock", gorm.Expr("stock - ?", quantity))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("insufficient stock")
		}

		order = &model.Order{
			OrderNo:    uuid.NewString(),
			UserID:     userID,
			ProductID:  productID,
			Quantity:   quantity,
			TotalPrice: product.Price * float64(quantity),
			Status:     0,
		}
		return tx.Create(order).Error
	})
	return order, err
}

func (r *OrderRepository) ListByUserID(userID uint) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) ListAll(limit, offset int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	if err := r.db.Model(&model.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Limit(limit).Offset(offset).Order("created_at desc").Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) PayOrder(userID, orderID uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return err
		}
		if order.Status != 0 {
			return fmt.Errorf("order is not pending payment")
		}
		now := time.Now()
		if err := tx.Model(&order).Updates(map[string]any{
			"status":  1,
			"paid_at": &now,
		}).Error; err != nil {
			return err
		}
		order.Status = 1
		order.PaidAt = &now
		return nil
	})
	return &order, err
}

func (r *OrderRepository) CancelUnpaidOrder(orderID uint) (*model.Order, bool, error) {
	var order model.Order
	cancelled := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&order, orderID).Error; err != nil {
			return err
		}
		if order.Status != 0 {
			return nil
		}

		if err := tx.Model(&order).Update("status", 3).Error; err != nil {
			return err
		}

		if order.SeckillProductID != nil {
			result := tx.Model(&model.SeckillProduct{}).
				Where("id = ?", *order.SeckillProductID).
				Update("stock", gorm.Expr("stock + ?", order.Quantity))
			if result.Error != nil {
				return result.Error
			}
		} else {
			result := tx.Model(&model.Product{}).
				Where("id = ?", order.ProductID).
				Update("stock", gorm.Expr("stock + ?", order.Quantity))
			if result.Error != nil {
				return result.Error
			}
		}

		order.Status = 3
		cancelled = true
		return nil
	})
	return &order, cancelled, err
}

func (r *OrderRepository) ShipOrder(orderID uint, trackingNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&order, orderID).Error; err != nil {
			return err
		}
		if order.Status != 1 {
			return fmt.Errorf("order is not ready to ship")
		}
		now := time.Now()
		if err := tx.Model(&order).Updates(map[string]any{
			"status":           2,
			"logistics_status": 1,
			"tracking_no":      trackingNo,
			"shipped_at":       &now,
		}).Error; err != nil {
			return err
		}
		order.Status = 2
		order.LogisticsStatus = 1
		order.TrackingNo = trackingNo
		order.ShippedAt = &now
		return nil
	})
	return &order, err
}

type ProductSalesStat struct {
	ProductID    uint
	ProductName  string
	PaidQuantity int
	PaidAmount   float64
}

func (r *OrderRepository) ProductSalesStats() ([]ProductSalesStat, error) {
	var stats []ProductSalesStat
	err := r.db.Table("orders").
		Select("orders.product_id, products.name AS product_name, COALESCE(SUM(orders.quantity), 0) AS paid_quantity, COALESCE(SUM(orders.total_price), 0) AS paid_amount").
		Joins("LEFT JOIN products ON products.id = orders.product_id").
		Where("orders.status IN ?", []int{1, 2}).
		Group("orders.product_id, products.name").
		Order("paid_quantity desc").
		Scan(&stats).Error
	return stats, err
}
