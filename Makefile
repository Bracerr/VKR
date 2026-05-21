ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ROOT := $(ROOT:/=)

COMPOSE_ENV :=
ifneq (,$(wildcard $(ROOT)/.env))
COMPOSE_ENV := --env-file $(ROOT)/.env
endif

COMPOSE_FULLSTACK ?= sed-service/e2e_tests/docker-compose.test.yml
COMPOSE_PROD_GATEWAY ?= prod-gateway/docker-compose.yaml

DOCKER_COMPOSE := docker compose

.DEFAULT_GOAL := test-config

SERVICES := auth-service warehouse-service sed-service production-service \
	procurement-service sales-service traceability-service platform

.PHONY: test-up test-down test-ps test-logs test-build test-config \
	prod-up prod-down prod-ps prod-logs prod-build prod-config \
	prod-gateway-up prod-gateway-down \
	test-unit test-e2e test-all seed-rag seed-rag-teardown test-rag-fixtures test-e2e-rag

test-unit:
	@for s in $(SERVICES); do \
	  echo "=== $$s ===" && (cd "$(ROOT)/$$s" && GOTOOLCHAIN=local go test ./... -race -count=1) || exit 1; \
	done

test-e2e:
	cd $(ROOT) && $(MAKE) test-up
	cd $(ROOT)/sed-service/e2e_tests && \
	  (test -d .venv || python3 -m venv .venv) && . .venv/bin/activate && \
	  pip install -q -r requirements.txt && \
	  EVENT_TRANSPORT=$${EVENT_TRANSPORT:-kafka} KAFKA_CONSUMER_OFFSET=earliest \
	  pytest -q tests
	cd $(ROOT) && $(MAKE) test-down

test-all: test-unit test-e2e

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

prod-gateway-up:
	cd $(ROOT) && $(DOCKER_COMPOSE) -f $(COMPOSE_PROD_GATEWAY) up -d

prod-gateway-down:
	cd $(ROOT) && $(DOCKER_COMPOSE) -f $(COMPOSE_PROD_GATEWAY) down

# RAG fixtures (требует RAG_FIXTURES_ENABLED=true в .env)
seed-rag:
	@test "$${RAG_FIXTURES_ENABLED}" = "true" || (echo "Set RAG_FIXTURES_ENABLED=true in .env"; exit 1)
	cd $(ROOT) && python3 scripts/seed_rag_fixtures.py

seed-rag-teardown:
	@test "$${RAG_FIXTURES_ENABLED}" = "true" || (echo "Set RAG_FIXTURES_ENABLED=true in .env"; exit 1)
	cd $(ROOT) && python3 scripts/seed_rag_fixtures.py --teardown

test-rag-fixtures:
	cd $(ROOT)/sed-service/e2e_tests && \
	  (test -d .venv || python3 -m venv .venv) && . .venv/bin/activate && \
	  pip install -q -r requirements.txt && \
	  RAG_FIXTURES_ENABLED=true \
	  pytest -q tests/test_rag_fixtures.py tests/test_rag_fixtures_warehouse.py

test-e2e-rag: test-up seed-rag test-rag-fixtures
