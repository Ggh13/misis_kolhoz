.PHONY: build run up down clean test

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

clean:
	docker-compose down -v

run:
	go run ./cmd/server

test:
	go test -v ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

deps:
	go mod download
