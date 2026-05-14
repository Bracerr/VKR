# VKR: запуск полного Docker-стенда (тест / прод-подобный).
# Все пути относительно корня репозитория (где лежит этот Makefile).

ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ROOT := $(ROOT:/=)

# Подстановка `${VAR}` в compose: из корневого `.env`, если файл существует.
COMPOSE_ENV :=
ifneq (,$(wildcard $(ROOT)/.env))
COMPOSE_ENV := --env-file $(ROOT)/.env
endif

# Один compose-файл полного стенда; test/prod различаются только именем проекта (-p).
COMPOSE_FULLSTACK ?= sed-service/e2e_tests/docker-compose.test.yml

DOCKER_COMPOSE := docker compose

.DEFAULT_GOAL := help

.PHONY: help test-up test-down test-ps test-logs test-build test-config \
	prod-up prod-down prod-ps prod-logs prod-build prod-config

help:
	@echo "VKR Docker стенд (тест и прод-подобный — один compose-файл, разные проекты)."
	@echo ""
	@echo "  make test-up       поднять стек (проект vkr-test)"
	@echo "  make test-down     остановить и удалить контейнеры vkr-test"
	@echo "  make test-ps       статус"
	@echo "  make test-logs     логи (follow)"
	@echo "  make test-build    только сборка образов"
	@echo "  make test-config   проверить compose"
	@echo ""
	@echo "  make prod-up       поднять стек (проект vkr-prod, тот же compose-файл)"
	@echo "  make prod-down     остановить vkr-prod"
	@echo "  make prod-ps / prod-logs / prod-build / prod-config — аналогично"
	@echo ""
	@echo "Секреты: скопируйте VKR/.env.example в VKR/.env и задайте значения; make передаёт"
	@echo "  docker compose --env-file при наличии .env в корне репозитория."
	@echo ""
	@echo "Не запускайте test-up и prod-up одновременно на одном хосте (порты 28xxx)."

# --- тестовый стенд (pytest, ручные проверки) ---

test-up:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-test up -d --build

test-down:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-test down -v

test-ps:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-test ps

test-logs:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-test logs -f --tail=200

test-build:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-test build

test-config:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-test config >/dev/null && echo "OK: vkr-test compose valid"

# --- прод-подобный стенд (тот же compose, проект vkr-prod) ---

prod-up:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-prod up -d --build

prod-down:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-prod down -v

prod-ps:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-prod ps

prod-logs:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-prod logs -f --tail=200

prod-build:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-prod build

prod-config:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) -p vkr-prod config >/dev/null && echo "OK: vkr-prod compose valid"
