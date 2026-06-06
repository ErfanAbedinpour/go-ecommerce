.PHONY: build run test test-unit test-integration lint fmt migrate-up migrate-down docker-up docker-down docker-build clean

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

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

migrate-up:
	migrate -path migrations -database "postgres://ecommerce:ecommerce@localhost:5432/ecommerce?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://ecommerce:ecommerce@localhost:5432/ecommerce?sslmode=disable" down 1

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
