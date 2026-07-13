package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"api_gateway/internal/apihandler"
	"api_gateway/internal/resolver"

	orderpb "api_gateway/internal/proto/orderpb"
	trackingpb "api_gateway/internal/proto/trackingpb"
	userpb "api_gateway/internal/proto/userpb"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

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

	retryPolicy := `{"methodConfig": [{"name": [{"service": ""}], "retryPolicy": {"MaxAttempts": 4, "InitialBackoff": "0.1s", "MaxBackoff": "1s", "BackoffMultiplier": 2.0}}]}`
	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
	}

	// Dial User Service
	userConn, err := grpc.NewClient(userAddr, grpcOpts...)
	if err != nil {
		log.Fatalf("Gateway failed to connect to User Service: %v", err)
	}
	defer userConn.Close()

	// Dial Order Service
	orderConn, err := grpc.NewClient(orderAddr, grpcOpts...)
	if err != nil {
		log.Fatalf("Gateway failed to connect to Order Service: %v", err)
	}
	defer orderConn.Close()

	// Dial Tracking Service
	trackingConn, err := grpc.NewClient(trackingAddr, grpcOpts...)
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
	graphqlHandler := apihandler.New(r)
	mux := http.NewServeMux()
	mux.Handle("/", graphqlHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("Unified API Gateway listening on :%s → /graphql", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Gateway server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Gateway server exiting")
}
