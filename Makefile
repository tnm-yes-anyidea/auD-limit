SHELL := /bin/bash
BINARY := bin/auD-limit

.PHONY: build fmt tidy test

build:
	@echo "Building..."
	@mkdir -p bin
	go build -o $(BINARY) ./cmd/auD-limit

fmt:
	gofmt -w .

tidy:
	go mod tidy

test:
	go test ./...