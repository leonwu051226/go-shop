package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	pb "seckill-system/services/order/proto/gen"
	"seckill-system/services/order/internal/service"
)

type GRPCServer struct {
	pb.UnimplementedOrderServiceServer
	orderService *service.OrderService
	port         string
}

func NewGRPCServer(orderService *service.OrderService, port string) *GRPCServer {
	return &GRPCServer{
		orderService: orderService,
		port:         port,
	}
}

func (s *GRPCServer) Run() error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, s)

	fmt.Printf("Order gRPC server starting on %s\n", s.port)
	return grpcServer.Serve(lis)
}

func (s *GRPCServer) GetOrderListByUserID(ctx context.Context, req *pb.GetOrderListByUserIDRequest) (*pb.GetOrderListByUserIDResponse, error) {
	orders, err := s.orderService.GetOrderListByUserID(uint(req.UserId))
	if err != nil {
		return &pb.GetOrderListByUserIDResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	var infos []*pb.OrderInfo
	for _, o := range orders {
		infos = append(infos, &pb.OrderInfo{
			Id:               uint64(o.ID),
			OrderNo:          o.OrderNo,
			UserId:           uint64(o.UserID),
			ProductId:        uint64(o.ProductID),
			SeckillProductId: uint64(o.SeckillProductID),
			Quantity:         int32(o.Quantity),
			TotalPrice:       o.TotalPrice,
			Status:           int32(o.Status),
			CreatedAt:        o.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetOrderListByUserIDResponse{
		Code:    0,
		Message: "success",
		Data:    infos,
	}, nil
}

func (s *GRPCServer) GetOrderDetail(ctx context.Context, req *pb.GetOrderDetailRequest) (*pb.GetOrderDetailResponse, error) {
	order, err := s.orderService.GetOrderDetail(uint(req.OrderId))
	if err != nil {
		return &pb.GetOrderDetailResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.GetOrderDetailResponse{
		Code:    0,
		Message: "success",
		Data: &pb.OrderInfo{
			Id:               uint64(order.ID),
			OrderNo:          order.OrderNo,
			UserId:           uint64(order.UserID),
			ProductId:        uint64(order.ProductID),
			SeckillProductId: uint64(order.SeckillProductID),
			Quantity:         int32(order.Quantity),
			TotalPrice:       order.TotalPrice,
			Status:           int32(order.Status),
			CreatedAt:        order.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}
