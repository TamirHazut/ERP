package producer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	"erp.localhost/internal/infra/model/event"
	eventv1 "erp.localhost/internal/infra/model/event/v1"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// Global singleton instance
	defaultProducer *Producer

	// Ensures Init is called only once
	once sync.Once

	// Tracks initialization state
	initialized    bool
	mu             sync.RWMutex
	ErrChannelFull = fmt.Errorf("events channel is full")
)

type messageWrapper struct {
	topic   event.Topic
	message *eventv1.Message
}

type Metrics struct {
	Pending int32
	Workers int32
	Sent    int64
	Failed  int64
}

type Producer struct {
	logger logger.Logger
	// Kafka writer
	writer *kafka.Writer

	// Channel for async sending
	ch chan *messageWrapper

	// DLQ handler
	dlq DLQHandler

	// Worker management
	workerCount atomic.Int32
	wg          sync.WaitGroup
	done        chan struct{}

	// Metrics
	pending atomic.Int32
	sent    atomic.Int64
	failed  atomic.Int64

	// Config
	config Config
}

func newProducer(cfg Config, dlq DLQHandler, logger logger.Logger) (*Producer, *infra_error.AppError) {
	// Create kafka writer (producer)
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Hash{}, // Hash by key for partition consistency
		MaxAttempts:  3,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		RequiredAcks: kafka.RequireAll, // Wait for all replicas (acks=all)
		Async:        false,            // Synchronous for error handling
		Compression:  kafka.Snappy,

		// Idempotence (requires Kafka 0.11+)
		// Note: kafka-go doesn't have built-in idempotence like Sarama
		// You need to ensure idempotence at application level or use transactions
	}

	p := &Producer{
		logger: logger,
		writer: writer,
		ch:     make(chan *messageWrapper, cfg.ChannelBuffer),
		dlq:    dlq,
		done:   make(chan struct{}),
		config: cfg,
	}

	return p, nil
}

func (p *Producer) start() *infra_error.AppError {
	for i := 0; i < p.config.InitialWorkers; i++ {
		p.addWorker()
	}

	p.logger.Info("Producer started",
		"workers", p.config.InitialWorkers,
		"buffer", p.config.ChannelBuffer,
	)

	return nil
}

func (p *Producer) send(topic event.Topic, message *eventv1.Message) *infra_error.AppError {
	msg := &messageWrapper{
		topic:   topic,
		message: message,
	}

	select {
	case p.ch <- msg:
		p.pending.Add(1)
		return nil
	default:
		// Channel full
		// metrics.ChannelFullErrors.Inc()
		return infra_error.Internal(infra_error.InternalCacheError, ErrChannelFull)
	}
}

func (p *Producer) getMetrics() Metrics {
	return Metrics{
		Pending: p.pending.Load(),
		Workers: p.workerCount.Load(),
		Sent:    p.sent.Load(),
		Failed:  p.failed.Load(),
	}
}

func (p *Producer) shutdown() *infra_error.AppError {
	p.logger.Info("Shutting down producer...")

	// Stop accepting new messages
	close(p.ch)

	// Signal workers to stop
	close(p.done)

	// Wait with timeout from config
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.Info("All workers finished")
	case <-time.After(p.config.ShutdownTimeout):
		pending := p.pending.Load()
		p.logger.Warn("Shutdown timeout", "pending", pending)
		return infra_error.Internal(infra_error.InternalUnexpectedError, fmt.Errorf("shutdown timeout: %d messages pending", pending))
	}

	// Close Kafka writer
	if err := p.writer.Close(); err != nil {
		closeErr := infra_error.Internal(infra_error.InternalKafkaError, err)
		p.logger.Error("Error closing Kafka writer", "error", closeErr)
		return closeErr
	}

	p.logger.Info("Producer shutdown complete",
		"sent", p.sent.Load(),
		"failed", p.failed.Load(),
	)

	return nil
}

func (p *Producer) addWorker() {
	p.wg.Add(1)
	p.workerCount.Add(1)

	go func() {
		defer p.wg.Done()
		defer p.workerCount.Add(-1)

		for {
			select {
			case msg, ok := <-p.ch:
				if !ok {
					return // Channel closed
				}

				p.pending.Add(-1)
				p.processMessage(msg)

			case <-p.done:
				return
			}
		}
	}()
}

// TODO: remove comments when metrics is available
func (p *Producer) processMessage(msgWrapper *messageWrapper) {
	topic := msgWrapper.topic
	msg := msgWrapper.message
	// Serialize protobuf
	bytes, err := proto.Marshal(msg)
	if err != nil {
		p.logger.Error("Failed to serialize message",
			"message_id", msg.Id,
			"error", err,
		)
		// metrics.SerializationErrors.Inc()
		return
	}

	// Create Kafka message
	kafkaMsg := kafka.Message{
		Topic: string(msgWrapper.topic),
		Key:   []byte(p.getPartitionKey(msg)),
		Value: bytes,
		Headers: []kafka.Header{
			{
				Key:   "tenant_id",
				Value: []byte(msg.TenantId),
			},
			{
				Key:   "event_type",
				Value: []byte(msg.EventType.String()),
			},
			{
				Key:   "message_id",
				Value: []byte(msg.Id),
			},
		},
		Time: msg.Timestamp.AsTime(),
	}

	// Send to Kafka
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = p.writer.WriteMessages(ctx, kafkaMsg)
	if err != nil {
		// Failed - write to DLQ
		p.logger.Error("Failed to send to Kafka, writing to DLQ",
			"message_id", msg.Id,
			"topic", topic,
			"error", err,
		)
		if dlqErr := p.dlq.Store(msg, topic, string(kafkaMsg.Key), err); dlqErr != nil {
			p.logger.Error("Failed to write to DLQ",
				"message_id", msg.Id,
				"error", dlqErr,
			)
		}

		p.failed.Add(1)
		// metrics.MessagesFailed.Inc()
	} else {
		// Success
		p.sent.Add(1)
		// metrics.MessagesSent.Inc()
	}
}

// Partition key
func (p *Producer) getPartitionKey(event *eventv1.Message) string {
	tenantID := event.GetTenantId()
	entityType := event.GetEntityType()
	entityID := event.GetEntityId()
	if entityType != "" && entityID != "" {
		// partition key will be <TenantID>_<EntityType>_<EntityID>
		return fmt.Sprintf("%s_%s_%s", tenantID, entityType, entityID)
	}
	return tenantID
}

// ========================
// Global singleton functions
// ========================

// Init initializes the global producer singleton
// Must be called once at application startup
func Init(cfg Config, dlq DLQHandler, logger logger.Logger) *infra_error.AppError {
	var initErr *infra_error.AppError

	once.Do(func() {
		defaultProducer, initErr = newProducer(cfg, dlq, logger)
		if initErr != nil {
			return
		}

		initErr = defaultProducer.start()
		if initErr != nil {
			return
		}

		mu.Lock()
		initialized = true
		mu.Unlock()
	})

	return initErr
}

func Send(topic event.Topic, tenantID string, eventType eventv1.EventType, data any) *infra_error.AppError {
	mu.RLock()
	if !initialized {
		mu.RUnlock()
		return infra_error.Internal(infra_error.InternalServiceUnavailable, errors.New("producer not initialized: call Init() first"))
	}
	mu.RUnlock()
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "tenant_id")
	}
	if data == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "event_data")
	}

	message, ok := data.(proto.Message)
	if !ok {
		return infra_error.Validation(infra_error.ValidationInvalidType, "data")
	}
	eventData, _ := anypb.New(message)
	wrapper := &eventv1.Message{
		Id:        uuid.New().String(),
		TenantId:  tenantID,
		Timestamp: timestamppb.Now(),
		EventType: eventv1.EventType_EVENT_TYPE_LOGIN_FAILED,
		EventData: eventData,
	}
	return defaultProducer.send(topic, wrapper)
}

// Shutdown gracefully shuts down the producer
func Shutdown() *infra_error.AppError {
	mu.RLock()
	if !initialized {
		mu.RUnlock()
		return infra_error.Internal(infra_error.InternalServiceUnavailable, errors.New("producer not initialized: call Init() first"))
	}
	mu.RUnlock()

	return defaultProducer.shutdown()
}

// GetMetrics returns current producer metrics
func GetMetrics() Metrics {
	mu.RLock()
	defer mu.RUnlock()

	if !initialized {
		return Metrics{}
	}

	return defaultProducer.getMetrics()
}
