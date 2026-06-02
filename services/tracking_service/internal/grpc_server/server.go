package grpc_server

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	_ "github.com/lib/pq"
	pb "tracking_service/internal/proto"
	orderpb "tracking_service/internal/proto/orderpb"
)

type TrackingServer struct {
	pb.UnimplementedTrackingServiceServer
	db          *sql.DB
	orderClient orderpb.OrderServiceClient
}

func NewTrackingServer() *TrackingServer {
	orderAddr := os.Getenv("ORDER_SERVICE_ADDR")
	if orderAddr == "" {
		orderAddr = "localhost:50052"
	}

	conn, err := grpc.NewClient(orderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Tracking service: failed to connect to Order service at %s: %v", orderAddr, err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://devuser:devpassword@localhost:5432/assignment_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Tracking service: failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Tracking service: failed to ping database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS trackings (
			id UUID PRIMARY KEY,
			order_id UUID NOT NULL,
			status VARCHAR(100) NOT NULL,
			location VARCHAR(255) NOT NULL
		);
	`)
	if err != nil {
		log.Fatalf("Tracking service: failed to create trackings table: %v", err)
	}

	log.Println("Tracking service connected to PostgreSQL successfully")

	return &TrackingServer{
		db:          db,
		orderClient: orderpb.NewOrderServiceClient(conn),
	}
}

func (s *TrackingServer) UpdateTracking(ctx context.Context, req *pb.UpdateTrackingRequest) (*pb.UpdateTrackingResponse, error) {
	log.Printf("[Tracking Service] UpdateTracking request: orderID=%s, status=%q, location=%q", req.GetOrderId(), req.GetStatus(), req.GetLocation())
	
	// ── Inter-service call: validate orderId ─────────────────────────────────
	_, err := s.orderClient.GetOrder(ctx, &orderpb.GetOrderRequest{Id: req.GetOrderId()})
	if err != nil {
		log.Printf("[Tracking Service] UpdateTracking validation failed for orderID %s: %v", req.GetOrderId(), err)
		return nil, status.Errorf(codes.InvalidArgument, "orderId validation failed: %v", err)
	}

	id := uuid.New().String()
	tracking := &pb.OrderTracking{
		Id:       id,
		OrderId:  req.GetOrderId(),
		Status:   req.GetStatus(),
		Location: req.GetLocation(),
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO trackings (id, order_id, status, location) VALUES ($1, $2, $3, $4)", id, tracking.OrderId, tracking.Status, tracking.Location)
	if err != nil {
		log.Printf("[Tracking Service] UpdateTracking database insert error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert tracking: %v", err)
	}

	log.Printf("[Tracking Service] UpdateTracking success: ID=%s", id)
	return &pb.UpdateTrackingResponse{Tracking: tracking}, nil
}

func (s *TrackingServer) TrackOrder(ctx context.Context, req *pb.TrackOrderRequest) (*pb.TrackOrderResponse, error) {
	log.Printf("[Tracking Service] TrackOrder request: orderID=%s", req.GetOrderId())
	rows, err := s.db.QueryContext(ctx, "SELECT id, order_id, status, location FROM trackings WHERE order_id = $1", req.GetOrderId())
	if err != nil {
		log.Printf("[Tracking Service] TrackOrder database query error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to query trackings: %v", err)
	}
	defer rows.Close()

	var trackings []*pb.OrderTracking
	for rows.Next() {
		var t pb.OrderTracking
		if err := rows.Scan(&t.Id, &t.OrderId, &t.Status, &t.Location); err != nil {
			log.Printf("[Tracking Service] TrackOrder scan error: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to scan tracking: %v", err)
		}
		trackings = append(trackings, &t)
	}
	log.Printf("[Tracking Service] TrackOrder success: returned %d tracking events", len(trackings))
	return &pb.TrackOrderResponse{Trackings: trackings}, nil
}
