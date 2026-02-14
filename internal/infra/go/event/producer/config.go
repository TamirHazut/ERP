package producer

import (
	"time"

	"erp.localhost/infra/env"
)

type Config struct {
	// Kafka
	Brokers []string

	// Channel
	ChannelBuffer int // Default: 1000

	// Workers
	InitialWorkers int // Default: 3
	MaxWorkers     int // Default: 10

	// Shutdown
	ShutdownTimeout time.Duration // Default: 30s

	// Auto-scaling (optional)
	EnableAutoScale    bool
	ScaleUpThreshold   float64 // Channel utilization % (e.g., 0.8 = 80%)
	ScaleCheckInterval time.Duration
}

func DefaultConfig() Config {
	brokers := env.GetEnvAsSlice("KAFKA_BROKERS", ",", []string{"localhost:9092"})
	channelBuffer := env.GetEnvAsInt("PRODUCER_CHANNEL_BUFFER", 1000)
	initialWorkers := env.GetEnvAsInt("PRODUCER_INITIAL_WORKERS", 3)
	maxWorkers := env.GetEnvAsInt("PRODUCER_MAX_WORKERS", 10)
	shutdownTimeout := env.GetEnvAsDuration("PRODUCER_SHUTDOWN_TIMEOUT_IN_SECONDS", 30)
	enableAutoScale := env.GetEnvAsBool("PRODUCER_ENABLE_AUTO_SCALE", false)
	scaleUpThreshold := env.GetEnvAsFloat("PRODUCER_SCALE_UP_THRESHHOLD", 0.8)
	scaleCheckInterval := env.GetEnvAsDuration("PRODUCER_SCALE_CHECK_INTERVAL_IN_SECONDS", 10)
	return Config{
		Brokers:            brokers,
		ChannelBuffer:      channelBuffer,
		InitialWorkers:     initialWorkers,
		MaxWorkers:         maxWorkers,
		ShutdownTimeout:    shutdownTimeout * time.Second,
		EnableAutoScale:    enableAutoScale,
		ScaleUpThreshold:   scaleUpThreshold,
		ScaleCheckInterval: scaleCheckInterval * time.Second,
	}
}
