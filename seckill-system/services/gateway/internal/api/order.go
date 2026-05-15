package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/response"
	"seckill-system/services/gateway/internal/service"
	pb "seckill-system/services/order/proto/gen"
)

func GetOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "user not authenticated")
		return
	}
	uid := userID.(uint)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
