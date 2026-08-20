.PHONY: run build test clean migrate seed

APP_NAME=wedding-invitation-backend
BUILD_DIR=bin

run:
	go run cmd/api/main.go

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

deps:
	go mod tidy
	go mod download

lint:
	golangci-lint run

seed:
	go run cmd/api/main.go
