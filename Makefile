.PHONY: dev stop logs test build verify deploy clean db-migrate db-seed

# Development
dev:
	docker compose up -d

# Stop all services
stop:
	docker compose down

# Logs
logs:
	docker compose logs -f

# Testing
test:
	cd apps/api && go test ./...
	cd packages/ssh-core && cargo test
	cd apps/web && npm test

# Build production artifacts
build:
	cd apps/api && go build -o ../../bin/api cmd/server/main.go
	cd packages/ssh-core && cargo build --release
	cd apps/web && npm run build

# Verification gates (no Rust/Flutter required)
verify:
	cd apps/api && go test ./...
	cd apps/api && go build ./...
	cd apps/web && npm run build
	cp -r apps/web/.next/static apps/web/.next/standalone/.next/static

# Deploy production manually (self-hosted Docker Compose only)
deploy:
	@echo "Production deploy: use docker-compose.prod.yml with your .env"
	@echo "docker compose -f docker-compose.prod.yml up -d"

# Clean
clean:
	docker compose down -v
	rm -rf apps/web/.next apps/web/node_modules
	rm -rf packages/ssh-core/target

# Database
db-migrate:
	cd apps/api && go run cmd/migrate/main.go

db-seed:
	cd apps/api && go run cmd/seed/main.go
