# Gateway API Service

## Properties
- Port: 4000

## Responsibilities:
- Single entry point for WebUI
- Query/Mutation routing
- Authentication middleware
- Rate limiting & throttling
- Response caching (Redis)
- Request aggregation

## Tech Stack:
- gqlgen (Go GraphQL library)
- Chi/Gin router
- Redis for caching