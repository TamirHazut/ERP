package event

import (
	"context"
	"fmt"

	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	"github.com/segmentio/kafka-go"
)

// TopicsInitializer handles Kafka topic creation
type TopicsInitializer struct {
	brokers []string
	logger  logger.Logger
}

// NewTopicsInitializer creates a new topics initializer
func NewTopicsInitializer(brokers []string, logger logger.Logger) *TopicsInitializer {
	return &TopicsInitializer{
		brokers: brokers,
		logger:  logger,
	}
}

// EnsureTopicsExist creates all required topics if they don't exist
func (ti *TopicsInitializer) EnsureTopicsExist(ctx context.Context) *infra_error.AppError {
	ti.logger.Info("Ensuring Kafka topics exist...")

	// Get topic configurations
	topicConfigs := GetAllTopicConfigs()

	// Check which topics already exist
	existingTopics, err := ti.getExistingTopics()
	if err != nil {
		return err
	}

	// Filter topics that need to be created
	var topicsToCreate []kafka.TopicConfig
	for topic, config := range topicConfigs {
		topicName := string(topic)

		if existingTopics[topicName] {
			ti.logger.Debug("Topic already exists", "topic", topicName)
			continue
		}

		ti.logger.Info("Preparing to create topic", "topic", topicName)
		topicsToCreate = append(topicsToCreate, toKafkaTopicConfig(topicName, config))
	}

	if len(topicsToCreate) == 0 {
		ti.logger.Info("All topics already exist")
		return nil
	}

	// Create topics
	if err := ti.createTopics(ctx, topicsToCreate); err != nil {
		ti.logger.Fatal("failed to create topics", "error", err)
		return err
	}

	ti.logger.Info("Topics created successfully",
		"created", len(topicsToCreate),
		"total", len(topicConfigs))

	return nil
}

// getExistingTopics returns a map of existing topic names
func (ti *TopicsInitializer) getExistingTopics() (map[string]bool, *infra_error.AppError) {
	conn, err := kafka.Dial("tcp", ti.brokers[0])
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalKafkaError, fmt.Errorf("failed to dial kafka: %w", err))
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalKafkaError, fmt.Errorf("failed to read partitions: %w", err))
	}

	existingTopics := make(map[string]bool)
	for _, p := range partitions {
		existingTopics[p.Topic] = true
	}

	return existingTopics, nil
}

// createTopics creates the specified topics
func (ti *TopicsInitializer) createTopics(ctx context.Context, topicsToCreate []kafka.TopicConfig) *infra_error.AppError {
	// Connect to broker
	conn, err := kafka.Dial("tcp", ti.brokers[0])
	if err != nil {
		return infra_error.Internal(infra_error.InternalKafkaError, fmt.Errorf("failed to dial kafka: %w", err))
	}
	defer conn.Close()

	// Get controller
	controller, err := conn.Controller()
	if err != nil {
		return infra_error.Internal(infra_error.InternalKafkaError, fmt.Errorf("failed to get controller: %w", err))
	}

	// Connect to controller
	controllerConn, err := kafka.Dial("tcp",
		fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return infra_error.Internal(infra_error.InternalKafkaError, fmt.Errorf("failed to dial controller: %w", err))
	}
	defer controllerConn.Close()

	// Set deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		controllerConn.SetDeadline(deadline)
	}

	// Create topics
	err = controllerConn.CreateTopics(topicsToCreate...)
	if err != nil {
		return infra_error.Internal(infra_error.InternalKafkaError, fmt.Errorf("failed to create topics: %w", err))
	}

	return nil
}

// VerifyTopicsExist verifies that all required topics exist
func (ti *TopicsInitializer) VerifyTopicsExist() *infra_error.AppError {
	ti.logger.Info("Verifying Kafka topics...")

	existingTopics, err := ti.getExistingTopics()
	if err != nil {
		return err
	}

	topicConfigs := GetAllTopicConfigs()

	var missingTopics []string
	for topic := range topicConfigs {
		if !existingTopics[string(topic)] {
			missingTopics = append(missingTopics, string(topic))
		}
	}

	if len(missingTopics) > 0 {
		return infra_error.Internal(infra_error.InternalKafkaError, fmt.Errorf("missing topics: %v", missingTopics))
	}

	ti.logger.Info("All topics verified", "count", len(topicConfigs))
	return nil
}

// toKafkaTopicConfig converts TopicConfig to kafka.TopicConfig
func toKafkaTopicConfig(name string, config TopicConfig) kafka.TopicConfig {
	configEntries := []kafka.ConfigEntry{
		{
			ConfigName:  "retention.ms",
			ConfigValue: fmt.Sprintf("%d", config.RetentionMs),
		},
	}

	if config.CleanupPolicy != "" {
		configEntries = append(configEntries, kafka.ConfigEntry{
			ConfigName:  "cleanup.policy",
			ConfigValue: config.CleanupPolicy,
		})
	}

	if config.CompressionType != "" {
		configEntries = append(configEntries, kafka.ConfigEntry{
			ConfigName:  "compression.type",
			ConfigValue: config.CompressionType,
		})
	}

	return kafka.TopicConfig{
		Topic:             name,
		NumPartitions:     config.NumPartitions,
		ReplicationFactor: config.ReplicationFactor,
		ConfigEntries:     configEntries,
	}
}
