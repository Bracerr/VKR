.PHONY: run build test test-cover lint swagger mocks migrate-up migrate-down e2e up down

DB_DSN ?= postgres://auth:auth@localhost:5432/auth?sslmode=disable

run:
	go run ./cmd/auth-service

build:
	mkdir -p bin
	go build -o bin/auth-service ./cmd/auth-service

test:
	go test ./... -race -count=1

test-cover:
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g cmd/auth-service/main.go -o docs

mocks:
	go run go.uber.org/mock/mockgen@v0.4.0 -destination=internal/mocks/keycloak.go -package=mocks github.com/industrial-sed/auth-service/internal/ports KeycloakClient
	go run go.uber.org/mock/mockgen@v0.4.0 -destination=internal/mocks/repos.go -package=mocks github.com/industrial-sed/auth-service/internal/usecases TenantRepository,UserCacheRepository,Notifier

migrate-up:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path migrations -database "$(DB_DSN)" up

migrate-down:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path migrations -database "$(DB_DSN)" down 1

up:
	docker compose up -d --build

down:
	docker compose down -v

e2e:
	docker compose -f e2e_tests/docker-compose.test.yml up -d --build
	python3 -m pip install -q -r e2e_tests/requirements.txt
	AUTH_BASE_URL=http://localhost:18080 TEST_SECRET=e2e-test-secret SERVICE_SECRET=e2e-service-secret python3 -m pytest e2e_tests/tests -v || (docker compose -f e2e_tests/docker-compose.test.yml logs --tail=80 auth-service; exit 1)
	docker compose -f e2e_tests/docker-compose.test.yml down -v
