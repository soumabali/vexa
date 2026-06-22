.PHONY: build test clean run migrate migrate-up migrate-down migrate-status migrate-seed migrate-reset lint db-logs

BINARY_NAME=vexa
BUILD_DIR=./build

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

test:
	go test -v -race -coverprofile=coverage.out ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

run:
	go run ./cmd/server

migrate:
	@echo "Use 'make migrate-up' or 'make migrate-status' instead"

migrate-up:
	go run -tags ignore cmd/migrate/main.go -dsn "$(DATABASE_URL)" -migrations ./internal/db/migrations up

migrate-down:
	go run -tags ignore cmd/migrate/main.go -dsn "$(DATABASE_URL)" -migrations ./internal/db/migrations down

migrate-status:
	go run -tags ignore cmd/migrate/main.go -dsn "$(DATABASE_URL)" -migrations ./internal/db/migrations status

migrate-seed:
	go run -tags ignore cmd/migrate/main.go -dsn "$(DATABASE_URL)" -migrations ./internal/db/migrations -seeds ./internal/db/seeds seed

migrate-reset:
	go run -tags ignore cmd/migrate/main.go -dsn "$(DATABASE_URL)" -migrations ./internal/db/migrations -seeds ./internal/db/seeds reset

migrate-rollback-to:
	@if [ -z "$(VERSION)" ]; then echo "VERSION required: make migrate-rollback-to VERSION=001_initial"; exit 1; fi
	go run -tags ignore cmd/migrate/main.go -dsn "$(DATABASE_URL)" -migrations ./internal/db/migrations down-to -version $(VERSION)

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

lint-fast:
	golangci-lint run --fast ./...

lint-ci:
	golangci-lint run --out-format=colored-line-number ./...

lint-new:
	golangci-lint run --new ./...

lint-diff:
	golangci-lint run --new-from-rev=origin/main ./...

dev:
	air -c .air.toml

docker-build:
	docker build -t vexa:latest .

docker-run:
	docker-compose up -d

db-logs:
	docker-compose exec postgres psql -U vexa -d vexa -c "SELECT * FROM schema_migrations ORDER BY applied_at;"
