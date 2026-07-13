package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"user_service/internal/grpc_server"
	pb "pb/userpb"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on :%s: %v", port, err)
	}

	s := grpc.NewServer()
	
	server := grpc_server.NewUserServer()
	pb.RegisterUserServiceServer(s, server)
	healthpb.RegisterHealthServer(s, health.NewServer())

	go func() {
		log.Printf("User gRPC server listening on :%s", port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down User gRPC server...")

	server.Close()
	s.GracefulStop()
	log.Println("User gRPC server exited gracefully")
}
