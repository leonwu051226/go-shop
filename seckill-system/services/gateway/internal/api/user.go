package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/response"
	"seckill-system/services/gateway/internal/middleware"
	pb "seckill-system/services/user/proto/gen"
)

func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := middleware.UserClient.Register(ctx, &pb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	if resp.Code != 0 {
		status := http.StatusBadRequest
		if resp.Code == http.StatusConflict {
			status = http.StatusConflict
		}
		response.Fail(c, status, resp.Message)
		return
	}

	response.Success(c, gin.H{
		"id":       resp.Data.Id,
		"username": resp.Data.Username,
		"role":     resp.Data.Role,
	})
}

func Login(c *gin.Context) {
	login(c, false)
}

func MerchantLogin(c *gin.Context) {
	login(c, true)
}

func login(c *gin.Context, merchantOnly bool) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		CaptchaID   string `json:"captcha_id" binding:"required"`
		CaptchaCode string `json:"captcha_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := middleware.UserClient.Login(ctx, &pb.LoginRequest{
		Username:    req.Username,
		Password:    req.Password,
		CaptchaId:   req.CaptchaID,
		CaptchaCode: req.CaptchaCode,
		ClientIp:    c.ClientIP(),
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	if resp.Code != 0 {
		response.Fail(c, http.StatusUnauthorized, resp.Message)
		return
	}

	role := int32(0)
	parseResp, err := middleware.UserClient.ParseToken(ctx, &pb.ParseTokenRequest{Token: resp.Token})
	if err != nil || parseResp.Code != 0 {
		response.Fail(c, http.StatusUnauthorized, "invalid token")
		return
	}
	role = parseResp.Role
	if merchantOnly && role != middleware.MerchantRole {
		response.Fail(c, http.StatusForbidden, "Only for Merchants")
		return
	}

	response.Success(c, gin.H{
		"token": resp.Token,
		"role":  role,
	})
}

func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "user not authenticated")
		return
	}
	uid := userID.(uint)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := middleware.UserClient.GetUserByID(ctx, &pb.GetUserByIDRequest{
		Id: uint64(uid),
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
		"id":       resp.Data.Id,
		"username": resp.Data.Username,
		"role":     resp.Data.Role,
	})
}
