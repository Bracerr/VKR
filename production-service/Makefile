.PHONY: run build test lint swagger migrate-up up down mocks

DB_DSN ?= postgres://prod:prod@localhost:5435/prod?sslmode=disable

run:
	CONFIG_PATH=./configs/local.yaml go run ./cmd/production-service

build:
	mkdir -p bin
	go build -o bin/production-service ./cmd/production-service

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g cmd/production-service/main.go -o docs

migrate-up:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path migrations -database "$(DB_DSN)" up

mocks:
	go run go.uber.org/mock/mockgen@v0.4.0 -destination=internal/usecases/mocks_test.go -package=usecases_test github.com/industrial-sed/production-service/internal/usecases WarehouseClient,SEDClient

up:
	docker compose up -d --build

down:
	docker compose down -v
