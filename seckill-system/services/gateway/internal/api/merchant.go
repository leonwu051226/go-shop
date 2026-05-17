package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/response"
	"seckill-system/services/gateway/internal/service"
	pb "seckill-system/services/product/proto/gen"
)

type AddProductRequest struct {
	ProductName        string  `json:"product_name" binding:"required"`
	ProductDescription string  `json:"product_description"`
	ProductPrice       float64 `json:"product_price" binding:"required"`
	Stock              int32   `json:"stock"`
}

func AddProduct(c *gin.Context) {
	var req AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ProductPrice <= 0 {
		response.Fail(c, http.StatusBadRequest, "product_price must be greater than 0")
		return
	}
	if req.Stock < 0 {
		response.Fail(c, http.StatusBadRequest, "stock cannot be negative")
		return
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	log.Printf("[Merchant] AddProduct called by userID=%v username=%v role=%v product=%s", userID, username, role, req.ProductName)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := service.ProductClient.CreateProduct(ctx, &pb.CreateProductRequest{
		ProductName:        req.ProductName,
		ProductDescription: req.ProductDescription,
		ProductPrice:       req.ProductPrice,
		Stock:              req.Stock,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if resp.Code != 0 {
		response.Fail(c, http.StatusBadRequest, resp.Msg)
		return
	}

	response.Success(c, gin.H{
		"product_id": resp.ProductId,
	})
}
