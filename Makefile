.PHONY: dev test build deploy clean

# Development
dev:
	docker-compose up -d

# Stop all services
stop:
	docker-compose down

# Logs
logs:
	docker-compose logs -f

# Testing
test:
	cd apps/api && go test ./...
	cd packages/ssh-core && cargo test
	cd apps/web && npm test

# Build production
build:
	cd apps/api && go build -o ../../bin/api cmd/server/main.go
	cd packages/ssh-core && cargo build --release
	cd apps/web && npm run build

# Deploy production manually (no K8s/Terraform in this release)
deploy:
	@echo "Production deploy: use docker-compose.prod.yml with your .env"
	@echo "docker compose -f docker-compose.prod.yml up -d"

# Clean
clean:
	docker-compose down -v
	rm -rf apps/web/.next apps/web/node_modules
	rm -rf packages/ssh-core/target

# Database
db-migrate:
	cd apps/api && go run cmd/migrate/main.go

db-seed:
	cd apps/api && go run cmd/seed/main.go
