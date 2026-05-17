package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/response"
	"seckill-system/services/gateway/internal/service"
	pb "seckill-system/services/order/proto/gen"
)

func CreateOrder(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}

	var req struct {
		ProductID uint64 `json:"product_id" binding:"required"`
		Quantity  int32  `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := service.OrderClient.CreateOrder(ctx, &pb.CreateOrderRequest{
		UserId:    uint64(uid),
		ProductId: req.ProductID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if resp.Code != 0 {
		response.Fail(c, http.StatusBadRequest, resp.Message)
		return
	}
	response.Success(c, resp.Data)
}

func PayOrder(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := service.OrderClient.PayOrder(ctx, &pb.PayOrderRequest{
		UserId:  uint64(uid),
		OrderId: orderID,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if resp.Code != 0 {
		response.Fail(c, http.StatusBadRequest, resp.Message)
		return
	}
	response.Success(c, resp.Data)
}

func GetOrders(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := service.OrderClient.GetOrderListByUserID(ctx, &pb.GetOrderListByUserIDRequest{
		UserId: uint64(uid),
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

func GetOrderDetail(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := service.OrderClient.GetOrderDetail(ctx, &pb.GetOrderDetailRequest{OrderId: orderID})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if resp.Code != 0 {
		response.Fail(c, http.StatusNotFound, resp.Message)
		return
	}
	if resp.Data.GetUserId() != uint64(uid) {
		response.Fail(c, http.StatusForbidden, "forbidden")
		return
	}
	response.Success(c, resp.Data)
}

func GetAllOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := service.OrderClient.GetAllOrders(ctx, &pb.GetAllOrdersRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
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

func ShipOrder(c *gin.Context) {
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		TrackingNo string `json:"tracking_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := service.OrderClient.ShipOrder(ctx, &pb.ShipOrderRequest{
		OrderId:    orderID,
		TrackingNo: req.TrackingNo,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if resp.Code != 0 {
		response.Fail(c, http.StatusBadRequest, resp.Message)
		return
	}
	response.Success(c, resp.Data)
}

func GetProductSalesStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := service.OrderClient.GetProductSalesStats(ctx, &pb.GetProductSalesStatsRequest{})
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

func currentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "user not authenticated")
		return 0, false
	}
	uid, ok := userID.(uint)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "invalid user identity")
		return 0, false
	}
	return uid, true
}

func parseUintParam(c *gin.Context, name string) (uint64, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return value, true
}
