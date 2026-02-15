package producer

// EnhancedMetrics contains comprehensive metrics about the event producer.
type EnhancedMetrics struct {
	// Existing metrics
	Pending int32 // Messages in channel waiting to be sent
	Workers int32 // Number of active worker goroutines
	Sent    int64 // Total successfully sent messages
	Failed  int64 // Total failed messages

	// DLQ metrics
	DLQDepth int64 // Number of entries in MongoDB DLQ

	// WAL metrics
	WalFileCount  int   // Number of WAL files
	WalTotalSize  int64 // Total size of WAL files in bytes
	WalTotalSizeGB float64 // Total size in GB (for convenience)

	// Circuit breaker metrics
	CircuitState    string // "CLOSED", "OPEN", "HALF_OPEN"
	CircuitFailures int32  // Current failure count
}
