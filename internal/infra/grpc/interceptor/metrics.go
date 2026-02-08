package interceptor

import (
	"context"
	"sync/atomic"
	"time"

	"erp.localhost/internal/infra/logging/logger"
	"google.golang.org/grpc"
)

// latencyBuckets defines the thresholds for latency buckets in milliseconds.
// Buckets are cumulative: bucket[i] counts all requests with latency <= threshold[i].
var latencyBuckets = [...]float64{1, 5, 10, 25, 50, 100, 250, 500}

// Compile-time assertion to ensure latencyBuckets has exactly 8 elements
var _ = [len(latencyBuckets) - 8]struct{}{}

// MetricsSnapshot represents a point-in-time snapshot of metrics.
type MetricsSnapshot struct {
	TotalRequests  int64
	FailedRequests int64
	LatencyBuckets [8]int64 // cumulative: bucket[i] = count of requests with latency <= threshold[i]
}

// MetricsCollector collects basic request metrics using atomic counters.
// Safe for concurrent use.
type MetricsCollector struct {
	totalRequests  atomic.Int64
	failedRequests atomic.Int64
	latencyBuckets [8]atomic.Int64
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// Snapshot returns a point-in-time snapshot of all metrics.
func (m *MetricsCollector) Snapshot() MetricsSnapshot {
	snapshot := MetricsSnapshot{
		TotalRequests:  m.totalRequests.Load(),
		FailedRequests: m.failedRequests.Load(),
	}

	for i := range m.latencyBuckets {
		snapshot.LatencyBuckets[i] = m.latencyBuckets[i].Load()
	}

	return snapshot
}

// recordRequest records metrics for a completed request.
func (m *MetricsCollector) recordRequest(durationMs float64, failed bool) {
	m.totalRequests.Add(1)

	if failed {
		m.failedRequests.Add(1)
	}

	// Record latency in buckets (cumulative)
	for i, threshold := range latencyBuckets {
		if durationMs <= threshold {
			m.latencyBuckets[i].Add(1)
		}
	}
}

// ServerMetricsInterceptor creates a server-side metrics collection interceptor.
// If collector is nil, the interceptor is a no-op.
func ServerMetricsInterceptor(collector *MetricsCollector, log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// If no collector provided, pass through
		if collector == nil {
			return handler(ctx, req)
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := float64(time.Since(start)) / float64(time.Millisecond)

		collector.recordRequest(duration, err != nil)

		return resp, err
	}
}

// ClientMetricsInterceptor creates a client-side metrics collection interceptor.
// If collector is nil, the interceptor is a no-op.
func ClientMetricsInterceptor(collector *MetricsCollector, log logger.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// If no collector provided, pass through
		if collector == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := float64(time.Since(start)) / float64(time.Millisecond)

		collector.recordRequest(duration, err != nil)

		return err
	}
}
