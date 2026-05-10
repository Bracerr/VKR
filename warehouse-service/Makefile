.PHONY: run build test test-e2e lint swagger migrate-up migrate-down up down

DB_DSN ?= postgres://wh:wh@localhost:5433/warehouse?sslmode=disable

run:
	CONFIG_PATH=./configs/local.yaml go run ./cmd/warehouse-service

build:
	mkdir -p bin
	go build -o bin/warehouse-service ./cmd/warehouse-service

test:
	go test ./... -race -count=1

test-e2e:
	cd e2e_tests && pytest -q

lint:
	golangci-lint run ./...

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g cmd/warehouse-service/main.go -o docs

migrate-up:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path migrations -database "$(DB_DSN)" up

migrate-down:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path migrations -database "$(DB_DSN)" down 1

up:
	docker compose up -d --build

down:
	docker compose down -v
