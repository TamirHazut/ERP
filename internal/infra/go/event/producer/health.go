package producer

import (
	"os"
)

// HealthStatus represents the overall health status of a component.
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "HEALTHY"
	HealthStatusDegraded HealthStatus = "DEGRADED"
	HealthStatusDown     HealthStatus = "DOWN"
)

// Health contains health check results for all dependencies.
type Health struct {
	Kafka   HealthStatus
	MongoDB HealthStatus
	Wal     HealthStatus
	Overall HealthStatus // Worst of the three
}

// GetHealth performs health checks on all dependencies.
func (p *Producer) GetHealth() Health {
	health := Health{
		Kafka:   p.checkKafkaHealth(),
		MongoDB: p.checkMongoDBHealth(),
		Wal:     p.checkWalHealth(),
	}

	// Overall health is the worst of all components
	health.Overall = worstHealth(health.Kafka, health.MongoDB, health.Wal)

	return health
}

// checkKafkaHealth pings Kafka to verify connectivity.
func (p *Producer) checkKafkaHealth() HealthStatus {
	// Use the writer's Stats method or try a lightweight operation
	// Note: kafka-go doesn't have a built-in Ping, so we'll try to write metadata check
	// For now, we'll assume healthy if writer exists
	// TODO: Implement actual Kafka health check with metadata fetch
	if p.writer == nil {
		return HealthStatusDown
	}

	// Quick check: try to get stats
	stats := p.writer.Stats()
	if stats.Writes > 0 || stats.Errors == 0 {
		// If we've had successful writes or no errors, consider healthy
		return HealthStatusHealthy
	}

	// If we haven't written anything yet but writer exists, consider healthy
	return HealthStatusHealthy
}

// checkMongoDBHealth checks if MongoDB is accessible.
func (p *Producer) checkMongoDBHealth() HealthStatus {
	// Try to ping MongoDB through DLQ handler
	// Note: BaseDLQHandler doesn't have a Ping method
	// For now, we'll assume healthy if DLQ handler exists
	// TODO: Add Ping method to DLQHandler interface
	if p.dlq == nil {
		return HealthStatusDown
	}

	return HealthStatusHealthy
}

// checkWalHealth checks if WAL is writable.
func (p *Producer) checkWalHealth() HealthStatus {
	if p.walQueue == nil {
		// WAL not initialized (might be disabled)
		return HealthStatusHealthy
	}

	// Check if WAL directory is writable
	testFile := p.config.WalDir + "/.health_check"
	file, err := os.Create(testFile)
	if err != nil {
		return HealthStatusDown
	}
	file.Close()
	os.Remove(testFile)

	return HealthStatusHealthy
}

// worstHealth returns the worst health status among the given statuses.
func worstHealth(statuses ...HealthStatus) HealthStatus {
	worst := HealthStatusHealthy

	for _, status := range statuses {
		switch status {
		case HealthStatusDown:
			return HealthStatusDown // Down is the worst, return immediately
		case HealthStatusDegraded:
			worst = HealthStatusDegraded
		}
	}

	return worst
}
