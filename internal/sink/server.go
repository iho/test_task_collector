package sink

import (
	"context"

	pb "github.com/iho/test_task_collector/proto/telemetry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Server implements the gRPC TelemetryService.
type Server struct {
	pb.UnimplementedTelemetryServiceServer
	writer  *BufferedWriter
	limiter *RateLimiter
}

// NewServer creates a new TelemetryService server.
func NewServer(writer *BufferedWriter, limiter *RateLimiter) *Server {
	return &Server{
		writer:  writer,
		limiter: limiter,
	}
}

// Publish receives a metric and writes it to the log.
func (s *Server) Publish(_ context.Context, metric *pb.Metric) (*pb.PublishResponse, error) {
	size := proto.Size(metric)

	if !s.limiter.Allow(size) {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
	}

	data, err := protojson.Marshal(metric)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal metric: %v", err)
	}

	if err := s.writer.Write(data); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write to log: %v", err)
	}

	return &pb.PublishResponse{}, nil
}
