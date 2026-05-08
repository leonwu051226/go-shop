package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"seckill-system/internal/service"
	"seckill-system/pkg/response"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserAPI struct {
	userService *service.UserService
}

func NewUserAPI(svc *service.UserService) *UserAPI {
	return &UserAPI{userService: svc}
}

func (a *UserAPI) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := a.userService.Register(req.Username, req.Password)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

func (a *UserAPI) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := a.userService.Login(req.Username, req.Password)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token": token,
	})
}
