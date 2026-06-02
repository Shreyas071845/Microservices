package grpc_server

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	_ "github.com/lib/pq"
	pb "user_service/internal/proto"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	db *sql.DB
}

func NewUserServer() *UserServer {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://devuser:devpassword@localhost:5432/assignment_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("User service: failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("User service: failed to ping database: %v", err)
	}

	// Create users table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL
		);
	`)
	if err != nil {
		log.Fatalf("User service: failed to create users table: %v", err)
	}

	log.Println("User service connected to PostgreSQL successfully")
	return &UserServer{db: db}
}

func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log.Printf("[User Service] CreateUser request: name=%q, email=%q", req.GetName(), req.GetEmail())
	id := uuid.New().String()
	user := &pb.User{Id: id, Name: req.GetName(), Email: req.GetEmail()}
	
	_, err := s.db.ExecContext(ctx, "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)", id, user.Name, user.Email)
	if err != nil {
		log.Printf("[User Service] CreateUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert user: %v", err)
	}
	log.Printf("[User Service] CreateUser success: ID=%s", id)
	return &pb.CreateUserResponse{User: user}, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log.Printf("[User Service] GetUser request: ID=%s", req.GetId())
	var user pb.User
	err := s.db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = $1", req.GetId()).Scan(&user.Id, &user.Name, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[User Service] GetUser info: user %q not found", req.GetId())
			return nil, status.Errorf(codes.NotFound, "user %q not found", req.GetId())
		}
		log.Printf("[User Service] GetUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to query user: %v", err)
	}
	log.Printf("[User Service] GetUser success: name=%q", user.Name)
	return &pb.GetUserResponse{User: &user}, nil
}

func (s *UserServer) ListUsers(ctx context.Context, _ *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	log.Println("[User Service] ListUsers request")
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, email FROM users")
	if err != nil {
		log.Printf("[User Service] ListUsers error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}
	defer rows.Close()

	var users []*pb.User
	for rows.Next() {
		var u pb.User
		if err := rows.Scan(&u.Id, &u.Name, &u.Email); err != nil {
			log.Printf("[User Service] ListUsers scan error: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to scan user: %v", err)
		}
		users = append(users, &u)
	}
	log.Printf("[User Service] ListUsers success: returned %d users", len(users))
	return &pb.ListUsersResponse{Users: users}, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	log.Printf("[User Service] UpdateUser request: ID=%s", req.GetId())
	var current pb.User
	err := s.db.QueryRowContext(ctx, "SELECT name, email FROM users WHERE id = $1", req.GetId()).Scan(&current.Name, &current.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[User Service] UpdateUser info: user %q not found", req.GetId())
			return nil, status.Errorf(codes.NotFound, "user %q not found", req.GetId())
		}
		log.Printf("[User Service] UpdateUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to query user: %v", err)
	}

	name := current.Name
	if req.GetName() != "" {
		name = req.GetName()
	}
	email := current.Email
	if req.GetEmail() != "" {
		email = req.GetEmail()
	}

	_, err = s.db.ExecContext(ctx, "UPDATE users SET name = $1, email = $2 WHERE id = $3", name, email, req.GetId())
	if err != nil {
		log.Printf("[User Service] UpdateUser exec error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	log.Printf("[User Service] UpdateUser success: ID=%s", req.GetId())
	return &pb.UpdateUserResponse{User: &pb.User{Id: req.GetId(), Name: name, Email: email}}, nil
}

func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	log.Printf("[User Service] DeleteUser request: ID=%s", req.GetId())
	res, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", req.GetId())
	if err != nil {
		log.Printf("[User Service] DeleteUser error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		log.Printf("[User Service] DeleteUser info: user %q not found", req.GetId())
		return nil, status.Errorf(codes.NotFound, "user %q not found", req.GetId())
	}
	log.Printf("[User Service] DeleteUser success: ID=%s", req.GetId())
	return &pb.DeleteUserResponse{Message: "user \"" + req.GetId() + "\" deleted successfully"}, nil
}
