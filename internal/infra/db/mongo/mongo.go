package mongo

import (
	"context"
	"errors"
	"time"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBManager struct {
	client *mongo.Client
	dbName model_mongo.DBName
	db     *mongo.Database
	logger logger.Logger
}

func NewMongoDBManager(dbName model_mongo.DBName, logger logger.Logger) (*MongoDBManager, error) {
	if logger == nil {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "logger")
	}
	m := &MongoDBManager{
		dbName: dbName,
		logger: logger,
	}
	if err := m.Init(); err != nil {
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return m, nil
}

func (m *MongoDBManager) Close() error {
	if err := m.client.Disconnect(context.Background()); err != nil {
		m.logger.Error("failed to disconnect from mongo", "error", err)
		return err
	}
	return nil
}

func (m *MongoDBManager) Init() error {
	uri := "mongodb://root:secret@localhost:27017"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		m.logger.Fatal("failed to connect to mongo", "error", err)
		return err
	}
	if err := client.Ping(ctx, nil); err != nil {
		m.logger.Fatal("failed to ping mongo", "error", err)
		return err
	}
	m.client = client
	if err := m.createDBIfNotExists(); err != nil {
		return err
	}
	return nil
}

func (m *MongoDBManager) CreateCollectionInDBIfNotExists(collectionName string) error {
	m.logger.Debug("checking if collection esists", "db", m.dbName, "collection", collectionName)
	filter := bson.M{"name": collectionName}
	names, err := m.db.ListCollectionNames(context.Background(), filter)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	if len(names) > 0 {
		m.logger.Debug("collection already exists", "db", m.dbName, "collection", collectionName)
		return nil
	}
	m.logger.Info("creating collection", "db", m.dbName, "collection", collectionName)
	if err := m.db.CreateCollection(context.Background(), collectionName); err != nil {
		m.logger.Error("failed to create collection", "db", m.dbName, "collection", collectionName, "error", err)
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}
	return nil
}

func (m *MongoDBManager) createDBIfNotExists() error {
	m.logger.Debug("checking if db esists", "dbName", m.dbName)
	m.db = m.client.Database(string(m.dbName))
	if m.db == nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, errors.New("database not found"))
	}
	return nil
}

func (m *MongoDBManager) Create(collectionName string, data any, opts ...map[string]any) (string, error) {
	m.logger.Debug("creating data", "collection", collectionName, "data", data)
	collection := m.db.Collection(collectionName)
	result, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (m *MongoDBManager) FindOne(collectionName string, filter map[string]any, result any) error {
	m.logger.Debug("finding one", "collection", collectionName, "filter", filter)
	if filter == nil {
		return errors.New("filter is required and cannot be nil")
	}
	collection := m.db.Collection(collectionName)
	m.convertFilterToMongoTypes(filter)
	item := collection.FindOne(context.Background(), filter)
	if err := item.Err(); err != nil {
		return err
	}
	if item == nil {
		return errors.New("no result found")
	}
	if err := item.Decode(result); err != nil {
		return err
	}
	return nil
}

func (m *MongoDBManager) FindAll(collectionName string, filter map[string]any, result any) error {
	m.logger.Debug("finding all", "collection", collectionName, "filter", filter)
	if filter == nil {
		return errors.New("filter is required and cannot be nil")
	}
	collection := m.db.Collection(collectionName)
	m.convertFilterToMongoTypes(filter)
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return err
	}
	if err := cursor.All(context.Background(), result); err != nil {
		return err
	}
	return nil
}

func (m *MongoDBManager) Update(collectionName string, filter map[string]any, data any, opts ...map[string]any) error {
	m.logger.Debug("updating data", "collection", collectionName, "filter", filter, "data", data)
	if filter == nil {
		return errors.New("filter is required and cannot be nil")
	}
	collection := m.db.Collection(collectionName)
	m.convertFilterToMongoTypes(filter)
	_, err := collection.UpdateOne(context.Background(), filter, bson.M{"$set": data})
	if err != nil {
		return err
	}
	return nil
}

func (m *MongoDBManager) Delete(collectionName string, filter map[string]any) error {
	m.logger.Debug("deleting data", "collection", collectionName, "filter", filter)
	if filter == nil {
		return errors.New("filter is required and cannot be nil")
	}
	collection := m.db.Collection(collectionName)
	m.convertFilterToMongoTypes(filter)
	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}

// EnsureIndexes creates indexes for a collection if they don't exist (idempotent)
func (m *MongoDBManager) EnsureIndexes(collectionName string, indexes []mongo.IndexModel) error {
	m.logger.Debug("ensuring indexes", "collection", collectionName, "count", len(indexes))
	collection := m.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	names, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		m.logger.Error("failed to create indexes", "collection", collectionName, "error", err)
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	m.logger.Info("indexes ensured", "collection", collectionName, "indexes", names)
	return nil
}

// ListIndexes returns all indexes for a collection
func (m *MongoDBManager) ListIndexes(collectionName string) ([]bson.M, error) {
	m.logger.Debug("listing indexes", "collection", collectionName)
	collection := m.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		m.logger.Error("failed to list indexes", "collection", collectionName, "error", err)
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		m.logger.Error("failed to decode indexes", "collection", collectionName, "error", err)
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	return indexes, nil
}

// DropIndex drops a specific index by name
func (m *MongoDBManager) DropIndex(collectionName, indexName string) error {
	m.logger.Debug("dropping index", "collection", collectionName, "index", indexName)
	collection := m.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.Indexes().DropOne(ctx, indexName)
	if err != nil {
		m.logger.Error("failed to drop index", "collection", collectionName, "index", indexName, "error", err)
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	m.logger.Info("index dropped", "collection", collectionName, "index", indexName)
	return nil
}

// Aggregate executes an aggregation pipeline on a collection
func (m *MongoDBManager) Aggregate(ctx context.Context, collectionName string, pipeline interface{}) (*mongo.Cursor, error) {
	m.logger.Debug("executing aggregation", "collection", collectionName)
	collection := m.db.Collection(collectionName)

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		m.logger.Error("aggregation failed", "collection", collectionName, "error", err)
		return nil, infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	return cursor, nil
}

// In your repository or collection wrapper
func (m *MongoDBManager) convertFilterToMongoTypes(filter map[string]any) {
	if value, ok := filter["_id"]; ok {
		if id, ok := value.(string); ok {
			objectID, err := primitive.ObjectIDFromHex(id)
			if err == nil {
				filter["_id"] = objectID
			}
		}
	}
}
