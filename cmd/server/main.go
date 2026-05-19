package main

import (
	"log"
	"net"
	"github.com/hibiken/asynq"
	"google.golang.org/grpc"

	"github.com/example/brainway/internal/handler"
	"github.com/example/brainway/pb"
)

func main() {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	defer client.Close()

	h := handler.New(client)

	grpcServer := grpc.NewServer()
	pb.RegisterTransactionServiceServer(grpcServer, h)

	addr := ":50051"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}
	log.Printf("TransactionService listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve error: %v", err)
	}
}
