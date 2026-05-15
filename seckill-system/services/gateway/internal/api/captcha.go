package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/captcha"
	"seckill-system/pkg/common/response"
	"seckill-system/pkg/database"
)

func Captcha(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	challenge, err := captcha.Generate(ctx, database.RDB)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, challenge)
}
