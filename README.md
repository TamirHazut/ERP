# ERP

## Purpose
An ERP system with multi-tenancy feature allowing many organizations maintain and control their inventory, vendors, users, customers, orders and much more.

## High Level Design
![High level design](./docs/resources/High%20Level%20Design.jpg)

## Services Docs

### 1. Auth
Doc: [Click Here](./docs/auth/README.md)

### 2. Config
Doc: [Click Here](./docs/config/README.md)

### 3. Core
Doc: [Click Here](./docs/core/README.md)

### 4. DB
Doc: [Click Here](./docs/db/README.md)

### 5. Events
Doc: [Click Here](./docs/events/README.md)

### 6. Gateway
Doc: [Click Here](./docs/gateway/README.md)

### 7. WebUI
Doc: [Click Here](./docs/webui/README.md)

## Flows

### 1. User Login Flow
1. WebUI → Gateway: mutation login(email, password)
2. Gateway → Auth: gRPC Authenticate()
3. Auth → Redis: Validate & create session
4. Auth → Gateway: Return JWT + refresh token
5. Gateway → WebUI: Return tokens + user info
6. WebUI: Store tokens, redirect to dashboard

### 2. Create Order Flow
1. WebUI → Gateway: mutation createOrder(input)
2. Gateway → Auth: Verify JWT token
3. Gateway → Core: gRPC CreateOrder()
4. Core → MongoDB: Insert order document
5. Core → Kafka: Publish "order.created" event
6. Core → Gateway: Return order data
7. Events Service: Consume event → Send notification
8. Gateway → WebUI: Return created order

### 3. Configuration Update Flow
1. Admin changes feature flag in WebUI
2. Gateway → Config: UpdateConfig()
3. Config → MongoDB: Update config document
4. Config → Redis: Invalidate cache
5. Config: Broadcast update to all services
6. Services: Reload configuration