package main

import (
	"fmt"
	"log"
	"net"

	newsv1 "news/buf/grpc/api/news/v1"
	grpc_server "news/buf/grpc/internal/grpc/server"
	"news/buf/grpc/internal/memstore"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	PORT_GRPC_SERVER = 50051
)

func main() {
	startGRPCServer(PORT_GRPC_SERVER)
}

func startGRPCServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	newsServer, err := grpc_server.NewServer(memstore.NewStore())
	if err != nil {
		log.Fatalf("failed to create news server: %v", err)
	}

	log.Println("Starting gRPC server on port", port)

	grpcServer := grpc.NewServer(opts...)
	newsv1.RegisterNewsServiceServer(grpcServer, newsServer)

	healthServer := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthServer)

	if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("failed to serve: %v", err)
	}

}
