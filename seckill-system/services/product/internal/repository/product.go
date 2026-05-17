package repository

import (
	"seckill-system/services/product/internal/model"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) CreateProduct(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) List(limit, offset int) ([]model.Product, error) {
	var products []model.Product
	err := r.db.Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}

func (r *ProductRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Product{}).Count(&count).Error
	return count, err
}

func (r *ProductRepository) GetByID(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}
