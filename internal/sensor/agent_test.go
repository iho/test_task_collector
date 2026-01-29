package sensor

import (
	"context"
	"testing"
	"time"

	pb "github.com/iho/test_task_collector/proto/telemetry"
	"google.golang.org/grpc"
)

// FakeClient implements pb.TelemetryServiceClient
type FakeClient struct {
	Published []*pb.Metric
}

func (f *FakeClient) Publish(_ context.Context, in *pb.Metric, _ ...grpc.CallOption) (*pb.PublishResponse, error) {
	f.Published = append(f.Published, in)
	return &pb.PublishResponse{}, nil
}

func TestAgent_Start(t *testing.T) {
	cfg := &Config{
		Rate: 10,
		Name: "test-sensor",
	}

	fakeClient := &FakeClient{}

	agent := &Agent{
		cfg:    cfg,
		client: fakeClient,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	agent.Start(ctx)

	count := len(fakeClient.Published)
	if count < 2 || count > 3 {
		t.Errorf("Expected 2 or 3 messages, got %d", count)
	}

	if count > 0 {
		m := fakeClient.Published[0]
		if m.Name != "test-sensor" {
			t.Errorf("Expected sensor name 'test-sensor', got %s", m.Name)
		}
		if m.Value < 0 || m.Value >= 100 {
			t.Errorf("Expected value between 0 and 100, got %d", m.Value)
		}
	}
}
