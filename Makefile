# Makefile for redis-go project

APP_NAME=redis-go
BIN_DIR=bin

.PHONY: all build run fmt vet test clean deps

all: build

build:
	@mkdir -p $(BIN_DIR)
	@go build -v -o $(BIN_DIR)/$(APP_NAME) .

run: build
	@./$(BIN_DIR)/$(APP_NAME)

# Run without building binary (useful during development)
run-local:
	@go run main.go

fmt:
	@gofmt -s -w .

vet:
	@go vet ./...

test:
	@go test ./... -v

deps:
	@go mod tidy

clean:
	@rm -rf $(BIN_DIR)
