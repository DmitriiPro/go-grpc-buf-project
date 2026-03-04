package main

import (
	"fmt"
	"log"
	"net"

	newsv1 "news/buf/grpc/api/news/v1"
	grpc_server "news/buf/grpc/internal/grpc/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	newsv1.RegisterNewsServiceServer(grpcServer, grpc_server.NewServer())

	healthServer := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthServer)

	if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("failed to serve: %v", err)
	}
}
