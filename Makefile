-include .env
COMPOSE_PROJECT_NAME ?= bookfly-tp5
export COMPOSE_PROJECT_NAME POSTGRES_DB POSTGRES_USER POSTGRES_PASSWORD APP_PORT POSTGRES_PORT

.PHONY: generate up up-api up-db run-app test test-hurl test-all down stop

generate:
	docker compose run --rm sqlc
	docker compose run --rm templ

up: up-api

up-api: generate
	docker compose up -d --build --wait --wait-timeout 120 api

up-db:
	docker compose up -d --wait --wait-timeout 60 database

# Arranque de entrega: prepara, ejecuta y limpia al salir (también con Ctrl+C).
run-app:
	bash scripts/run-app.sh

test:
	@set -e; \
	trap '$(MAKE) down' EXIT; \
	$(MAKE) down; \
	$(MAKE) up-db; \
	$(MAKE) generate; \
	docker compose run --rm --no-deps go-tests

test-hurl:
	@set -e; \
	trap '$(MAKE) down' EXIT; \
	$(MAKE) down; \
	$(MAKE) up-api; \
	docker compose run --rm --no-deps hurl

test-all:
	$(MAKE) test
	$(MAKE) test-hurl

stop:
	docker compose --profile test down --remove-orphans

down:
	docker compose --profile test down -v --remove-orphans
