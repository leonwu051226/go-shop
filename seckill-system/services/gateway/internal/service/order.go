package service

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "seckill-system/services/order/proto/gen"
)

var OrderClient pb.OrderServiceClient

func InitOrderGRPCClient(host string, port int) error {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	OrderClient = pb.NewOrderServiceClient(conn)
	return nil
}
