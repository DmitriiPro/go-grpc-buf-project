package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	newsv1 "news/buf/grpc/api/news/v1"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"buf.build/go/protovalidate"
)

const (
	PORT_GRPC_CLIENT = 50051
)

func unaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	end := time.Now()

	log.Printf("Unary Client Interceptor: %s, start time: %s, end time: %s, err: %v", method, start.Format("Basic"), end.Format(time.RFC3339), err)
	return err
}

func streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Printf("Stream Client Interceptor: %s", method)
	return streamer(ctx, desc, cc, method, opts...)
}

func startGRPCClient(port string) *grpc.ClientConn {
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryInterceptor),
		grpc.WithStreamInterceptor(streamInterceptor),
	)

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

	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("failed to create validator: %v", err)
	}

	for i := 0; i < 5; i++ {
		newId := uuid.New().String()
		msg := &newsv1.NewsServiceCreateRequest{
			Id:      newId,
			Title:   fmt.Sprintf("Breaking News %d", i),
			Content: fmt.Sprintf("This is the content of the breaking news. %d", i),
			Author:  fmt.Sprintf("John Doe %d", i),
			Summary: fmt.Sprintf("This is a summary of the breaking news. %d", i),
			Source:  fmt.Sprintf("News Agency %d", i),
			Tags:    []string{"breaking", "news", "world"},
		}
		if err := validator.Validate(msg); err != nil {
			log.Fatalf("Validation failed for Create request: %v", err)
		}

		_, err := client.Create(
			ctx,
			msg,
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
	// log.Printf("allNews: %v", allNews)

	//* Client-side streaming RPC
	var streamUpdateNews grpc.ClientStreamingClient[newsv1.NewsServiceCreateRequest, emptypb.Empty]

	streamUpdateNews, err = client.UpdateNews(ctx)
	if err != nil {
		log.Fatalf("failed to update news: %v", err)
	}

	for i := 0; i < 5; i++ {
		if err := streamUpdateNews.Send(
			&newsv1.NewsServiceCreateRequest{
				Id:      uuid.New().String(),
				Title:   fmt.Sprintf("Breaking News %d", i),
				Content: fmt.Sprintf("This is the content of the breaking news. %d", i),
				Author:  fmt.Sprintf("John Doe %d", i),
				Summary: fmt.Sprintf("This is a summary of the breaking news. %d", i),
				Source:  fmt.Sprintf("News Agency %d", i),
				Tags:    []string{"breaking", "news", "world"},
			},
		); err != nil {
			log.Fatalf("failed to send news: %v", err)
		}
	}

	if err := streamUpdateNews.CloseSend(); err != nil {
		log.Fatalf("failed to close send: %v", err)
	}

	streamGetAll, err = client.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("failed to streaming get all news: %v", err)
	}

	// allNews := make([]*newsv1.NewsServiceGetResponse, 0)

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

	//* bidirectional streaming Client
	streamDeleteNews, err := client.DeleteNews(ctx)
	if err != nil {
		log.Fatalf("failed to delete news: %v", err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := streamDeleteNews.Recv()
			if errors.Is(err, io.EOF) {
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("Failed to receive a news: %v", err)
			}

			log.Printf("DeleteNews: %v", in)
		}
	}()

	for _, news := range allNews {
		if err := streamDeleteNews.Send(&newsv1.NewsID{Id: news.GetId()}); err != nil {
			log.Fatalf("failed to send news: %v", err)
		}
	}

	streamDeleteNews.CloseSend()
	<-waitc

}
