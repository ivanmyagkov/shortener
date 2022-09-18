package grpc

import (
	"context"

	"github.com/ivanmyagkov/shortener.git/internal/grpc/handlers"
	pb "github.com/ivanmyagkov/shortener.git/internal/grpc/proto"
)

type gRPCServer struct {
	pb.UnimplementedShortenerServer
	grpcHandler *handlers.GRPCHandler
}

// NewGRPCServer returns a gRPCServer object
func NewGRPCServer(grpcHandler *handlers.GRPCHandler) (server *gRPCServer, err error) {
	return &gRPCServer{grpcHandler: grpcHandler}, nil
}

func (s *gRPCServer) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	resp, err := s.grpcHandler.GetPingDB()

	return resp, err
}

func (s *gRPCServer) GetStats(_ context.Context, _ *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	resp, err := s.grpcHandler.GetStats()

	return resp, err
}

func (s *gRPCServer) GetURL(_ context.Context, request *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	result, err := s.grpcHandler.GetURL(request)
	return result, err
}

func (s *gRPCServer) PostURL(ctx context.Context, request *pb.PostURLRequest) (*pb.PostURLResponse, error) {
	result, err := s.grpcHandler.PostURL(ctx, request)
	return result, err
}

func (s *gRPCServer) GetURLsByUserID(ctx context.Context, _ *pb.GetURLsByUserIDRequest) (*pb.GetURLsByUserIDResponse, error) {
	result, err := s.grpcHandler.GetURLsByUserID(ctx)
	return result, err
}
func (s *gRPCServer) PostURLBatch(ctx context.Context, request *pb.PostURLBatchRequest) (*pb.PostURLBatchResponse, error) {
	result, err := s.grpcHandler.PostURLBatch(ctx, request)
	return result, err
}
func (s *gRPCServer) DeleteURLBatch(ctx context.Context, request *pb.DeleteURLBatchRequest) (*pb.DeleteURLBatchResponse, error) {
	result, err := s.grpcHandler.DeleteURLBatch(ctx, request)
	return result, err
}
