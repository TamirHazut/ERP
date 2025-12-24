# Auth / RBAC Service

## Properties
- Database: Redis  
- Port: 5000

## Responsibilities:
- User authentication (login/logout)
- JWT token generation & validation
- Permission checking
- Role management
- Session management
- Password hashing (bcrypt)

## Tech Stack:
- Go
- JWT-go library
- Redis for sessions/tokens
- gRPC for inter-service communication

## Key Endpoints:
- POST /auth/login
- POST /auth/logout
- POST /auth/refresh
- GET /auth/verify
- POST /rbac/check-permission