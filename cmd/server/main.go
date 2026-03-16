package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	newsv1 "news/buf/grpc/api/news/v1"
	grpc_server "news/buf/grpc/internal/grpc/server"
	"news/buf/grpc/internal/memstore"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"buf.build/go/protovalidate"
	protvalidate_interceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
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


	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("failed to create validator: %v", err)
	}



	newsServer, err := grpc_server.NewServer(memstore.NewStore())
	if err != nil {
		log.Fatalf("failed to create news server: %v", err)
	}

	log.Println("Starting gRPC server on port", port)

	// Interceptors
	/*
	   1 Server Side Unary Interceptors
	   2 Client Side Unary Interceptors
	   3 Server Side Stream Interceptors
	   4 Client Side Stream Interceptors
	*/

	// grpc.UnaryInterceptor
	// grpc.ChainUnaryInterceptor
	opts = append(opts, grpc.ChainUnaryInterceptor(
		protvalidate_interceptor.UnaryServerInterceptor(validator),
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		log.Printf("Unary Interceptor server: %+v", info)
		log.Printf("time taken: %s", time.Since(start))
		return
	}, func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any,  error) {
		log.Printf("Second interceptor")
		return handler(ctx, req)
	}), grpc.ChainStreamInterceptor(func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler ) error  {
		log.Printf("Server Side Stream Interceptor: %+v", info)
		return handler(srv, ss)
	}, protvalidate_interceptor.StreamServerInterceptor(validator)))

	grpcServer := grpc.NewServer(opts...)
	newsv1.RegisterNewsServiceServer(grpcServer, newsServer)

	healthServer := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthServer)

	if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("failed to serve: %v", err)
	}

}
