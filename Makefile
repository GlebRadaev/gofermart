ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

MIGRATIONS_PATH ?= migrations
GOOSE_BIN=goose

DB_DSN=postgres://${DATABASE_USERNAME}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=disable

ifeq ($(strip $(DB_DSN)),)
$(error DATABASE_DSN is not set. Please configure DATABASE_DSN in the .env file)
endif

.PHONY: migrate-up
migrate-up:
	@echo "Running migrations up..."
	$(GOOSE_BIN) -dir $(MIGRATIONS_PATH) postgres "$(DB_DSN)" up

.PHONY: migrate-down
migrate-down:
	@echo "Reverting migrations..."
	$(GOOSE_BIN) -dir $(MIGRATIONS_PATH) postgres "$(DB_DSN)" down

.PHONY: create-migration
create-migration:
	@read -p "Enter migration name: " name; \
	$(GOOSE_BIN) create "$$name" sql -dir $(MIGRATIONS_PATH)


.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: coverage
coverage:
	@echo "Running coverage..." 
	go test ./... -v -parallel=32 -coverprofile=coverage.txt -covermode=atomic && go tool cover -html=coverage.txt && rm -rf coverage.txt

.PHONY: build
build:
	@echo "Building binary..."
	cd cmd/gophermart && go build .

.PHONY: run
run:
	@echo "Running binary..."
	cd cmd/gophermart && go run .

.PHONY: swagger
swagger: 
	@echo "Generating swagger docs..."
	swag fmt
	swag init -g cmd/gophermart/main.go 


MOCKGEN := mockgen
PROJECT_ROOT := $(shell pwd)

SERVICES := balanceservice orderservice authservice
SERVICE_DIR := internal/service

.PHONY: generate-mocks-service
generate-mocks-service:
	@echo "Generating mocks for services..."
	@for SERVICE in $(SERVICES); do \
	  SRC_FILE=$(SERVICE_DIR)/$$SERVICE/$$SERVICE.go; \
	  DEST_FILE=$(SERVICE_DIR)/$$SERVICE/$$SERVICE"_mock.go"; \
	  $(MOCKGEN) \
	    -source=$(PROJECT_ROOT)/$$SRC_FILE \
	    -destination=$(PROJECT_ROOT)/$$DEST_FILE \
	    -package=$$SERVICE; \
	  if [ $$? -eq 0 ]; then \
	    echo "Generated mock for $$SERVICE at $$DEST_FILE"; \
	  else \
	    echo "Failed to generate mock for $$SERVICE"; \
	  fi; \
	done
	@echo "All mocks for services generated!"

HANDLERS := auth balance orders
HANDLER_DIR := internal/handlers

.PHONY: generate-mocks-handler
generate-mocks-handler:
	@echo "Generating mocks for handlers..."
	@for HANDLER in $(HANDLERS); do \
	  SRC_FILE=$(HANDLER_DIR)/$$HANDLER/$$HANDLER.go; \
	  DEST_FILE=$(HANDLER_DIR)/$$HANDLER/$$HANDLER"_mock.go"; \
	  $(MOCKGEN) \
	    -source=$(PROJECT_ROOT)/$$SRC_FILE \
	    -destination=$(PROJECT_ROOT)/$$DEST_FILE \
	    -package=$$HANDLER; \
	  if [ $$? -eq 0 ]; then \
	    echo "Generated mock for $$HANDLER at $$DEST_FILE"; \
	  else \
	    echo "Failed to generate mock for $$HANDLER"; \
	  fi; \
	done
	@echo "All mocks for handlers generated!"
