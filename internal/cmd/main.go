package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	pb "identity-engine/proto"
)

// server is used to implement identity.IdentityServiceServer.
type server struct {
	pb.UnimplementedIdentityServiceServer
	mu     sync.Mutex
	users  map[string]*pb.User
	nextID int
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	user := &pb.User{
		Id:       fmt.Sprintf("%d", s.nextID),
		Username: req.Username,
		Email:    req.Email,
	}
	s.users[user.Id] = user
	return user, nil
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	return user, nil
}

func (s *server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.User == nil {
		return nil, status.Errorf(codes.InvalidArgument, "user cannot be nil")
	}

	_, exists := s.users[req.User.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	s.users[req.User.Id] = req.User
	return req.User, nil
}

func (s *server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.users[req.Id]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	delete(s.users, req.Id)
	return &pb.DeleteUserResponse{Success: true}, nil
}

func (s *server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var users []*pb.User
	for _, user := range s.users {
		users = append(users, user)
	}

	return &pb.ListUsersResponse{Users: users}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterIdentityServiceServer(s, &server{
		users: make(map[string]*pb.User),
	})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	fmt.Println("Identity gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
