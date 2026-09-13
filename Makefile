.PHONY: test test-all generate up up-db wait-db down

generate:
	docker compose run --rm sqlc

test:
	@trap '$(MAKE) down' EXIT; \
	$(MAKE) up-db; \
	$(MAKE) generate; \
	$(MAKE) wait-db; \
	go test ./db/... -v

test-all: test

up:
	docker compose up -d

up-db:
	docker compose up -d database

wait-db:
	@until docker compose exec -T database pg_isready -U postgres -d prueba >/dev/null 2>&1; do \
		sleep 1; \
	done

down:
	docker compose down -v
