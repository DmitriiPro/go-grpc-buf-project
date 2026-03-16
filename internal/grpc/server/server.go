package grpc_server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	newsv1 "news/buf/grpc/api/news/v1"
	"news/buf/grpc/internal/memstore"

	"buf.build/go/protovalidate"
	"github.com/google/uuid"
	// "github.com/google/uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	newsv1.UnimplementedNewsServiceServer
	validator protovalidate.Validator
	store     memstore.MemStore
}

func NewServer(store memstore.MemStore) (*Server, error) {
	validator, err := protovalidate.New()

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create validator: %v", err)
	}

	return &Server{
		validator: validator,
		store:     store,
	}, nil
}

func (s *Server) Validator(req interface{}) error {
	if v, ok := req.(interface{ Validate() error }); ok {
		return v.Validate()
	}

	return errors.New("invalid request")
}


func (s *Server) DeleteNews(stream newsv1.NewsService_DeleteNewsServer) error {


	for {
		in, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		newsUUID, err := uuid.Parse(in.Id)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		s.store.DeleteNews(newsUUID)
		fmt.Printf("delete news: %v", in)
		if err = stream.Send(&newsv1.NewsIdResponse{Id: newsUUID.String()}); err != nil {
			return err
		}

	}
}

func (s *Server) UpdateNews(stream newsv1.NewsService_UpdateNewsServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			return err
		}

		if err := s.Validator(req); err != nil {
			log.Printf("Validation failed for UpdateNews request: %v", err)
			return status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
		}

		fmt.Printf("update news: %v", req)
		s.store.UpdateNews(req)

	}

}

func (s *Server) GetAll(in *emptypb.Empty, stream newsv1.NewsService_GetAllServer) error {

	for _, news := range s.store.GetAll() {
		response := &newsv1.NewsServiceGetResponse{
			Id:        news.GetId(),
			Author:    news.GetAuthor(),
			Title:     news.GetTitle(),
			Summary:   news.GetSummary(),
			Content:   news.GetContent(),
			Source:    news.GetSource(),
			Tags:      news.GetTags(),
			CreatedAt: news.GetCreatedAt(),
			UpdatedAt: news.GetUpdatedAt(),
			DeletedAt: news.GetDeletedAt(),
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Create(_ context.Context, req *newsv1.NewsServiceCreateRequest) (*newsv1.NewsServiceCreateResponse, error) {

	if err := s.Validator(req); err != nil {
		log.Printf("Validation failed for Create request: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	now := timestamppb.Now()
	// newID := uuid.New().String()

	response := &newsv1.NewsServiceCreateResponse{
		Id:        req.GetId(),
		Author:    req.GetAuthor(),
		Title:     req.GetTitle(),
		Summary:   req.GetSummary(),
		Content:   req.GetContent(),
		Source:    req.GetSource(),
		Tags:      req.GetTags(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	s.store.Create(response)

	return response, nil

}

func (s *Server) Get(_ context.Context, req *newsv1.NewsServiceGetRequest) (*newsv1.NewsServiceGetResponse, error) {
	if err := s.Validator(req); err != nil {
		log.Printf("Validation failed for Get request: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	news, err := s.store.Get(req.GetId())

	if err != nil {
		log.Printf("News not found: %v", err)
		return nil, status.Errorf(codes.NotFound, "news not found: %v", err)
	}

	response := &newsv1.NewsServiceGetResponse{
		Id:        news.GetId(),
		Author:    news.GetAuthor(),
		Title:     news.GetTitle(),
		Summary:   news.GetSummary(),
		Content:   news.GetContent(),
		Source:    news.GetSource(),
		Tags:      news.GetTags(),
		CreatedAt: news.GetCreatedAt(),
		UpdatedAt: news.GetUpdatedAt(),
		DeletedAt: news.GetDeletedAt(),
	}

	return response, nil
}
