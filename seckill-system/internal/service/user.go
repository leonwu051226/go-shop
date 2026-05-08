package service

import (
	"errors"

	"seckill-system/internal/model"
	"seckill-system/internal/repository"
	"seckill-system/pkg/utils"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{userRepo: repo}
}

func (s *UserService) Register(username, password string) (*model.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: hash,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(username, password string) (string, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return "", errors.New("invalid username or password")
	}

	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", err
	}
	return token, nil
}
