package mongo

import (
	"context"
	"errors"
	"time"

	db "erp.localhost/internal/db"
	logging "erp.localhost/internal/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBManager struct {
	client *mongo.Client
	dbName DBName
	db     *mongo.Database
	logger *logging.Logger
}

func NewMongoDBManager(dbName DBName) *MongoDBManager {
	m := &MongoDBManager{
		dbName: dbName,
		logger: logging.NewLogger(logging.ModuleDB),
	}
	if _, ok := dbToCollection[string(dbName)]; !ok {
		m.logger.Fatal("db not found", "db", dbName)
		return nil
	}
	if err := m.Init(); err != nil {
		return nil
	}
	return m
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
	if err := m.createDBAndCollections(); err != nil {
		return err
	}
	return nil
}

func (m *MongoDBManager) createDBAndCollections() error {
	m.db = m.client.Database(string(m.dbName))
	existingCollections, err := m.db.ListCollectionNames(context.Background(), nil)
	if err != nil {
		m.logger.Fatal("failed to list collections", "db", m.dbName, "error", err)
		return err
	}
	missingCollections := db.SlicesDiff(dbToCollection[string(m.dbName)], existingCollections)
	for _, collection := range missingCollections {
		if err := m.db.CreateCollection(context.Background(), collection); err != nil {
			m.logger.Fatal("failed to create collection", "db", m.dbName, "collection", collection, "error", err)
			return err
		}
	}
	return nil
}

func (m *MongoDBManager) Create(collectionName string, data any) (string, error) {
	collection := m.db.Collection(collectionName)
	result, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(string), nil
}

func (m *MongoDBManager) Find(collectionName string, filter map[string]any) ([]any, error) {
	collection := m.db.Collection(collectionName)
	if filter == nil {
		return nil, errors.New("filter is required and cannot be nil")
	}
	if _, ok := filter["tenant_id"]; !ok {
		return nil, errors.New("tenant id is required")
	}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	var results []any
	if err := cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (m *MongoDBManager) Update(collectionName string, filter map[string]any, data any) error {
	collection := m.db.Collection(collectionName)
	_, err := collection.UpdateOne(context.Background(), filter, bson.M{"$set": data})
	if err != nil {
		return err
	}
	return nil
}

func (m *MongoDBManager) Delete(collectionName string, filter map[string]any) error {
	collection := m.db.Collection(collectionName)
	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}
