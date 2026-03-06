package main

import (
	"context"
	"fmt"
	"log"
	newsv1 "news/buf/grpc/api/news/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	PORT_GRPC_CLIENT = 50051
)

func startGRPCClient(port string) *grpc.ClientConn {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(port, opts...)
	if err != nil {
		log.Fatal("failed to connect GRPC Client: %v", err)
	}

	return conn
}

func main() {
	port := fmt.Sprintf("localhost:%d", PORT_GRPC_CLIENT)
	conn := startGRPCClient(port)
	defer conn.Close()

	log.Println("Starting Client GRPC ", port)

	client := newsv1.NewNewsServiceClient(conn)

	newId := uuid.New().String()
	ctx := context.Background()

	resp, err := client.Create(
		ctx,
		&newsv1.NewsServiceCreateRequest{
			Id:      newId,
			Title:   "Breaking News",
			Content: "This is the content of the breaking news.",
			Author:  "John Doe",
			Summary: "This is a summary of the breaking news.",
			Source:  "News Agency",
			Tags:    []string{"breaking", "news", "world"},
		},
	)

	if err != nil {
		log.Fatalf("failed to create news: %v", err)
	}

	log.Printf("News created: %v", resp)

	getNews, err := client.Get(
		ctx,
		&newsv1.NewsServiceGetRequest{
			Id: newId,
		},
	)
	if err != nil {
		log.Fatalf("failed to get news: %v", err)
	}

	log.Printf("News retrieved: %v", getNews)

}
