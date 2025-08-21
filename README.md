# Product Views Service

A high-performance Go service for tracking and analyzing product views with asynchronous processing via Kafka and efficient top-N product retrieval using PostgreSQL.

## Table of Contents
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start with Docker](#quick-start-with-docker)
- [API Documentation](#api-documentation)
- [Swagger Documentation](#swagger-documentation)
- [pgAdmin Dashboard](#pgadmin-dashboard)
- [SQL Queries](#sql-queries)
- [cURL Request Examples](#curl-request-examples)
- [Development](#development)
- [Architecture](#architecture)
- [Troubleshooting](#troubleshooting)

## Features

- Track product views with high throughput using Kafka
- Get top N most viewed products (max 100)
- Swagger API documentation
- Containerized with Docker
- Database migrations

## Prerequisites

- Docker and Docker Compose
- Go 1.23 or later (for local development)

## Quick Start with Docker

### 1. Clone the repository
```bash
git clone <repository-url>
cd Handpickd-Assignment
```

### 2. Start the services
```bash
docker-compose up -d
```
This will start:
- PostgreSQL database
- Kafka with ZooKeeper
- Product Views API service

### 3. Verify services are running
```bash
docker-compose ps
docker-compose logs -f
```

### 4. Access the services
| Service | URL | Credentials |
|---------|-----|-------------|
| API | http://localhost:8080 | - |
| Swagger UI | http://localhost:8080/swagger/index.html | - |
| pgAdmin | http://localhost:5050 | Email: admin@handpickd.com<br>Password: admin |
| PostgreSQL | localhost:5432 | User: postgres<br>Password: postgres<br>Database: product_views |

## API Documentation

### Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/products/view` | Record a product view |
| GET | `/api/v1/products/top` | Get top N most viewed products |
| GET | `/api/v1/products/{id}` | Get product by ID |
| POST | `/api/v1/products` | Create a new product |

## Swagger Documentation

### Accessing Swagger UI
1. Ensure the service is running
2. Navigate to: http://localhost:8080/swagger/index.html
3. Explore and test API endpoints directly from the browser

### Generating/Updating Swagger Docs
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/api/main.go -o docs
```

## pgAdmin Dashboard

### Initial Setup
1. Access pgAdmin at http://localhost:5050
2. Login with:
   - Email: `admin@handpickd.com`
   - Password: `admin`

## cURL Request Examples

### 1. Record a Product View
```bash
curl -X POST http://localhost:8080/api/v1/products/view \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "550e8400-e29b-41d4-a716-446655440001"
  }'
```

### 2. Get Top Viewed Products
```bash
curl -X GET http://localhost:8080/api/v1/products/top
curl -X GET "http://localhost:8080/api/v1/products/top?limit=5"
curl -X GET "http://localhost:8080/api/v1/products/top?limit=20"
```

### 3. Get Product by ID
```bash
curl -X GET http://localhost:8080/api/v1/products/550e8400-e29b-41d4-a716-446655440001
```

### 4. Create a New Product
```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New MacBook Air M3",
    "description": "Latest MacBook Air with M3 chip and 15-inch display",
    "view_count": 0
  }'
```

### 5. Health Check
```bash
curl -X GET http://localhost:8080/health
```

## Development

### Running Tests
```bash
go test -v ./...
```

### Running Migrations Manually
```bash
docker-compose exec product-views /app/product-views migrate
```

### Stopping Services
```bash
docker-compose down
```

## Architecture

### Components
- **API Layer**: Handles HTTP requests and responses
- **Kafka Producer**: Publishes view events to Kafka
- **Kafka Consumer**: Consumes view events and updates the database
- **Repository Layer**: Handles database operations
- **Database**: PostgreSQL for data persistence

### Data Flow
1. Client sends a request to record a product view
2. API publishes an event to Kafka
3. Kafka consumer processes the event asynchronously
4. View count is updated in the database
5. Clients can query for top viewed products

## Performance Considerations

- The service is designed to handle high throughput of view events
- View count updates are processed asynchronously via Kafka
- The database is optimized for read-heavy workloads with appropriate indexes
- The top N products query uses an index on the view_count column

## Monitoring and Observability

### Health Check
```
GET /health
```

