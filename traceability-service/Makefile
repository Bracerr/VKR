.PHONY: run build test lint migrate-up up down

DB_DSN ?= postgres://trace:trace@localhost:5438/trace?sslmode=disable

run:
	CONFIG_PATH=./configs/local.yaml go run ./cmd/traceability-service

build:
	mkdir -p bin
	go build -o bin/traceability-service ./cmd/traceability-service

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

migrate-up:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path migrations -database "$(DB_DSN)" up

up:
	docker compose up -d --build

down:
	docker compose down -v

