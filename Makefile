.PHONY: up down build seed logs

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
