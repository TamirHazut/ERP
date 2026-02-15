package producer

import (
	"time"

	"erp.localhost/infra/event/producer/wal"
	"erp.localhost/infra/logging/logger"
	"erp.localhost/infra/model/event"
	eventv1 "erp.localhost/infra/model/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WalQueueHandler implements the queue interface using Write-Ahead Log files.
// This provides a local persistent queue when both Kafka and MongoDB are unavailable.
type WalQueueHandler struct {
	wal        *wal.WAL
	checkpoint *wal.Checkpoint
	logger     logger.Logger
}

// NewWalQueueHandler creates a new WAL-based queue handler.
func NewWalQueueHandler(dir string, maxFileSize int64, logger logger.Logger) (*WalQueueHandler, error) {
	// Initialize WAL
	walInstance, err := wal.NewWAL(dir, maxFileSize)
	if err != nil {
		return nil, err
	}

	// Initialize checkpoint
	checkpoint, err := wal.NewCheckpoint(dir)
	if err != nil {
		walInstance.Close()
		return nil, err
	}

	return &WalQueueHandler{
		wal:        walInstance,
		checkpoint: checkpoint,
		logger:     logger,
	}, nil
}

// Store writes an event to the WAL.
// This method preserves the original message_id (UUID) for consumer deduplication.
func (h *WalQueueHandler) Store(message *eventv1.Message, topic event.Topic, partitionKey string, err error) error {
	// Create DLQ entry for WAL storage (unified proto model)
	entry := &eventv1.DlqEntry{
		MessageId:    message.Id, // Preserve original UUID for deduplication
		Topic:        string(topic),
		PartitionKey: partitionKey,
		Message:      message, // Store full message (not just bytes)
		CreatedAt:    timestamppb.Now(),
		UpdatedAt:    timestamppb.Now(),
		Retries:      0,
		MaxRetries:   10, // Default max retries
		NextRetryAt:  timestamppb.New(time.Now().Add(30 * time.Second)),
		LastError:    err.Error(),
		State:        eventv1.DlqEntryState_DLQ_ENTRY_STATE_PENDING,
	}

	// Append to WAL
	if appendErr := h.wal.Append(entry); appendErr != nil {
		h.logger.Error("Failed to append entry to WAL",
			"message_id", message.Id,
			"error", appendErr)
		return appendErr
	}

	h.logger.Debug("Event written to WAL",
		"message_id", message.Id,
		"topic", topic,
		"partition_key", partitionKey)

	return nil
}

// Count returns the total number of WAL files (excluding checkpoint).
// Used for metrics monitoring.
func (h *WalQueueHandler) Count() (int, error) {
	return h.wal.GetFileCount()
}

// GetTotalSize returns the total size of all WAL files in bytes.
// Used for disk usage monitoring.
func (h *WalQueueHandler) GetTotalSize() (int64, error) {
	return h.wal.GetTotalSize()
}

// Close flushes and closes the WAL.
func (h *WalQueueHandler) Close() error {
	return h.wal.Close()
}
