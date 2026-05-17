package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"seckill-system/services/product/internal/service"
	pb "seckill-system/services/product/proto/gen"
)

type GRPCServer struct {
	pb.UnimplementedProductServiceServer
	productService *service.ProductService
	port           string
}

func NewGRPCServer(productService *service.ProductService, port string) *GRPCServer {
	return &GRPCServer{
		productService: productService,
		port:           port,
	}
}

func (s *GRPCServer) Run() error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, s)

	fmt.Printf("Product gRPC server starting on %s\n", s.port)
	return grpcServer.Serve(lis)
}

func (s *GRPCServer) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	products, err := s.productService.GetSeckillProducts()
	if err != nil {
		return &pb.GetProductsResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	var infos []*pb.SeckillProductInfo
	for _, p := range products {
		infos = append(infos, &pb.SeckillProductInfo{
			Id:                 uint64(p.ID),
			ProductId:          uint64(p.ProductID),
			ProductName:        p.Product.Name,
			ProductDescription: p.Product.Description,
			ProductPrice:       p.Product.Price,
			SeckillPrice:       p.SeckillPrice,
			Stock:              int32(p.Stock),
			StartTime:          p.StartTime.Format(time.RFC3339),
			EndTime:            p.EndTime.Format(time.RFC3339),
			Status:             int32(p.Status),
		})
	}

	return &pb.GetProductsResponse{
		Code:    0,
		Message: "success",
		Data:    infos,
	}, nil
}

func (s *GRPCServer) CheckStock(ctx context.Context, req *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	stock, err := s.productService.CheckStock(uint(req.SeckillProductId))
	if err != nil {
		return &pb.CheckStockResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.CheckStockResponse{
		Code:    0,
		Message: "success",
		Stock:   int32(stock),
	}, nil
}

func (s *GRPCServer) DeductStock(ctx context.Context, req *pb.DeductStockRequest) (*pb.DeductStockResponse, error) {
	reservation, err := s.productService.DoSeckill(uint(req.UserId), uint(req.SeckillProductId), req.RequestId)
	if err != nil {
		return &pb.DeductStockResponse{
			Code:    500,
			Message: err.Error(),
			Result:  int32(reservation.Result),
		}, nil
	}

	return &pb.DeductStockResponse{
		Code:          0,
		Message:       "success",
		Result:        int32(reservation.Result),
		ReservationId: reservation.ReservationID,
		ProductId:     uint64(reservation.ProductID),
		SeckillPrice:  reservation.SeckillPrice,
	}, nil
}

func (s *GRPCServer) GetProductList(ctx context.Context, req *pb.GetProductListRequest) (*pb.GetProductListResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}
	offset := int(req.Offset)

	products, total, err := s.productService.GetProductList(req.Keyword, req.MinPrice, req.MaxPrice, limit, offset)
	if err != nil {
		return &pb.GetProductListResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	var infos []*pb.ProductInfo
	for _, p := range products {
		infos = append(infos, &pb.ProductInfo{
			Id:          uint64(p.ID),
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       int32(p.Stock),
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetProductListResponse{
		Code:    0,
		Message: "success",
		Data:    infos,
		Total:   int32(total),
	}, nil
}

func (s *GRPCServer) GetProductDetail(ctx context.Context, req *pb.GetProductDetailRequest) (*pb.GetProductDetailResponse, error) {
	product, err := s.productService.GetProductDetail(uint(req.Id))
	if err != nil {
		return &pb.GetProductDetailResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &pb.GetProductDetailResponse{
		Code:    0,
		Message: "success",
		Data: &pb.ProductInfo{
			Id:          uint64(product.ID),
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       int32(product.Stock),
			CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *GRPCServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	product, err := s.productService.CreateProduct(
		req.ProductName,
		req.ProductDescription,
		req.ProductPrice,
		int(req.Stock),
	)
	if err != nil {
		return &pb.CreateProductResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.CreateProductResponse{
		Code:      0,
		Msg:       "success",
		ProductId: int64(product.ID),
	}, nil
}

func (s *GRPCServer) ReleaseSeckillStock(ctx context.Context, req *pb.ReleaseSeckillStockRequest) (*pb.ReleaseSeckillStockResponse, error) {
	if err := s.productService.ReleaseSeckillStock(uint(req.UserId), uint(req.SeckillProductId), int(req.Quantity)); err != nil {
		return &pb.ReleaseSeckillStockResponse{
			Code:    500,
			Message: err.Error(),
		}, nil
	}
	return &pb.ReleaseSeckillStockResponse{
		Code:    0,
		Message: "success",
	}, nil
}
