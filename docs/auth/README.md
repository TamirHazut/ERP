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