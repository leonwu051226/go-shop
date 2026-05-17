package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"seckill-system/pkg/common/captcha"
	"seckill-system/pkg/common/jwt"
	"seckill-system/pkg/common/utils"
	"seckill-system/pkg/database"
	"seckill-system/services/user/internal/model"
	"seckill-system/services/user/internal/repository"
)

const (
	maxLoginFailures = 5
	loginFailTTL     = 15 * time.Minute
	loginLockTTL     = 10 * time.Minute
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{userRepo: repo}
}

func (s *UserService) Register(username, password string) (*model.User, error) {
	username = strings.TrimSpace(username)
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: hash,
		Role:         0,
	}
	if err := s.userRepo.Create(user); err != nil {
		if isDuplicateErr(err) {
			return nil, errors.New("username already exists")
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, username, password, captchaID, captchaCode, clientIP string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", errors.New("username and password are required")
	}

	identity := loginIdentity(username, clientIP)
	if err := ensureNotLocked(ctx, identity); err != nil {
		return "", err
	}
	if err := captcha.Validate(ctx, database.RDB, captchaID, captchaCode); err != nil {
		recordLoginFailure(ctx, identity)
		return "", err
	}

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		recordLoginFailure(ctx, identity)
		return "", errors.New("invalid username or password")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		recordLoginFailure(ctx, identity)
		return "", errors.New("invalid username or password")
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", err
	}
	clearLoginFailures(ctx, identity)
	return token, nil
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func validateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return errors.New("username must be 3-32 characters and contain only letters, numbers, or underscores")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return errors.New("password must be 8-72 characters")
	}
	return nil
}

func loginIdentity(username, clientIP string) string {
	if clientIP == "" {
		clientIP = "unknown"
	}
	return strings.ToLower(strings.TrimSpace(username)) + ":" + clientIP
}

func loginFailKey(identity string) string {
	return "auth:login:fail:" + identity
}

func loginLockKey(identity string) string {
	return "auth:login:lock:" + identity
}

func ensureNotLocked(ctx context.Context, identity string) error {
	if database.RDB == nil {
		return errors.New("redis is not initialized")
	}
	ttl, err := database.RDB.TTL(ctx, loginLockKey(identity)).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	if ttl > 0 {
		return fmt.Errorf("too many login attempts, try again in %d seconds", int(ttl.Seconds()))
	}
	return nil
}

func recordLoginFailure(ctx context.Context, identity string) {
	if database.RDB == nil {
		return
	}
	key := loginFailKey(identity)
	count, err := database.RDB.Incr(ctx, key).Result()
	if err != nil {
		return
	}
	if count == 1 {
		database.RDB.Expire(ctx, key, loginFailTTL)
	}
	if count >= maxLoginFailures {
		database.RDB.Set(ctx, loginLockKey(identity), "1", loginLockTTL)
	}
}

func clearLoginFailures(ctx context.Context, identity string) {
	if database.RDB == nil {
		return
	}
	database.RDB.Del(ctx, loginFailKey(identity), loginLockKey(identity))
}

func isDuplicateErr(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "1062")
}
