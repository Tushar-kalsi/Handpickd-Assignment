.PHONY: help build run test clean docker-up docker-down seed-db swagger

help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application locally"
	@echo "  make test        - Run all tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker-up   - Start all services with Docker Compose"
	@echo "  make docker-down - Stop all Docker services"
	@echo "  make seed-db     - Seed the database with sample data"
	@echo "  make swagger     - Generate Swagger documentation"

build:
	go build -o bin/product-views cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/ coverage.out coverage.html

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

seed-db:
	docker exec -i product-views-postgres psql -U postgres -d product_views < scripts/seed.sql

swagger:
	swag init -g cmd/api/main.go -o docs

deps:
	go mod download
	go mod tidy

lint:
	golangci-lint run

migrate-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/product_views?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/product_views?sslmode=disable" down
