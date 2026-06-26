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
	cd apps/web && npm test

# Build production artifacts (no Rust/Flutter required)
build:
	cd apps/api && go build -o ../../bin/api cmd/server/main.go
	cd apps/web && npm run build

# Verification gates (no Rust/Flutter required)
verify:
	cd apps/api && go test ./...
	cd apps/api && go build ./...
	cd apps/web && npm run build
	cp -r apps/web/.next/static apps/web/.next/standalone/.next/static

# Deploy production (images built by GitHub Actions → GHCR)
deploy:
	@echo "Pull latest images from GHCR..."
	docker compose -f docker-compose.prod.yml pull
	@echo "Restart services..."
	docker compose -f docker-compose.prod.yml up -d

# Clean
clean:
	docker compose down -v
	rm -rf apps/web/.next apps/web/node_modules

# Database
db-migrate:
	cd apps/api && go run cmd/migrate/main.go

db-seed:
	cd apps/api && go run cmd/seed/main.go

# Backup (production)
backup:
	docker compose -f docker-compose.prod.yml run --rm backup

# Backup cron install (production host)
backup-cron:
	./scripts/install-backup-cron.sh
