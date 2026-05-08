package model

import "time"

type Product struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Name        string    `json:"name" gorm:"size:128;not null"`
	Description string    `json:"description" gorm:"size:512"`
	Price       float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	Stock       int       `json:"stock" gorm:"not null;default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
