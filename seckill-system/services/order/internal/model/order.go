package model

import "time"

type Order struct {
	ID               uint       `json:"id" gorm:"primarykey"`
	OrderNo          string     `json:"order_no" gorm:"uniqueIndex;size:64;not null"`
	UserID           uint       `json:"user_id" gorm:"not null;index;uniqueIndex:idx_user_seckill"`
	ProductID        uint       `json:"product_id" gorm:"not null;index"`
	SeckillProductID *uint      `json:"seckill_product_id" gorm:"index;uniqueIndex:idx_user_seckill"`
	Quantity         int        `json:"quantity" gorm:"not null;default:1"`
	TotalPrice       float64    `json:"total_price" gorm:"type:decimal(10,2);not null"`
	Status           int        `json:"status" gorm:"default:0"`
	LogisticsStatus  int        `json:"logistics_status" gorm:"default:0"`
	TrackingNo       string     `json:"tracking_no" gorm:"size:128"`
	PaidAt           *time.Time `json:"paid_at"`
	ShippedAt        *time.Time `json:"shipped_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
