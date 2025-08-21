# Product Views Service

A high-performance Go service for tracking and analyzing product views with asynchronous processing via Kafka and efficient top-N product retrieval using PostgreSQL. The service uses PostgreSQL for data persistence and Kafka for asynchronous processing of view events.

## Features

- Track product views with high throughput using Kafka
- Get top N most viewed products (max 100)
- Swagger API documentation
- Containerized with Docker
- Database migrations

## Prerequisites

- Docker and Docker Compose
- Go 1.23 or later (for local development)

## Getting Started

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd Handpickd-Assignment
   ```

2. **Start the services**
   ```bash
   docker-compose up -d
   ```
   This will start:
   - PostgreSQL database
   - Kafka with ZooKeeper
   - Product Views API service

3. **Access the services**
   - API: http://localhost:8080
   - Swagger UI: http://localhost:8080/swagger/index.html
   - PostgreSQL: localhost:5432 (user: postgres, password: postgres)

## API Endpoints

### Record a Product View
```
POST /api/v1/products/view
Content-Type: application/json

{
  "product_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Get Top Viewed Products
```
GET /api/v1/products/top?limit=10
```

### Get Product by ID
```
GET /api/v1/products/{id}
```

### Create a New Product
```
POST /api/v1/products
Content-Type: application/json

{
  "name": "Example Product",
  "description": "This is an example product",
  "view_count": 0
}
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

