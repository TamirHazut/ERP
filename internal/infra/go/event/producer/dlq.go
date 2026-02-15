package producer

import (
	"time"

	collection_mongo "erp.localhost/infra/db/mongo/collection"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/event/collection"
	"erp.localhost/infra/logging/logger"
	"erp.localhost/infra/model/event"
	eventv1 "erp.localhost/infra/model/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DLQHandler interface {
	Store(message *eventv1.Message, topic event.Topic, partionKey string, err error) error
	Count() (int64, error) // Get the number of entries in DLQ
}

type BaseDLQHandler struct {
	collection collection_mongo.CollectionHandler[eventv1.DlqEntry]
	logger     logger.Logger
}

func NewBaseDLQHandler(logger logger.Logger) (*BaseDLQHandler, *infra_error.AppError) {
	collection, err := collection.NewDlqCollection(logger)
	if err != nil {
		logger.Error("failed to create user collection handler", "error", err)
		return nil, err
	}
	return &BaseDLQHandler{
		collection: collection,
		logger:     logger,
	}, nil
}

func (h *BaseDLQHandler) Store(message *eventv1.Message, topic event.Topic, partionKey string, err error) error {
	entry := &eventv1.DlqEntry{
		Topic:        string(topic),
		PartitionKey: partionKey,
		Message:      message,
		Retries:      0,
		MaxRetries:   10,
		NextRetryAt:  timestamppb.New(time.Now().Add(30 * time.Second)),
		CreatedAt:    timestamppb.Now(),
		UpdatedAt:    timestamppb.Now(),
		LastError:    err.Error(),
	}
	_, storeErr := h.collection.Create(entry)
	return storeErr
}

// Count returns the total number of entries in the DLQ.
// TODO: Implement actual count when CollectionHandler supports Count() method
func (h *BaseDLQHandler) Count() (int64, error) {
	// For now, return 0 as the CollectionHandler doesn't have a Count method yet
	// This will need to be implemented when we add Count support to CollectionHandler
	return 0, nil
}
