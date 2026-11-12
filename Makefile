.PHONY: all build build-web build-go test test-cover run clean

all: build test

build: build-web build-go

build-web:
	@echo "Building React 19 web assets..."
	cd web && npm run build

build-go:
	@echo "Building single Go binary..."
	go build -ldflags="-s -w" -o feedbackloom ./cmd/feedbackloom

test:
	@echo "Running all unit and integration tests..."
	go test -v ./...

test-cover:
	@echo "Running tests with coverage analysis..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	./feedbackloom serve --port 8080 --db ./feedbackloom.db

clean:
	@echo "Cleaning artifacts..."
	rm -f feedbackloom feedbackloom.exe coverage.out
	rm -rf web/dist
