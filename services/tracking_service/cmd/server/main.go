package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"tracking_service/internal/grpc_server"
	pb "tracking_service/internal/proto"
)

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen on :50053: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTrackingServiceServer(s, grpc_server.NewTrackingServer())

	log.Println("Tracking gRPC server listening on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
