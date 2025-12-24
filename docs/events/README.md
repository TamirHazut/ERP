# Events Service

## Properties
- Port: 5003

## Responsibilities:
- Event ingestion from Kafka
- Event processing & routing
- Notifications (Email, SMS, Push)
- Audit logging
- Alerting & monitoring
- Observability metrics

## Tech Stack:
- Go
- Kafka consumer (sarama/confluent-kafka-go)
- OpenTelemetry
- Prometheus metrics
- Integration with notification providers

## Event Types:
- user.created
- order.placed
- product.updated
- vendor.approved
- system.alert