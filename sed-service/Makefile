.PHONY: run build test lint swagger migrate-up up down

DB_DSN ?= postgres://sed:sed@localhost:5434/sed?sslmode=disable

run:
	CONFIG_PATH=./configs/local.yaml go run ./cmd/sed-service

build:
	mkdir -p bin
	go build -o bin/sed-service ./cmd/sed-service

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g cmd/sed-service/main.go -o docs

migrate-up:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path migrations -database "$(DB_DSN)" up

up:
	docker compose up -d --build

down:
	docker compose down -v
