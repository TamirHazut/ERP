package aggregation

import (
	"context"

	"erp.localhost/internal/infra/db/mongo"
	"erp.localhost/internal/infra/logging/logger"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AggregationHandler generic interface for MongoDB aggregation operations
// Follows same pattern as CollectionHandler[T] for consistency
type AggregationHandler[T any] interface {
	// Aggregate executes an aggregation pipeline and returns results of type T
	Aggregate(ctx context.Context, pipeline []bson.M, fields []string) ([]*T, error)

	// BatchGetByIDs retrieves multiple documents by IDs using $in operator
	BatchGetByIDs(ctx context.Context, tenantID string, ids []string, fields []string) ([]*T, error)
}

// BaseAggregationHandler provides generic aggregation functionality
// Follows same pattern as BaseCollectionHandler[T]
type BaseAggregationHandler[T any] struct {
	dbHandler  *mongo.MongoDBManager
	collection string
	logger     logger.Logger
}

// NewBaseAggregationHandler creates a new generic aggregation handler
func NewBaseAggregationHandler[T any](dbName model_mongo.DBName, collection model_mongo.Collection, logger logger.Logger) (*BaseAggregationHandler[T], error) {
	dbHandler, err := mongo.NewMongoDBManager(dbName, logger)
	if dbHandler == nil {
		logger.Fatal("failed to create mongo db manager for aggregation handler", "error", err)
		return nil, err
	}

	return &BaseAggregationHandler[T]{
		dbHandler:  dbHandler,
		collection: string(collection),
		logger:     logger,
	}, nil
}

// Aggregate executes an aggregation pipeline with optional field projection
func (h *BaseAggregationHandler[T]) Aggregate(
	ctx context.Context,
	pipeline []bson.M,
	fields []string,
) ([]*T, error) {
	// Apply field projection if specified
	if len(fields) > 0 {
		projection := bson.M{}
		for _, field := range fields {
			projection[field] = 1
		}
		pipeline = append(pipeline, bson.M{"$project": projection})
	}

	h.logger.Debug("executing aggregation pipeline", "collection", h.collection, "stages", len(pipeline))

	// Execute aggregation using dbHandler's Aggregate method
	cursor, err := h.dbHandler.Aggregate(ctx, h.collection, pipeline)
	if err != nil {
		h.logger.Error("aggregation failed", "error", err, "collection", h.collection)
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode results
	results := make([]*T, 0)
	if err := cursor.All(ctx, &results); err != nil {
		h.logger.Error("failed to decode aggregation results", "error", err, "collection", h.collection)
		return nil, err
	}

	h.logger.Debug("aggregation completed", "collection", h.collection, "results_count", len(results))
	return results, nil
}

// BatchGetByIDs retrieves multiple documents by IDs using $in operator
func (h *BaseAggregationHandler[T]) BatchGetByIDs(
	ctx context.Context,
	tenantID string,
	ids []string,
	fields []string,
) ([]*T, error) {
	// Convert string IDs to ObjectIDs
	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			h.logger.Warn("invalid object id", "id", id, "error", err)
			continue
		}
		objectIDs = append(objectIDs, objectID)
	}

	if len(objectIDs) == 0 {
		h.logger.Debug("no valid IDs to fetch", "collection", h.collection)
		return []*T{}, nil
	}

	// Build pipeline with $match using $in operator
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"_id":       bson.M{"$in": objectIDs},
			},
		},
	}

	return h.Aggregate(ctx, pipeline, fields)
}
