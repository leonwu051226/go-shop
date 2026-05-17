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

func Captcha(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	resp, err := middleware.UserClient.GenerateCaptcha(ctx, &pb.GenerateCaptchaRequest{})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if resp.Code != 0 {
		response.Fail(c, http.StatusInternalServerError, resp.Message)
		return
	}
	response.Success(c, gin.H{
		"captcha_id":    resp.CaptchaId,
		"captcha_image": resp.CaptchaImage,
		"expires_in":    resp.ExpiresIn,
	})
}
