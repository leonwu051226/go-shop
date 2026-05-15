package server

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"seckill-system/pkg/common/jwt"
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
		},
	}, nil
}

func (s *GRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	token, err := s.userService.Login(
		ctx,
		req.Username,
		req.Password,
		firstMetadata(md, "captcha-id"),
		firstMetadata(md, "captcha-code"),
		firstMetadata(md, "client-ip"),
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

func firstMetadata(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
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
		},
	}, nil
}
