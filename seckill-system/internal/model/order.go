package model

import "time"

type Order struct {
	ID               uint      `json:"id" gorm:"primarykey"`
	OrderNo          string    `json:"order_no" gorm:"uniqueIndex;size:64;not null"`
	UserID           uint      `json:"user_id" gorm:"not null;index"`
	ProductID        uint      `json:"product_id" gorm:"not null;index"`
	SeckillProductID uint      `json:"seckill_product_id" gorm:"index"`
	Quantity         int       `json:"quantity" gorm:"not null;default:1"`
	TotalPrice       float64   `json:"total_price" gorm:"type:decimal(10,2);not null"`
	Status           int       `json:"status" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
