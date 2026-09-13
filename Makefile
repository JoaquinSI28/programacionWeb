.PHONY: test-all up test down

test-all:
	@$(MAKE) up
	@$(MAKE) test || ( $(MAKE) down && exit 1 )
	@$(MAKE) down

up:
	docker compose up -d
	sleep 3

test:
	go test ./db/sqlc -v

down:
	docker compose down -v
