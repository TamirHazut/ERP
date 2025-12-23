# Internal Event Bus

This is the decoupling mechanism.

## Characteristics

* In-process
* Asynchronous
* Fire-and-forget (initially)

## Frameworks

* Kafka
* Go Lang
    * Confluent Kafka

## Dependencies

* Config
* Core
* DB
* Logging

# Side-effect modules (event consumers)

## Observability Module

### Purpose

* Business metrics
* Usage statistics
* Performance counters

### Frameworks

* Go Lang
    * prometheus/client_golang
    * OpenTelemetry

### Consumes

* Domain events
* Request metadata

### Produces

* Aggregates
* Time-series data
* Dashboards inputs

## Alarms / Thresholds Module

### Purpose

* Inventory alerts
* SLA violations
* Config-based triggers

### Consumes

* Inventory events
* Config thresholds

### Produces

* AlarmRaised events

## Notifications & Webhooks Module

### Purpose

* Email notifications
* Webhook callbacks
* External callbacks

### Consumes

* AlarmRaised
* InventoryRestocked

### Rules

* Retry-safe
* Idempotent
* Never blocks core logic