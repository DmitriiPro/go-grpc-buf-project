package grpc_server

import newsv1 "news/buf/grpc/api/news/v1"

type Server struct {
	newsv1.UnimplementedNewsServiceServer
}

func NewServer() *Server {
	return &Server{}
}