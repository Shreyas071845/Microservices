package main

import (
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"api_gateway/internal/apihandler"
	"api_gateway/internal/resolver"

	orderpb "api_gateway/internal/proto/orderpb"
	trackingpb "api_gateway/internal/proto/trackingpb"
	userpb "api_gateway/internal/proto/userpb"
)

func main() {
	userAddr := os.Getenv("USER_SERVICE_ADDR")
	if userAddr == "" {
		userAddr = "localhost:50051"
	}
	orderAddr := os.Getenv("ORDER_SERVICE_ADDR")
	if orderAddr == "" {
		orderAddr = "localhost:50052"
	}
	trackingAddr := os.Getenv("TRACKING_SERVICE_ADDR")
	if trackingAddr == "" {
		trackingAddr = "localhost:50053"
	}

	// Dial User Service
	userConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Gateway failed to connect to User Service: %v", err)
	}
	defer userConn.Close()

	// Dial Order Service
	orderConn, err := grpc.NewClient(orderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Gateway failed to connect to Order Service: %v", err)
	}
	defer orderConn.Close()

	// Dial Tracking Service
	trackingConn, err := grpc.NewClient(trackingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Gateway failed to connect to Tracking Service: %v", err)
	}
	defer trackingConn.Close()

	// Initialize the Resolver with the 3 gRPC clients
	r := &resolver.Resolver{
		UserClient:     userpb.NewUserServiceClient(userConn),
		OrderClient:    orderpb.NewOrderServiceClient(orderConn),
		TrackingClient: trackingpb.NewTrackingServiceClient(trackingConn),
	}

	// Mount the HTTP handler and start serving
	mux := apihandler.New(r)
	log.Println("Unified API Gateway listening on :8080 → /graphql")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
