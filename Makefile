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

migrate-up:
	go run cmd/migrate/main.go -action=up

migrate-down:
	go run cmd/migrate/main.go -action=down -steps=1

migrate-status:
	go run cmd/migrate/main.go -action=status

migrate-version:
	go run cmd/migrate/main.go -action=version

seed:
	go run cmd/api/main.go
