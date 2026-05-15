package repository

import (
	"seckill-system/services/product/internal/model"

	"gorm.io/gorm"
)

type SeckillProductRepository struct {
	db *gorm.DB
}

func NewSeckillProductRepository(db *gorm.DB) *SeckillProductRepository {
	return &SeckillProductRepository{db: db}
}

func (r *SeckillProductRepository) List() ([]model.SeckillProduct, error) {
	var products []model.SeckillProduct
	err := r.db.Preload("Product").Where("status = ?", 1).Find(&products).Error
	return products, err
}

func (r *SeckillProductRepository) GetByID(id uint) (*model.SeckillProduct, error) {
	var product model.SeckillProduct
	err := r.db.Preload("Product").First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}
