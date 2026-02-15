package producer

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"erp.localhost/infra/event/producer/wal"
	"erp.localhost/infra/logging/logger"
	"erp.localhost/infra/model/event"
	eventv1 "erp.localhost/infra/model/event/v1"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WalWorker processes WAL entries in the background and retries sending them to Kafka or DLQ.
type WalWorker struct {
	walQueue          *WalQueueHandler
	producer          *Producer // Access to Kafka writer and DLQ
	pollInterval      time.Duration
	batchSize         int
	diskAlertThreshold int64
	stopChan          chan struct{}
	wg                sync.WaitGroup
	logger            logger.Logger

	// File state tracking for deletion
	fileStates   map[int64]*fileState // fileSeq -> state
	fileStatesMu sync.RWMutex
}

// fileState tracks the state of entries in a WAL file.
type fileState struct {
	totalEntries int                              // Total entries in file
	sentEntries  int                              // Number of entries successfully sent
	entryStates  map[string]eventv1.DlqEntryState // messageID -> state
	mu           sync.RWMutex
}

// NewWalWorker creates a new WAL background worker.
func NewWalWorker(
	walQueue *WalQueueHandler,
	producer *Producer,
	pollInterval time.Duration,
	batchSize int,
	diskAlertThreshold int64,
	logger logger.Logger,
) *WalWorker {
	return &WalWorker{
		walQueue:           walQueue,
		producer:           producer,
		pollInterval:       pollInterval,
		batchSize:          batchSize,
		diskAlertThreshold: diskAlertThreshold,
		stopChan:           make(chan struct{}),
		logger:             logger,
		fileStates:         make(map[int64]*fileState),
	}
}

// Start begins the background processing loop.
func (w *WalWorker) Start() {
	w.wg.Add(2)

	// Start processing loop
	go w.processLoop()

	// Start disk monitoring loop
	go w.monitorDiskUsage()

	w.logger.Info("WAL worker started",
		"poll_interval", w.pollInterval,
		"batch_size", w.batchSize,
		"disk_alert_threshold_gb", w.diskAlertThreshold/(1024*1024*1024))
}

// Stop gracefully stops the worker.
func (w *WalWorker) Stop() {
	w.logger.Info("Stopping WAL worker...")
	close(w.stopChan)
	w.wg.Wait()
	w.logger.Info("WAL worker stopped")
}

// processLoop polls the WAL and processes entries.
func (w *WalWorker) processLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.processBatch(); err != nil {
				w.logger.Error("Error processing WAL batch", "error", err)
			}

		case <-w.stopChan:
			w.logger.Info("Process loop stopped")
			return
		}
	}
}

// processBatch reads and processes a batch of WAL entries.
func (w *WalWorker) processBatch() error {
	// Load checkpoint to know where to start reading
	fileSeq, offset, err := w.walQueue.checkpoint.Load()
	if err != nil {
		return fmt.Errorf("failed to load checkpoint: %w", err)
	}

	// Open iterator at checkpoint position
	iter, err := w.walQueue.wal.ReadFrom(fileSeq, offset)
	if err != nil {
		// File doesn't exist yet (no WAL entries written)
		if err.Error() == "failed to open WAL file for reading: open" {
			return nil
		}
		return fmt.Errorf("failed to create WAL iterator: %w", err)
	}
	defer iter.Close()

	// Read batch of entries
	entries, newOffset, err := iter.NextN(w.batchSize)
	if err != nil && err != wal.ErrEOF {
		return fmt.Errorf("failed to read WAL entries: %w", err)
	}

	if len(entries) == 0 {
		// No entries to process
		return nil
	}

	w.logger.Debug("Processing WAL batch",
		"file_seq", fileSeq,
		"offset", offset,
		"entries", len(entries))

	// Initialize file state if not already tracked
	w.initializeFileState(fileSeq, entries)

	// Track how many entries were successfully sent (to update checkpoint)
	processedCount := 0

	// Process each entry
	for _, entry := range entries {
		// Check if entry is ready for retry (respect exponential backoff)
		if entry.NextRetryAt != nil && time.Now().Before(entry.NextRetryAt.AsTime()) {
			w.logger.Debug("Entry not ready for retry yet",
				"message_id", entry.MessageId,
				"next_retry_at", entry.NextRetryAt.AsTime())
			continue
		}

		// Try to send the entry
		success, err := w.processEntry(entry)
		if err != nil {
			w.logger.Error("Error processing WAL entry",
				"message_id", entry.MessageId,
				"error", err)
			// Continue processing other entries even if one fails
			continue
		}

		if success {
			processedCount++
			// Mark entry as SENT in file state
			w.markEntrySent(fileSeq, entry.MessageId)
		}
	}

	// Update checkpoint only if we processed entries successfully
	if processedCount > 0 {
		// Save checkpoint at the new offset
		if err := w.walQueue.checkpoint.Save(fileSeq, newOffset); err != nil {
			w.logger.Error("Failed to save checkpoint",
				"file_seq", fileSeq,
				"offset", newOffset,
				"error", err)
			// Don't return error - we'll retry on next poll
		} else {
			w.logger.Debug("Checkpoint updated",
				"file_seq", fileSeq,
				"new_offset", newOffset,
				"processed", processedCount)
		}

		// Check if we should delete the file (all entries processed)
		if err == wal.ErrEOF && w.canDeleteFile(fileSeq) {
			w.logger.Info("All entries in WAL file processed, deleting file",
				"file_seq", fileSeq)

			if deleteErr := w.walQueue.wal.DeleteFile(fileSeq); deleteErr != nil {
				w.logger.Error("Failed to delete WAL file",
					"file_seq", fileSeq,
					"error", deleteErr)
			} else {
				// Remove file state tracking
				w.removeFileState(fileSeq)
				w.logger.Info("WAL file deleted successfully", "file_seq", fileSeq)
			}
		}
	}

	return nil
}

// initializeFileState initializes tracking for a WAL file if not already tracked.
func (w *WalWorker) initializeFileState(fileSeq int64, entries []*eventv1.DlqEntry) {
	w.fileStatesMu.Lock()
	defer w.fileStatesMu.Unlock()

	// Check if already initialized
	if _, exists := w.fileStates[fileSeq]; exists {
		return
	}

	// Create new file state
	state := &fileState{
		totalEntries: 0, // Will be incremented as we see entries
		sentEntries:  0,
		entryStates:  make(map[string]eventv1.DlqEntryState),
	}

	// Add all entries in this batch to the state
	for _, entry := range entries {
		state.entryStates[entry.MessageId] = entry.State
		state.totalEntries++

		// If entry is already marked as SENT, count it
		if entry.State == eventv1.DlqEntryState_DLQ_ENTRY_STATE_SENT {
			state.sentEntries++
		}
	}

	w.fileStates[fileSeq] = state

	w.logger.Debug("Initialized file state tracking",
		"file_seq", fileSeq,
		"total_entries", state.totalEntries,
		"sent_entries", state.sentEntries)
}

// markEntrySent marks an entry as successfully sent.
func (w *WalWorker) markEntrySent(fileSeq int64, messageID string) {
	w.fileStatesMu.RLock()
	state, exists := w.fileStates[fileSeq]
	w.fileStatesMu.RUnlock()

	if !exists {
		w.logger.Warn("Cannot mark entry as sent - file state not tracked",
			"file_seq", fileSeq,
			"message_id", messageID)
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	// Check if entry was already sent
	if state.entryStates[messageID] == eventv1.DlqEntryState_DLQ_ENTRY_STATE_SENT {
		return // Already counted
	}

	// Mark as sent and increment counter
	state.entryStates[messageID] = eventv1.DlqEntryState_DLQ_ENTRY_STATE_SENT
	state.sentEntries++

	w.logger.Debug("Entry marked as sent",
		"file_seq", fileSeq,
		"message_id", messageID,
		"sent_entries", state.sentEntries,
		"total_entries", state.totalEntries)
}

// canDeleteFile checks if all entries in a file have been sent.
func (w *WalWorker) canDeleteFile(fileSeq int64) bool {
	w.fileStatesMu.RLock()
	state, exists := w.fileStates[fileSeq]
	w.fileStatesMu.RUnlock()

	if !exists {
		// No state tracking - cannot safely delete
		return false
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	// Can delete if all entries are sent
	canDelete := state.sentEntries == state.totalEntries && state.totalEntries > 0

	if canDelete {
		w.logger.Debug("File is safe to delete",
			"file_seq", fileSeq,
			"total_entries", state.totalEntries,
			"sent_entries", state.sentEntries)
	}

	return canDelete
}

// removeFileState removes state tracking for a deleted file.
func (w *WalWorker) removeFileState(fileSeq int64) {
	w.fileStatesMu.Lock()
	defer w.fileStatesMu.Unlock()

	delete(w.fileStates, fileSeq)
}

// processEntry attempts to send a single DLQ entry to Kafka or DLQ.
// Returns (true, nil) if successfully sent, (false, nil) if failed but will retry later, (false, err) on error.
func (w *WalWorker) processEntry(entry *eventv1.DlqEntry) (bool, error) {
	// Get the message directly (no need to unmarshal, it's already in the DlqEntry)
	message := entry.Message
	if message == nil {
		return false, fmt.Errorf("DLQ entry has nil message")
	}

	// Try sending to Kafka first (prefer primary path)
	if err := w.sendToKafka(message, event.Topic(entry.Topic), entry.PartitionKey); err == nil {
		// Success! Entry can be marked as sent
		w.logger.Info("WAL entry sent to Kafka",
			"message_id", entry.MessageId,
			"topic", entry.Topic,
			"retries", entry.Retries)
		return true, nil
	} else {
		w.logger.Debug("Kafka send failed for WAL entry, trying DLQ",
			"message_id", entry.MessageId,
			"error", err)
	}

	// Kafka failed, try MongoDB DLQ
	if err := w.producer.dlq.Store(message, event.Topic(entry.Topic), entry.PartitionKey, fmt.Errorf("WAL retry: %s", entry.LastError)); err == nil {
		// DLQ succeeded, entry can be marked as sent
		w.logger.Info("WAL entry sent to DLQ",
			"message_id", entry.MessageId,
			"topic", entry.Topic,
			"retries", entry.Retries)
		return true, nil
	} else {
		w.logger.Debug("DLQ send failed for WAL entry",
			"message_id", entry.MessageId,
			"error", err)
	}

	// Both Kafka and DLQ failed - update retry metadata
	entry.Retries++
	entry.NextRetryAt = timestamppb.New(calculateNextRetry(entry.Retries))
	entry.LastError = fmt.Sprintf("Retry %d failed: Kafka and DLQ unavailable", entry.Retries)

	w.logger.Warn("WAL entry will be retried later",
		"message_id", entry.MessageId,
		"retries", entry.Retries,
		"next_retry_at", entry.NextRetryAt.AsTime())

	// Entry stays in WAL for next retry
	return false, nil
}

// sendToKafka sends a message directly to Kafka (bypassing the producer's channel).
func (w *WalWorker) sendToKafka(msg *eventv1.Message, topic event.Topic, partitionKey string) error {
	// Serialize protobuf
	bytes, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Create Kafka message
	kafkaMsg := kafka.Message{
		Topic: string(topic),
		Key:   []byte(partitionKey),
		Value: bytes,
		Headers: []kafka.Header{
			{Key: "tenant_id", Value: []byte(msg.TenantId)},
			{Key: "event_type", Value: []byte(msg.EventType.String())},
			{Key: "message_id", Value: []byte(msg.Id)}, // Preserve original UUID
		},
		Time: msg.Timestamp.AsTime(),
	}

	// Send to Kafka with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return w.producer.writer.WriteMessages(ctx, kafkaMsg)
}

// monitorDiskUsage periodically checks WAL disk usage and alerts if threshold exceeded.
func (w *WalWorker) monitorDiskUsage() {
	defer w.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			totalSize, err := w.walQueue.GetTotalSize()
			if err != nil {
				w.logger.Error("Failed to get WAL total size", "error", err)
				continue
			}

			totalSizeGB := float64(totalSize) / (1024 * 1024 * 1024)

			if totalSize > w.diskAlertThreshold {
				w.logger.Error("WAL DISK USAGE CRITICAL",
					"total_size_gb", fmt.Sprintf("%.2f", totalSizeGB),
					"threshold_gb", w.diskAlertThreshold/(1024*1024*1024),
					"action", "Ops team intervention required - investigate Kafka/MongoDB availability")
			} else if totalSize > (w.diskAlertThreshold / 2) {
				w.logger.Warn("WAL disk usage high",
					"total_size_gb", fmt.Sprintf("%.2f", totalSizeGB),
					"threshold_gb", w.diskAlertThreshold/(1024*1024*1024))
			} else {
				w.logger.Debug("WAL disk usage normal",
					"total_size_gb", fmt.Sprintf("%.2f", totalSizeGB))
			}

		case <-w.stopChan:
			w.logger.Info("Disk monitor stopped")
			return
		}
	}
}

// calculateNextRetry calculates the next retry time using exponential backoff.
// Backoff: 30s, 60s, 120s, 240s, ..., capped at 1 hour.
func calculateNextRetry(retries int32) time.Time {
	baseDelay := 30.0 // 30 seconds
	maxDelay := 3600.0 // 1 hour

	backoff := baseDelay * math.Pow(2, float64(retries))
	if backoff > maxDelay {
		backoff = maxDelay
	}

	return time.Now().Add(time.Duration(backoff) * time.Second)
}
