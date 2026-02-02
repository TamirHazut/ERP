package aggregation

import (
	"context"

	"erp.localhost/internal/infra/db/mongo"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AggregationHandler generic interface for MongoDB aggregation operations
// Follows same pattern as CollectionHandler[T] for consistency
type AggregationHandler[T any] interface {
	// Aggregate executes an aggregation pipeline and returns results of type T
	Aggregate(ctx context.Context, pipeline []bson.M, fields []string) ([]*T, *infra_error.AppError)

	// BatchGetByIDs retrieves multiple documents by IDs using $in operator
	BatchGetByIDs(ctx context.Context, tenantID string, ids []string, fields []string) ([]*T, *infra_error.AppError)
}

// BaseAggregationHandler provides generic aggregation functionality
// Follows same pattern as BaseCollectionHandler[T]
type BaseAggregationHandler[T any] struct {
	dbHandler  *mongo.MongoDBManager
	collection string
	logger     logger.Logger
}

// NewBaseAggregationHandler creates a new generic aggregation handler
func NewBaseAggregationHandler[T any](dbName model_mongo.DBName, collection model_mongo.Collection, logger logger.Logger) (*BaseAggregationHandler[T], *infra_error.AppError) {
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
) ([]*T, *infra_error.AppError) {
	return h.AggregateFrom(ctx, h.collection, pipeline, fields)
}

// AggregateFrom executes an aggregation pipeline on a specific collection with optional field projection
// Use this when the pipeline needs to start from a different collection than the handler's default
func (h *BaseAggregationHandler[T]) AggregateFrom(
	ctx context.Context,
	collectionName string,
	pipeline []bson.M,
	fields []string,
) ([]*T, *infra_error.AppError) {
	// Apply field projection if specified
	if len(fields) > 0 {
		projection := bson.M{}
		for _, field := range fields {
			projection[field] = 1
		}
		pipeline = append(pipeline, bson.M{"$project": projection})
	}

	h.logger.Debug("executing aggregation pipeline", "collection", collectionName, "stages", len(pipeline))

	// Execute aggregation using dbHandler's Aggregate method
	cursor, err := h.dbHandler.Aggregate(ctx, collectionName, pipeline)
	if err != nil {
		h.logger.Error("aggregation failed", "error", err, "collection", collectionName)
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode results
	results := make([]*T, 0)
	if err := cursor.All(ctx, &results); err != nil {
		h.logger.Error("failed to decode aggregation results", "error", err, "collection", collectionName)
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	h.logger.Debug("aggregation completed", "collection", collectionName, "results_count", len(results))
	return results, nil
}

// BatchGetByIDs retrieves multiple documents by IDs using $in operator
func (h *BaseAggregationHandler[T]) BatchGetByIDs(
	ctx context.Context,
	tenantID string,
	ids []string,
	fields []string,
) ([]*T, *infra_error.AppError) {
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
