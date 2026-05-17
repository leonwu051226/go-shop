package model

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	Username     string    `json:"username" gorm:"uniqueIndex;size:64;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	Role         int8      `json:"role" gorm:"not null;default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
