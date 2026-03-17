package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	newsv1 "news/buf/grpc/api/news/v1"
	grpc_server "news/buf/grpc/internal/grpc/server"
	"news/buf/grpc/internal/memstore"

	"buf.build/go/protovalidate"
	protvalidate_interceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"golang.org/x/sync/errgroup"
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
	// lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	// if err != nil {
	// 	log.Fatalf("failed to listen: %v", err)
	// }
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
	opts = append(opts,
		grpc.ChainUnaryInterceptor(
			protvalidate_interceptor.UnaryServerInterceptor(validator),
			func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				start := time.Now()
				resp, err = handler(ctx, req)
				log.Printf("Unary Interceptor server: %+v", info)
				log.Printf("time taken: %s", time.Since(start))
				return
			}, func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
				log.Printf("Second interceptor")
				return handler(ctx, req)
			}),
		grpc.ChainStreamInterceptor(
			func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				log.Printf("Server Side Stream Interceptor: %+v", info)
				return handler(srv, ss)
			}, protvalidate_interceptor.StreamServerInterceptor(validator)),
			
	)


	grpcServer := grpc.NewServer(opts...)
	newsv1.RegisterNewsServiceServer(grpcServer, newsServer)

	healthServer := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthServer)
	// healthServer.SetServingStatus("service-name", healthv1.HealthCheckResponse_SERVING)

	// if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
	// 	log.Fatalf("failed to serve: %v", err)
	// }

	//* Graceful Shutdown
	ctx := context.Background()
	grp, grpCtx := errgroup.WithContext(ctx)

	grp.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()

		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			err = fmt.Errorf("failed to listen: %v", err)
		}

		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			err = fmt.Errorf("failed to serve: %v", err)
		}

		return err
	})

	grp.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()

		interceptSignals(grpCtx)
		healthServer.Shutdown()

		return shutdown(grpCtx, grpcServer)
	})

	if err := grp.Wait(); err != nil {
		log.Fatalf("server shutdwon: %v", err)
	}

}

func interceptSignals(ctx context.Context) {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-ctx.Done():
		return
	case s := <-sig:
		log.Printf("intercepted signal received: %s", s.String())
		return
	}
}

func shutdown(ctx context.Context, srv *grpc.Server) (err error) {
	done := make(chan struct{}, 1)

	go func() {
		srv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		err = fmt.Errorf("grpc server forcibly shutdown: %v", ctx.Err())
		srv.Stop()
	}

	return
}
