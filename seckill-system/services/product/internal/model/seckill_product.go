package model

import "time"

type SeckillProduct struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	ProductID    uint      `json:"product_id" gorm:"not null;index"`
	Product      Product   `json:"product" gorm:"foreignKey:ProductID"`
	SeckillPrice float64   `json:"seckill_price" gorm:"type:decimal(10,2);not null"`
	Stock        int       `json:"stock" gorm:"not null;default:0"`
	StartTime    time.Time `json:"start_time" gorm:"not null"`
	EndTime      time.Time `json:"end_time" gorm:"not null"`
	Status       int       `json:"status" gorm:"default:1"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
