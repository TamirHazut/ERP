package interceptor

import (
	"context"
	"errors"
	"testing"

	"erp.localhost/internal/infra/logging/logger/mock"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

func TestServerMetricsInterceptor_NilCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := mock.NewMockLogger(ctrl)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	interceptor := ServerMetricsInterceptor(nil, log)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	resp, err := interceptor(context.Background(), "request", info, handler)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if resp != "response" {
		t.Errorf("expected response, got: %v", resp)
	}
}

func TestServerMetricsInterceptor_SuccessfulRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := mock.NewMockLogger(ctrl)

	collector := NewMetricsCollector()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	interceptor := ServerMetricsInterceptor(collector, log)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, err := interceptor(context.Background(), "request", info, handler)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	snapshot := collector.Snapshot()
	if snapshot.TotalRequests != 1 {
		t.Errorf("expected TotalRequests=1, got: %d", snapshot.TotalRequests)
	}
	if snapshot.FailedRequests != 0 {
		t.Errorf("expected FailedRequests=0, got: %d", snapshot.FailedRequests)
	}

	// At least one latency bucket should be incremented
	hasLatency := false
	for _, count := range snapshot.LatencyBuckets {
		if count > 0 {
			hasLatency = true
			break
		}
	}
	if !hasLatency {
		t.Error("expected at least one latency bucket to be incremented")
	}
}

func TestServerMetricsInterceptor_FailedRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := mock.NewMockLogger(ctrl)

	collector := NewMetricsCollector()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New("handler error")
	}

	interceptor := ServerMetricsInterceptor(collector, log)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, err := interceptor(context.Background(), "request", info, handler)
	if err == nil {
		t.Error("expected error from handler")
	}

	snapshot := collector.Snapshot()
	if snapshot.TotalRequests != 1 {
		t.Errorf("expected TotalRequests=1, got: %d", snapshot.TotalRequests)
	}
	if snapshot.FailedRequests != 1 {
		t.Errorf("expected FailedRequests=1, got: %d", snapshot.FailedRequests)
	}
}

func TestServerMetricsInterceptor_MultipleRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := mock.NewMockLogger(ctrl)

	collector := NewMetricsCollector()
	successHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	failHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New("error")
	}

	interceptor := ServerMetricsInterceptor(collector, log)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	// 3 successful requests
	for i := 0; i < 3; i++ {
		interceptor(context.Background(), "request", info, successHandler)
	}

	// 2 failed requests
	for i := 0; i < 2; i++ {
		interceptor(context.Background(), "request", info, failHandler)
	}

	snapshot := collector.Snapshot()
	if snapshot.TotalRequests != 5 {
		t.Errorf("expected TotalRequests=5, got: %d", snapshot.TotalRequests)
	}
	if snapshot.FailedRequests != 2 {
		t.Errorf("expected FailedRequests=2, got: %d", snapshot.FailedRequests)
	}
}

func TestClientMetricsInterceptor_SuccessfulRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := mock.NewMockLogger(ctrl)

	collector := NewMetricsCollector()
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	interceptor := ClientMetricsInterceptor(collector, log)

	err := interceptor(context.Background(), "/test.Service/Method", "request", "reply", nil, invoker)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	snapshot := collector.Snapshot()
	if snapshot.TotalRequests != 1 {
		t.Errorf("expected TotalRequests=1, got: %d", snapshot.TotalRequests)
	}
	if snapshot.FailedRequests != 0 {
		t.Errorf("expected FailedRequests=0, got: %d", snapshot.FailedRequests)
	}
}

func TestClientMetricsInterceptor_NilCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := mock.NewMockLogger(ctrl)

	invokerCalled := false
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		invokerCalled = true
		return nil
	}

	interceptor := ClientMetricsInterceptor(nil, log)

	err := interceptor(context.Background(), "/test.Service/Method", "request", "reply", nil, invoker)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !invokerCalled {
		t.Error("invoker was not called")
	}
}

func TestMetricsSnapshot(t *testing.T) {
	collector := NewMetricsCollector()

	// Manually record some metrics
	collector.recordRequest(5.0, false)
	collector.recordRequest(50.0, true)
	collector.recordRequest(150.0, false)

	snapshot := collector.Snapshot()

	if snapshot.TotalRequests != 3 {
		t.Errorf("expected TotalRequests=3, got: %d", snapshot.TotalRequests)
	}
	if snapshot.FailedRequests != 1 {
		t.Errorf("expected FailedRequests=1, got: %d", snapshot.FailedRequests)
	}

	// Verify latency buckets are cumulative
	// 5.0ms should hit buckets: 5, 10, 25, 50, 100, 250, 500 (buckets 1-7)
	// 50.0ms should hit buckets: 50, 100, 250, 500 (buckets 4-7)
	// 150.0ms should hit buckets: 250, 500 (buckets 6-7)

	expectedBuckets := map[int]int64{
		0: 0, // 1ms: 0 hits (all requests > 1ms)
		1: 1, // 5ms: 1 hit (5.0)
		2: 1, // 10ms: 1 hit (5.0)
		3: 1, // 25ms: 1 hit (5.0)
		4: 2, // 50ms: 2 hits (5.0, 50.0)
		5: 2, // 100ms: 2 hits (5.0, 50.0)
		6: 3, // 250ms: 3 hits (5.0, 50.0, 150.0)
		7: 3, // 500ms: 3 hits (5.0, 50.0, 150.0)
	}

	for i, expected := range expectedBuckets {
		if snapshot.LatencyBuckets[i] != expected {
			t.Errorf("bucket[%d] (threshold %.0fms): expected %d, got: %d",
				i, latencyBuckets[i], expected, snapshot.LatencyBuckets[i])
		}
	}
}
