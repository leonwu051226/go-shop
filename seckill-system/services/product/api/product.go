package api

import (
	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/response"
	"seckill-system/services/product/internal/service"
)

type ProductAPI struct {
	productService *service.ProductService
}

func NewProductAPI(svc *service.ProductService) *ProductAPI {
	return &ProductAPI{productService: svc}
}

func (a *ProductAPI) GetSeckillProducts(c *gin.Context) {
	products, err := a.productService.GetSeckillProducts()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, products)
}
