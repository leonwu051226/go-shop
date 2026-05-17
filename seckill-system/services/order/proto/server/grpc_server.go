package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"seckill-system/services/order/internal/model"
	"seckill-system/services/order/internal/service"
	pb "seckill-system/services/order/proto/gen"
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

func (s *GRPCServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	order, err := s.orderService.CreateOrder(uint(req.UserId), uint(req.ProductId), int(req.Quantity))
	if err != nil {
		return &pb.CreateOrderResponse{Code: 500, Message: err.Error()}, nil
	}
	return &pb.CreateOrderResponse{Code: 0, Message: "success", Data: orderToPB(order)}, nil
}

func (s *GRPCServer) PayOrder(ctx context.Context, req *pb.PayOrderRequest) (*pb.PayOrderResponse, error) {
	order, err := s.orderService.PayOrder(uint(req.UserId), uint(req.OrderId))
	if err != nil {
		return &pb.PayOrderResponse{Code: 500, Message: err.Error()}, nil
	}
	return &pb.PayOrderResponse{Code: 0, Message: "success", Data: orderToPB(order)}, nil
}

func (s *GRPCServer) GetOrderListByUserID(ctx context.Context, req *pb.GetOrderListByUserIDRequest) (*pb.GetOrderListByUserIDResponse, error) {
	orders, err := s.orderService.GetOrderListByUserID(uint(req.UserId))
	if err != nil {
		return &pb.GetOrderListByUserIDResponse{Code: 500, Message: err.Error()}, nil
	}

	infos := make([]*pb.OrderInfo, 0, len(orders))
	for _, o := range orders {
		infos = append(infos, orderToPB(&o))
	}
	return &pb.GetOrderListByUserIDResponse{Code: 0, Message: "success", Data: infos}, nil
}

func (s *GRPCServer) GetOrderDetail(ctx context.Context, req *pb.GetOrderDetailRequest) (*pb.GetOrderDetailResponse, error) {
	order, err := s.orderService.GetOrderDetail(uint(req.OrderId))
	if err != nil {
		return &pb.GetOrderDetailResponse{Code: 500, Message: err.Error()}, nil
	}
	return &pb.GetOrderDetailResponse{Code: 0, Message: "success", Data: orderToPB(order)}, nil
}

func (s *GRPCServer) GetAllOrders(ctx context.Context, req *pb.GetAllOrdersRequest) (*pb.GetAllOrdersResponse, error) {
	orders, total, err := s.orderService.GetAllOrders(int(req.Limit), int(req.Offset))
	if err != nil {
		return &pb.GetAllOrdersResponse{Code: 500, Message: err.Error()}, nil
	}
	infos := make([]*pb.OrderInfo, 0, len(orders))
	for _, o := range orders {
		infos = append(infos, orderToPB(&o))
	}
	return &pb.GetAllOrdersResponse{Code: 0, Message: "success", Data: infos, Total: int32(total)}, nil
}

func (s *GRPCServer) ShipOrder(ctx context.Context, req *pb.ShipOrderRequest) (*pb.ShipOrderResponse, error) {
	order, err := s.orderService.ShipOrder(uint(req.OrderId), req.TrackingNo)
	if err != nil {
		return &pb.ShipOrderResponse{Code: 500, Message: err.Error()}, nil
	}
	return &pb.ShipOrderResponse{Code: 0, Message: "success", Data: orderToPB(order)}, nil
}

func (s *GRPCServer) GetProductSalesStats(ctx context.Context, req *pb.GetProductSalesStatsRequest) (*pb.GetProductSalesStatsResponse, error) {
	stats, err := s.orderService.GetProductSalesStats()
	if err != nil {
		return &pb.GetProductSalesStatsResponse{Code: 500, Message: err.Error()}, nil
	}
	data := make([]*pb.ProductSalesStat, 0, len(stats))
	for _, item := range stats {
		data = append(data, &pb.ProductSalesStat{
			ProductId:    uint64(item.ProductID),
			ProductName:  item.ProductName,
			PaidQuantity: int32(item.PaidQuantity),
			PaidAmount:   item.PaidAmount,
		})
	}
	return &pb.GetProductSalesStatsResponse{Code: 0, Message: "success", Data: data}, nil
}

func orderToPB(order *model.Order) *pb.OrderInfo {
	if order == nil {
		return nil
	}
	var seckillProductID uint64
	if order.SeckillProductID != nil {
		seckillProductID = uint64(*order.SeckillProductID)
	}
	return &pb.OrderInfo{
		Id:               uint64(order.ID),
		OrderNo:          order.OrderNo,
		UserId:           uint64(order.UserID),
		ProductId:        uint64(order.ProductID),
		SeckillProductId: seckillProductID,
		Quantity:         int32(order.Quantity),
		TotalPrice:       order.TotalPrice,
		Status:           int32(order.Status),
		CreatedAt:        formatTime(&order.CreatedAt),
		LogisticsStatus:  int32(order.LogisticsStatus),
		TrackingNo:       order.TrackingNo,
		PaidAt:           formatTime(order.PaidAt),
		ShippedAt:        formatTime(order.ShippedAt),
	}
}

func formatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
