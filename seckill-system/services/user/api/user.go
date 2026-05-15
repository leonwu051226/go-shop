package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"seckill-system/pkg/common/captcha"
	"seckill-system/pkg/common/response"
	"seckill-system/pkg/database"
	"seckill-system/services/user/internal/service"
)

type UserAPI struct {
	userService *service.UserService
}

func NewUserAPI(svc *service.UserService) *UserAPI {
	return &UserAPI{userService: svc}
}

func (a *UserAPI) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := a.userService.Register(req.Username, req.Password)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			status = http.StatusConflict
		}
		response.Fail(c, status, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

func (a *UserAPI) Captcha(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	challenge, err := captcha.Generate(ctx, database.RDB)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, challenge)
}

func (a *UserAPI) Login(c *gin.Context) {
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

	token, err := a.userService.Login(ctx, req.Username, req.Password, req.CaptchaID, req.CaptchaCode, c.ClientIP())
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token": token,
	})
}
