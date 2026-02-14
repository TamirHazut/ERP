package event

import "erp.localhost/infra/model/event"

// TopicConfig holds configuration for a Kafka topic
type TopicConfig struct {
	NumPartitions     int
	ReplicationFactor int
	RetentionMs       int64  // -1 = forever, else milliseconds
	CleanupPolicy     string // "delete" or "compact"
	CompressionType   string // "snappy", "gzip", "lz4", etc.
}

// Predefined topic configuration templates
var (
	// HighThroughputConfig for topics with high message volume
	HighThroughputConfig = TopicConfig{
		NumPartitions:     10,
		ReplicationFactor: 3,
		RetentionMs:       7 * 24 * 60 * 60 * 1000, // 7 days
		CleanupPolicy:     "delete",
		CompressionType:   "snappy",
	}

	// MediumThroughputConfig for topics with moderate message volume
	MediumThroughputConfig = TopicConfig{
		NumPartitions:     5,
		ReplicationFactor: 2,
		RetentionMs:       7 * 24 * 60 * 60 * 1000, // 7 days
		CleanupPolicy:     "delete",
		CompressionType:   "snappy",
	}

	// LowThroughputConfig for topics with low message volume
	LowThroughputConfig = TopicConfig{
		NumPartitions:     3,
		ReplicationFactor: 2,
		RetentionMs:       7 * 24 * 60 * 60 * 1000, // 7 days
		CleanupPolicy:     "delete",
		CompressionType:   "snappy",
	}

	// AuditConfig for audit/compliance topics with long retention
	AuditConfig = TopicConfig{
		NumPartitions:     5,
		ReplicationFactor: 3,
		RetentionMs:       90 * 24 * 60 * 60 * 1000, // 90 days
		CleanupPolicy:     "delete",
		CompressionType:   "snappy",
	}

	// SecurityConfig for security events with very long retention
	SecurityConfig = TopicConfig{
		NumPartitions:     5,
		ReplicationFactor: 3,
		RetentionMs:       180 * 24 * 60 * 60 * 1000, // 180 days
		CleanupPolicy:     "delete",
		CompressionType:   "snappy",
	}
)

// GetAllTopicConfigs returns configurations for all topics
func GetAllTopicConfigs() map[event.Topic]TopicConfig {
	return map[event.Topic]TopicConfig{
		// ============================================
		// Auth Domain - User Events
		// ============================================
		event.AuthUserCreated: MediumThroughputConfig,
		event.AuthUserUpdated: MediumThroughputConfig,
		event.AuthUserDeleted: LowThroughputConfig,
		event.AuthUserLogin:   HighThroughputConfig,
		event.AuthUserLogout:  MediumThroughputConfig,

		// ============================================
		// Auth Domain - Authentication Events
		// ============================================
		event.AuthTokenRefreshed: HighThroughputConfig,
		event.AuthTokenRevoked:   MediumThroughputConfig,

		// ============================================
		// Auth Domain - RBAC Events
		// ============================================
		event.AuthRoleCreated:  LowThroughputConfig,
		event.AuthRoleUpdated:  LowThroughputConfig,
		event.AuthRoleDeleted:  LowThroughputConfig,
		event.AuthRoleAssigned: MediumThroughputConfig,
		event.AuthRoleRevoked:  MediumThroughputConfig,

		event.AuthPermissionCreated: LowThroughputConfig,
		event.AuthPermissionUpdated: LowThroughputConfig,
		event.AuthPermissionDeleted: LowThroughputConfig,

		// ============================================
		// Auth Domain - Tenant Events
		// ============================================
		event.AuthTenantCreated:       LowThroughputConfig,
		event.AuthTenantUpdated:       LowThroughputConfig,
		event.AuthTenantDeleted:       LowThroughputConfig,
		event.AuthTenantTokensRevoked: LowThroughputConfig,

		// ============================================
		// Core Domain - Order Events (when added)
		// ============================================
		// event.TopicOrderCreated:  HighThroughputConfig,
		// event.TopicOrderUpdated:  HighThroughputConfig,
		// event.TopicOrderShipped:  MediumThroughputConfig,

		// Add more topics as needed...
	}
}

// GetTopicConfig returns configuration for a specific topic
func GetTopicConfig(topic event.Topic) (TopicConfig, bool) {
	configs := GetAllTopicConfigs()
	config, exists := configs[topic]
	return config, exists
}

// Helper functions for common retention periods
func RetentionDays(days int) int64 {
	return int64(days) * 24 * 60 * 60 * 1000
}

func RetentionHours(hours int) int64 {
	return int64(hours) * 60 * 60 * 1000
}

// Custom topic configuration builder
type TopicConfigBuilder struct {
	config TopicConfig
}

func NewTopicConfigBuilder() *TopicConfigBuilder {
	return &TopicConfigBuilder{
		config: MediumThroughputConfig, // Start with defaults
	}
}

func (b *TopicConfigBuilder) WithPartitions(n int) *TopicConfigBuilder {
	b.config.NumPartitions = n
	return b
}

func (b *TopicConfigBuilder) WithReplication(n int) *TopicConfigBuilder {
	b.config.ReplicationFactor = n
	return b
}

func (b *TopicConfigBuilder) WithRetentionDays(days int) *TopicConfigBuilder {
	b.config.RetentionMs = RetentionDays(days)
	return b
}

func (b *TopicConfigBuilder) WithCompression(compressionType string) *TopicConfigBuilder {
	b.config.CompressionType = compressionType
	return b
}

func (b *TopicConfigBuilder) Build() TopicConfig {
	return b.config
}
