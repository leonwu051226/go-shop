package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"seckill-system/pkg/common/config"
	"seckill-system/pkg/common/response"
	pb "seckill-system/services/user/proto/gen"
)

var UserClient pb.UserServiceClient

const MerchantRole int32 = 1

func InitUserGRPCClient(cfg config.UserGRPCConfig) error {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	UserClient = pb.NewUserServiceClient(conn)
	return nil
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !injectIdentity(c) {
			return
		}
		c.Next()
	}
}

func MerchantAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !injectIdentity(c) {
			return
		}
		role, _ := c.Get("role")
		roleValue, ok := role.(int32)
		if !ok || roleValue != MerchantRole {
			response.Fail(c, http.StatusForbidden, "Only for Merchants")
			c.Abort()
			return
		}
		c.Next()
	}
}

func injectIdentity(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Fail(c, http.StatusUnauthorized, "missing authorization header")
		c.Abort()
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.Fail(c, http.StatusUnauthorized, "invalid authorization header format")
		c.Abort()
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := UserClient.ParseToken(ctx, &pb.ParseTokenRequest{
		Token: parts[1],
	})
	if err != nil || resp.Code != 0 {
		response.Fail(c, http.StatusUnauthorized, "invalid or expired token")
		c.Abort()
		return false
	}

	c.Set("user_id", uint(resp.UserId))
	c.Set("username", resp.Username)
	c.Set("role", resp.Role)
	return true
}
