package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"order_service/internal/grpc_server"
	pb "order_service/internal/proto"
)

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen on :50052: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, grpc_server.NewOrderServer())

	log.Println("Order gRPC server listening on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
