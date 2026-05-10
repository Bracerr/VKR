.PHONY: run build test lint migrate-up up down

DB_DSN ?= postgres://proc:proc@localhost:5436/procurement?sslmode=disable

run:
	CONFIG_PATH=./configs/local.yaml go run ./cmd/procurement-service

build:
	mkdir -p bin
	go build -o bin/procurement-service ./cmd/procurement-service

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

