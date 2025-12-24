# Config Service

## Properties
- Database: MongoDB (Collection: config_db)
- Port: 5002

## Responsibilities:
- Centralized configuration management
- Feature flags
- Environment-specific configs
- Dynamic configuration updates
- Config versioning
- Config validation

## Tech Stack:
- Go
- MongoDB for persistence
- Redis for caching
- gRPC API

## Config Structure:
- Service-specific configs
- Global settings
- Feature flags
- Environment variables