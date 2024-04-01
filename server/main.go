package main

import (
	"context"
	"log"
	"net"

	pb "github.com/danieljcksn/connect-four/proto"

	"google.golang.org/grpc"
)

type gameServer struct {
	pb.UnimplementedGameServiceServer
}

func (s *gameServer) Connect(
	ctx context.Context,
	req *pb.ConnectRequest,
) (*pb.ConnectResponse, error) {
	return &pb.ConnectResponse{
		Message: "Connected to the server!",
	}, nil
}

func newGameServer() *gameServer {
	s := &gameServer{}

	return s
}

func main() {
	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGameServiceServer(grpcServer, newGameServer())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
