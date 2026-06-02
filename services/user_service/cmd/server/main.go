package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"user_service/internal/grpc_server"
	pb "user_service/internal/proto"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen on :50051: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, grpc_server.NewUserServer())

	log.Println("User gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
