.PHONY: docker-up
docker-up:
	$(info #up docker...)
	docker compose -p "$(CI_PROJECT_ID)" -f docker-compose.yml up --build -d

.PHONY: docker-down
docker-down:
	$(info #down docker...)
	docker compose -p "$(CI_PROJECT_ID)" -f docker-compose.yml down

export PG_DB_DSN=host=localhost port=6432 user=test password=password database=dating

.PHONY: migrate-up
migrate-up:
	goose -dir migrations/ -allow-missing postgres "$(PG_DB_DSN)" up

.PHONY: migrate-down
migrate-down:
	goose -dir migrations/ -allow-missing postgres "$(PG_DB_DSN)" down

.PHONY: migrate-reset
migrate-reset:
	goose -dir migrations/ -allow-missing postgres "$(PG_DB_DSN)" reset

.PHONY: migrate-generate
migrate-generate:
	goose -dir migrations/ create "$(name)" sql

.PHONY: migrate-status
migrate-status:
	goose -dir migrations/ -allow-missing postgres "$(PG_DB_DSN)" status
