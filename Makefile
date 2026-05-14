ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ROOT := $(ROOT:/=)

COMPOSE_ENV :=
ifneq (,$(wildcard $(ROOT)/.env))
COMPOSE_ENV := --env-file $(ROOT)/.env
endif

COMPOSE_FULLSTACK ?= sed-service/e2e_tests/docker-compose.test.yml

LEGACY_CPU ?=
COMPOSE_LEGACY :=
ifneq ($(strip $(LEGACY_CPU)),)
COMPOSE_LEGACY := -f sed-service/e2e_tests/docker-compose.legacy-cpu.yml
endif

DOCKER_COMPOSE := docker compose

.DEFAULT_GOAL := test-config

.PHONY: test-up test-down test-ps test-logs test-build test-config \
	prod-up prod-down prod-ps prod-logs prod-build prod-config

test-up:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-test up -d --build

test-down:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-test down -v

test-ps:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-test ps

test-logs:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-test logs -f --tail=200

test-build:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-test build

test-config:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-test config >/dev/null && echo "OK: vkr-test compose valid"

prod-up:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-prod up -d --build

prod-down:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-prod down -v

prod-ps:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-prod ps

prod-logs:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-prod logs -f --tail=200

prod-build:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-prod build

prod-config:
	cd $(ROOT) && $(DOCKER_COMPOSE) $(COMPOSE_ENV) -f $(COMPOSE_FULLSTACK) $(COMPOSE_LEGACY) -p vkr-prod config >/dev/null && echo "OK: vkr-prod compose valid"
