package server

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"seckill-system/pkg/common/captcha"
	"seckill-system/pkg/common/jwt"
	"seckill-system/pkg/database"
	"seckill-system/services/user/internal/service"
	pb "seckill-system/services/user/proto/gen"
)

type GRPCServer struct {
	pb.UnimplementedUserServiceServer
	userService *service.UserService
	port        string
}

func NewGRPCServer(userService *service.UserService, port string) *GRPCServer {
	return &GRPCServer{
		userService: userService,
		port:        port,
	}
}

func (s *GRPCServer) Run() error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, s)

	fmt.Printf("User gRPC server starting on %s\n", s.port)
	return grpcServer.Serve(lis)
}

func (s *GRPCServer) GenerateCaptcha(ctx context.Context, req *pb.GenerateCaptchaRequest) (*pb.GenerateCaptchaResponse, error) {
	challenge, err := captcha.Generate(ctx, database.RDB)
	if err != nil {
		return &pb.GenerateCaptchaResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}
	return &pb.GenerateCaptchaResponse{
		Code:         0,
		Message:      "success",
		CaptchaId:    challenge.ID,
		CaptchaImage: challenge.Image,
		ExpiresIn:    int32(challenge.ExpiresIn),
	}, nil
}

func (s *GRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := s.userService.Register(req.Username, req.Password)
	if err != nil {
		code := int32(400)
		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			code = 409
		}
		return &pb.RegisterResponse{
			Code:    code,
			Message: err.Error(),
		}, nil
	}

	return &pb.RegisterResponse{
		Code:    0,
		Message: "success",
		Data: &pb.UserInfo{
			Id:        uint64(user.ID),
			Username:  user.Username,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			Role:      int32(user.Role),
		},
	}, nil
}

func (s *GRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := s.userService.Login(
		ctx,
		req.Username,
		req.Password,
		req.CaptchaId,
		req.CaptchaCode,
		req.ClientIp,
	)
	if err != nil {
		return &pb.LoginResponse{
			Code:    401,
			Message: err.Error(),
		}, nil
	}

	return &pb.LoginResponse{
		Code:    0,
		Message: "success",
		Token:   token,
	}, nil
}

func (s *GRPCServer) ParseToken(ctx context.Context, req *pb.ParseTokenRequest) (*pb.ParseTokenResponse, error) {
	claims, err := jwt.ParseToken(req.Token)
	if err != nil {
		return &pb.ParseTokenResponse{
			Code:    401,
			Message: "invalid or expired token",
		}, nil
	}

	return &pb.ParseTokenResponse{
		Code:     0,
		Message:  "success",
		UserId:   uint64(claims.UserID),
		Username: claims.Username,
		Role:     int32(claims.Role),
	}, nil
}

func (s *GRPCServer) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.GetUserByIDResponse, error) {
	user, err := s.userService.GetByID(uint(req.Id))
	if err != nil {
		return &pb.GetUserByIDResponse{
			Code:    404,
			Message: err.Error(),
		}, nil
	}

	return &pb.GetUserByIDResponse{
		Code:    0,
		Message: "success",
		Data: &pb.UserInfo{
			Id:        uint64(user.ID),
			Username:  user.Username,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			Role:      int32(user.Role),
		},
	}, nil
}
