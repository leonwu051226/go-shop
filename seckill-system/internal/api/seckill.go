package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"seckill-system/internal/service"
	"seckill-system/pkg/response"
)

type SeckillDoRequest struct {
	SeckillProductID uint `json:"seckill_product_id" binding:"required"`
}

type SeckillAPI struct {
	seckillService *service.SeckillService
}

func NewSeckillAPI(svc *service.SeckillService) *SeckillAPI {
	return &SeckillAPI{seckillService: svc}
}

func (a *SeckillAPI) GetSeckillProducts(c *gin.Context) {
	products, err := a.seckillService.GetSeckillProducts()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, products)
}

func (a *SeckillAPI) DoSeckill(c *gin.Context) {
	var req SeckillDoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "user not authenticated")
		return
	}
	uid := userID.(uint)

	result, err := a.seckillService.DoSeckill(uid, req.SeckillProductID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "system error, please try again")
		return
	}

	switch result {
	case -1:
		response.Fail(c, http.StatusForbidden, "you have already purchased this item")
	case 0:
		response.Fail(c, http.StatusGone, "sold out")
	case 1:
		// Get seckill product info for price
		products, _ := a.seckillService.GetSeckillProducts()
		var price float64
		for _, p := range products {
			if p.ID == req.SeckillProductID {
				price = p.SeckillPrice
				break
			}
		}

		orderID := uuid.New().String()
		if err := a.seckillService.SendOrderMessage(uid, req.SeckillProductID, orderID, price); err != nil {
			response.Fail(c, http.StatusInternalServerError, "failed to queue order")
			return
		}

		response.Success(c, gin.H{
			"order_id": orderID,
			"status":   "queuing",
			"message":  "your order is being processed",
		})
	default:
		response.Fail(c, http.StatusInternalServerError, "unknown error")
	}
}
