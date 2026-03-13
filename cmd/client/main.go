package main

import (
	"context"
	"fmt"
	"io"
	"log"
	newsv1 "news/buf/grpc/api/news/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
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

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		newId := uuid.New().String()

		_, err := client.Create(
			ctx,
			&newsv1.NewsServiceCreateRequest{
				Id:      newId,
				Title:   fmt.Sprintf("Breaking News %d", i),
				Content: fmt.Sprintf("This is the content of the breaking news. %d", i),
				Author:  fmt.Sprintf("John Doe %d", i),
				Summary: fmt.Sprintf("This is a summary of the breaking news. %d", i),
				Source:  fmt.Sprintf("News Agency %d", i),
				Tags:    []string{"breaking", "news", "world"},
			},
		)

		if err != nil {
			log.Fatalf("failed to create news: %v", err)
		}
	}

	// log.Printf("News created: %v", resp)
	/*
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
	*/

	// *Server-side streaming RPC
	streamGetAll, err := client.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("failed to streaming get all news: %v", err)
	}

	allNews := make([]*newsv1.NewsServiceGetResponse, 0)

	for {
		feature, err := streamGetAll.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.GetAll = _, %v", client, err)
		}
		log.Printf("feature: %v\n", feature)

		allNews = append(allNews, feature)

	}
	log.Printf("allNews: %v", allNews)

}
