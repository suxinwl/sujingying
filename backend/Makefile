.PHONY: dev build test tidy

dev:
	go run ./cmd/server/main.go

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy
