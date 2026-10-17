-include .env
export POSTGRES_DB POSTGRES_USER POSTGRES_PASSWORD

.PHONY: test test-hurl test-all generate up up-api up-db down

generate:
	docker compose run --rm sqlc

test:
	@set -e; \
	trap '$(MAKE) down' EXIT; \
	$(MAKE) down; \
	$(MAKE) up-db; \
	$(MAKE) generate; \
	go test ./db/... -v

test-hurl:
	@set -e; \
	trap '$(MAKE) down' EXIT; \
	$(MAKE) down; \
	$(MAKE) up-db; \
	$(MAKE) up-api; \
	docker compose run --rm --no-deps hurl

test-all:
	$(MAKE) test
	$(MAKE) test-hurl

up: up-api

up-api: generate
	docker compose up -d --build --wait --wait-timeout 90 api

up-db:
	docker compose up -d --wait --wait-timeout 60 database

down:
	docker compose --profile test down -v --remove-orphans
