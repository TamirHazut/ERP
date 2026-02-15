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

	// WAL (Write-Ahead Log) configuration
	WalEnabled           bool          // Default: true
	WalDir               string        // Default: "./data/event_wal"
	WalMaxFileSize       int64         // Default: 100MB
	WalRetryInterval     time.Duration // Default: 30s
	WalBatchSize         int           // Default: 100
	WalDiskAlertThreshold int64        // Default: 10GB

	// Circuit Breaker configuration
	CircuitBreakerEnabled        bool          // Default: true
	CircuitBreakerFailureThreshold int32       // Default: 5
	CircuitBreakerSuccessThreshold int32       // Default: 2
	CircuitBreakerTimeout        time.Duration // Default: 60s
}

func DefaultConfig() Config {
	// Kafka settings
	brokers := env.GetEnvAsSlice("KAFKA_BROKERS", ",", []string{"localhost:9092"})
	channelBuffer := env.GetEnvAsInt("PRODUCER_CHANNEL_BUFFER", 1000)
	initialWorkers := env.GetEnvAsInt("PRODUCER_INITIAL_WORKERS", 3)
	maxWorkers := env.GetEnvAsInt("PRODUCER_MAX_WORKERS", 10)
	shutdownTimeout := env.GetEnvAsDuration("PRODUCER_SHUTDOWN_TIMEOUT_IN_SECONDS", 30)
	enableAutoScale := env.GetEnvAsBool("PRODUCER_ENABLE_AUTO_SCALE", false)
	scaleUpThreshold := env.GetEnvAsFloat("PRODUCER_SCALE_UP_THRESHHOLD", 0.8)
	scaleCheckInterval := env.GetEnvAsDuration("PRODUCER_SCALE_CHECK_INTERVAL_IN_SECONDS", 10)

	// WAL settings
	walEnabled := env.GetEnvAsBool("PRODUCER_WAL_ENABLED", true)
	walDir := env.GetEnv("PRODUCER_WAL_DIR", "./data/event_wal")
	walMaxFileSize := int64(env.GetEnvAsInt("PRODUCER_WAL_MAX_FILE_SIZE", 100*1024*1024)) // 100MB
	walRetryInterval := env.GetEnvAsDuration("PRODUCER_WAL_RETRY_INTERVAL_IN_SECONDS", 30)
	walBatchSize := env.GetEnvAsInt("PRODUCER_WAL_BATCH_SIZE", 100)
	walDiskAlertThreshold := int64(env.GetEnvAsInt("PRODUCER_WAL_DISK_ALERT_THRESHOLD", 10*1024*1024*1024)) // 10GB

	// Circuit Breaker settings
	circuitBreakerEnabled := env.GetEnvAsBool("PRODUCER_CIRCUIT_BREAKER_ENABLED", true)
	circuitBreakerFailureThreshold := env.GetEnvAsInt("PRODUCER_CIRCUIT_BREAKER_FAILURE_THRESHOLD", 5)
	circuitBreakerSuccessThreshold := env.GetEnvAsInt("PRODUCER_CIRCUIT_BREAKER_SUCCESS_THRESHOLD", 2)
	circuitBreakerTimeout := env.GetEnvAsDuration("PRODUCER_CIRCUIT_BREAKER_TIMEOUT_IN_SECONDS", 60)

	return Config{
		// Kafka
		Brokers:            brokers,
		ChannelBuffer:      channelBuffer,
		InitialWorkers:     initialWorkers,
		MaxWorkers:         maxWorkers,
		ShutdownTimeout:    shutdownTimeout * time.Second,
		EnableAutoScale:    enableAutoScale,
		ScaleUpThreshold:   scaleUpThreshold,
		ScaleCheckInterval: scaleCheckInterval * time.Second,

		// WAL
		WalEnabled:            walEnabled,
		WalDir:                walDir,
		WalMaxFileSize:        walMaxFileSize,
		WalRetryInterval:      walRetryInterval * time.Second,
		WalBatchSize:          walBatchSize,
		WalDiskAlertThreshold: walDiskAlertThreshold,

		// Circuit Breaker
		CircuitBreakerEnabled:          circuitBreakerEnabled,
		CircuitBreakerFailureThreshold: int32(circuitBreakerFailureThreshold),
		CircuitBreakerSuccessThreshold: int32(circuitBreakerSuccessThreshold),
		CircuitBreakerTimeout:          circuitBreakerTimeout * time.Second,
	}
}
