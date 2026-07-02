.PHONY: build run test test-unit test-integration test-e2e lint fmt migrate-up migrate-down migrate-version migrate-force db-reset docker-up docker-down docker-build clean swagger

APP_NAME := ecommerce-api
BUILD_DIR := bin
MAIN_PKG := ./cmd/api
DOCKER_COMPOSE := docker compose -f docker/docker-compose.yml

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PKG)

run: build
	./$(BUILD_DIR)/$(APP_NAME)

dev:
	go run $(MAIN_PKG)

test:
	go test -race -count=1 ./...

test-unit:
	go test -race -count=1 -short ./...

test-integration:
	go test -race -count=1 -tags=integration ./tests/integration/...

test-e2e:
	go test -race -count=1 -tags=e2e ./tests/e2e/...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

DB_DSN ?= postgres://admin:admin@localhost:5432/ecommerce-db?sslmode=disable

migrate-up:
	migrate -path migrations -database "$(DB_DSN)" up

migrate-down:
	migrate -path migrations -database "$(DB_DSN)" down 1

migrate-version:
	migrate -path migrations -database "$(DB_DSN)" version

# Set schema_migrations.version when files were squashed (e.g. VERSION=2)
migrate-force:
	@test -n "$(VERSION)" || (echo "Usage: make migrate-force VERSION=2" && exit 1)
	migrate -path migrations -database "$(DB_DSN)" force $(VERSION)

# Drop all objects and re-apply migrations via the API (requires RUN_MIGRATIONS=true in .env)
db-reset:
	psql "$(DB_DSN)" -v ON_ERROR_STOP=1 -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "Schema dropped. Run 'make dev' to apply migrations on startup."

docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

docker-build:
	$(DOCKER_COMPOSE) build

docker-logs:
	$(DOCKER_COMPOSE) logs -f api

clean:
	rm -rf $(BUILD_DIR)
	go clean -cache

tidy:
	go mod tidy

vet:
	go vet ./...

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init \
		-g cmd/api/main.go \
		-o docs/swagger \
		--parseDependency \
		--parseInternal
