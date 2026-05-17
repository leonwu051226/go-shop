package service

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"seckill-system/pkg/database"
	"seckill-system/pkg/rabbitmq"
	pb "seckill-system/services/product/proto/gen"
)

var ProductClient pb.ProductServiceClient

func InitProductGRPCClient(host string, port int) error {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	ProductClient = pb.NewProductServiceClient(conn)
	return nil
}

type SeckillService struct{}

type SeckillResult struct {
	Result        int
	ReservationID string
	ProductID     uint
	SeckillPrice  float64
}

func NewSeckillService() *SeckillService {
	return &SeckillService{}
}

func (s *SeckillService) GetSeckillProducts() ([]*pb.SeckillProductInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := ProductClient.GetProducts(ctx, &pb.GetProductsRequest{})
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("%s", resp.Message)
	}
	return resp.Data, nil
}

func (s *SeckillService) DoSeckill(userID, seckillProductID uint, requestID string) (*SeckillResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := ProductClient.DeductStock(ctx, &pb.DeductStockRequest{
		SeckillProductId: uint64(seckillProductID),
		UserId:           uint64(userID),
		RequestId:        requestID,
	})
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return &SeckillResult{Result: int(resp.Result)}, fmt.Errorf("%s", resp.Message)
	}
	return &SeckillResult{
		Result:        int(resp.Result),
		ReservationID: resp.ReservationId,
		ProductID:     uint(resp.ProductId),
		SeckillPrice:  resp.SeckillPrice,
	}, nil
}

func (s *SeckillService) SendOrderMessage(userID, seckillProductID uint, orderID string, totalPrice float64) error {
	msg := &rabbitmq.OrderMessage{
		OrderID:          orderID,
		UserID:           userID,
		SeckillProductID: seckillProductID,
		Quantity:         1,
		TotalPrice:       totalPrice,
		CreatedAt:        time.Now().Format(time.RFC3339),
	}
	return rabbitmq.MQ.PublishOrderMessage(msg)
}

func (s *SeckillService) ReleaseReservation(userID, seckillProductID uint) {
	if database.RDB == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stockKey := fmt.Sprintf("seckill:stock:%d", seckillProductID)
	userKey := fmt.Sprintf("seckill:users:%d", seckillProductID)
	if removed, err := database.RDB.SRem(ctx, userKey, userID).Result(); err == nil && removed > 0 {
		database.RDB.Incr(ctx, stockKey)
	}
}
