package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/response"
	"seckill-system/services/gateway/internal/service"
	pb "seckill-system/services/product/proto/gen"
)

func GetProductList(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	keyword := c.Query("keyword")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	minPrice, _ := strconv.ParseFloat(c.DefaultQuery("min_price", "0"), 64)
	maxPrice, _ := strconv.ParseFloat(c.DefaultQuery("max_price", "0"), 64)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := service.ProductClient.GetProductList(ctx, &pb.GetProductListRequest{
		Limit:    int32(limit),
		Offset:   int32(offset),
		Keyword:  keyword,
		MinPrice: minPrice,
		MaxPrice: maxPrice,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if resp.Code != 0 {
		response.Fail(c, http.StatusInternalServerError, resp.Message)
		return
	}

	response.Success(c, gin.H{
		"list":  resp.Data,
		"total": resp.Total,
	})
}

func GetProductDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid product id")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := service.ProductClient.GetProductDetail(ctx, &pb.GetProductDetailRequest{
		Id: id,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if resp.Code != 0 {
		response.Fail(c, http.StatusInternalServerError, resp.Message)
		return
	}

	response.Success(c, resp.Data)
}
