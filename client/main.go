package main

import (
	"context"
	"log"
	"time"

	pb "github.com/danieljcksn/connect-four/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("failed to connect to gRPC server at localhost:50051: %v", err)
	}

	defer conn.Close()

	client := pb.NewGameServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	r, err := client.Connect(ctx, &pb.ConnectRequest{})

	if err != nil {
		log.Fatalf("error calling function SayHello: %v", err)
	}

	log.Printf("Response from gRPC server's SayHello function: %s", r.GetMessage())
}
