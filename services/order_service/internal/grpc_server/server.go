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
	pb "order_service/internal/proto"
	userpb "order_service/internal/proto/userpb"
)

type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	db         *sql.DB
	userClient userpb.UserServiceClient
}

func NewOrderServer() *OrderServer {
	userAddr := os.Getenv("USER_SERVICE_ADDR")
	if userAddr == "" {
		userAddr = "localhost:50051"
	}

	conn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Order service: failed to connect to User service at %s: %v", userAddr, err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://devuser:devpassword@localhost:5432/assignment_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Order service: failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Order service: failed to ping database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			product_name VARCHAR(255) NOT NULL,
			quantity INT NOT NULL,
			status VARCHAR(50) NOT NULL
		);
	`)
	if err != nil {
		log.Fatalf("Order service: failed to create orders table: %v", err)
	}

	log.Println("Order service connected to PostgreSQL successfully")

	return &OrderServer{
		db:         db,
		userClient: userpb.NewUserServiceClient(conn),
	}
}

func (s *OrderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	log.Printf("[Order Service] CreateOrder request: userID=%s, product=%q, qty=%d", req.GetUserId(), req.GetProductName(), req.GetQuantity())
	
	// ── Inter-service call: validate userId ──────────────────────────────────
	_, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{Id: req.GetUserId()})
	if err != nil {
		log.Printf("[Order Service] CreateOrder validation failed for userID %s: %v", req.GetUserId(), err)
		return nil, status.Errorf(codes.InvalidArgument, "userId validation failed: %v", err)
	}

	id := uuid.New().String()
	order := &pb.Order{
		Id:          id,
		UserId:      req.GetUserId(),
		ProductName: req.GetProductName(),
		Quantity:    req.GetQuantity(),
		Status:      "PLACED",
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO orders (id, user_id, product_name, quantity, status) VALUES ($1, $2, $3, $4, $5)", id, order.UserId, order.ProductName, order.Quantity, order.Status)
	if err != nil {
		log.Printf("[Order Service] CreateOrder database insert error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert order: %v", err)
	}

	log.Printf("[Order Service] CreateOrder success: ID=%s", id)
	return &pb.CreateOrderResponse{Order: order}, nil
}

func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	log.Printf("[Order Service] GetOrder request: ID=%s", req.GetId())
	var order pb.Order
	err := s.db.QueryRowContext(ctx, "SELECT id, user_id, product_name, quantity, status FROM orders WHERE id = $1", req.GetId()).Scan(&order.Id, &order.UserId, &order.ProductName, &order.Quantity, &order.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[Order Service] GetOrder info: order %q not found", req.GetId())
			return nil, status.Errorf(codes.NotFound, "order %q not found", req.GetId())
		}
		log.Printf("[Order Service] GetOrder database error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to query order: %v", err)
	}
	log.Printf("[Order Service] GetOrder success: product=%q", order.ProductName)
	return &pb.GetOrderResponse{Order: &order}, nil
}

func (s *OrderServer) ListOrders(ctx context.Context, _ *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	log.Println("[Order Service] ListOrders request")
	rows, err := s.db.QueryContext(ctx, "SELECT id, user_id, product_name, quantity, status FROM orders")
	if err != nil {
		log.Printf("[Order Service] ListOrders database error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list orders: %v", err)
	}
	defer rows.Close()

	var orders []*pb.Order
	for rows.Next() {
		var o pb.Order
		if err := rows.Scan(&o.Id, &o.UserId, &o.ProductName, &o.Quantity, &o.Status); err != nil {
			log.Printf("[Order Service] ListOrders scan error: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to scan order: %v", err)
		}
		orders = append(orders, &o)
	}
	log.Printf("[Order Service] ListOrders success: returned %d orders", len(orders))
	return &pb.ListOrdersResponse{Orders: orders}, nil
}

func (s *OrderServer) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	log.Printf("[Order Service] UpdateOrder request: ID=%s", req.GetId())
	var current pb.Order
	err := s.db.QueryRowContext(ctx, "SELECT product_name, quantity, status, user_id FROM orders WHERE id = $1", req.GetId()).Scan(&current.ProductName, &current.Quantity, &current.Status, &current.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[Order Service] UpdateOrder info: order %q not found", req.GetId())
			return nil, status.Errorf(codes.NotFound, "order %q not found", req.GetId())
		}
		log.Printf("[Order Service] UpdateOrder database query error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to query order: %v", err)
	}

	pname := current.ProductName
	if req.GetProductName() != "" {
		pname = req.GetProductName()
	}
	qty := current.Quantity
	if req.GetQuantity() > 0 {
		qty = req.GetQuantity()
	}
	st := current.Status
	if req.GetStatus() != "" {
		st = req.GetStatus()
	}

	_, err = s.db.ExecContext(ctx, "UPDATE orders SET product_name = $1, quantity = $2, status = $3 WHERE id = $4", pname, qty, st, req.GetId())
	if err != nil {
		log.Printf("[Order Service] UpdateOrder database exec error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update order: %v", err)
	}

	log.Printf("[Order Service] UpdateOrder success: ID=%s status=%s", req.GetId(), st)
	return &pb.UpdateOrderResponse{Order: &pb.Order{Id: req.GetId(), UserId: current.UserId, ProductName: pname, Quantity: qty, Status: st}}, nil
}

func (s *OrderServer) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	log.Printf("[Order Service] DeleteOrder request: ID=%s", req.GetId())
	res, err := s.db.ExecContext(ctx, "DELETE FROM orders WHERE id = $1", req.GetId())
	if err != nil {
		log.Printf("[Order Service] DeleteOrder database error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete order: %v", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		log.Printf("[Order Service] DeleteOrder info: order %q not found", req.GetId())
		return nil, status.Errorf(codes.NotFound, "order %q not found", req.GetId())
	}
	log.Printf("[Order Service] DeleteOrder success: ID=%s", req.GetId())
	return &pb.DeleteOrderResponse{Message: "order \"" + req.GetId() + "\" deleted successfully"}, nil
}
