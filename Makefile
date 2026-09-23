.PHONY: build run test t clean mock fmt lint

# Variables
BINARY_NAME=bookerbot
INSTANCE ?= local
GO=$(shell which go)

# Build the application
build:
	go build -o bin/$(BINARY_NAME) cmd/bookerbot/main.go

# Run the application
run:
	INSTANCE=$(INSTANCE) go run cmd/bookerbot/main.go

# Test the application
test: t
	$(GO) test -v $(shell go list ./... | grep -v /vendor/)

t:
	$(GO) test -v ./... -count=2 -race -coverprofile=cover.out.tmp $(RUN_ARGS)

# Clean build artifacts
clean:
	go clean
	rm -rf bin/

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy
