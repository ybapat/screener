.PHONY: help up down build seed logs logs-backend logs-frontend migrate-up reset clean ps

# Default target: show help
.DEFAULT_GOAL := help

# Show available commands
help:
	@echo "Available commands:"
	@echo "  make up              - Start all services (build + detached)"
	@echo "  make down            - Stop all services"
	@echo "  make build           - Rebuild and restart all services"
	@echo "  make logs            - View logs from all services"
	@echo "  make logs-backend    - View backend logs only"
	@echo "  make logs-frontend   - View frontend logs only"
	@echo "  make seed            - Populate database with test data"
	@echo "  make migrate-up      - Run database migrations manually"
	@echo "  make reset           - Destroy volumes and rebuild from scratch"
	@echo "  make clean           - Remove stopped containers and dangling images"
	@echo "  make ps              - Show running containers"

# Start all services
up:
	docker compose up --build -d

# Stop all services
down:
	docker compose down

# Rebuild and restart
build:
	docker compose up --build -d

# View logs
logs:
	docker compose logs -f

# View backend logs
logs-backend:
	docker compose logs -f backend

# View frontend logs
logs-frontend:
	docker compose logs -f frontend

# Run seed data (requires backend to be running)
seed:
	cd scripts && go run seed.go

# Run database migrations manually
migrate-up:
	docker compose run --rm migrate

# Reset database
reset:
	docker compose down -v
	docker compose up --build -d

# Clean up Docker artifacts
clean:
	docker compose down
	docker system prune -f

# Show running containers
ps:
	docker compose ps
