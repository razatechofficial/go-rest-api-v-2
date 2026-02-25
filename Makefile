# ── variables ─────────────────────────────────────────────────────────────
BINARY_NAME = api
MIGRATE_CMD = go run cmd/migrate/main.go

# ── migration commands ─────────────────────────────────────────────────────
.PHONY: migrate-up migrate-down migrate-down-all migrate-version migrate-create

## migrate-up: apply all pending migrations
migrate-up:
	$(MIGRATE_CMD) up

## migrate-down: rollback last migration
migrate-down:
	$(MIGRATE_CMD) down

## migrate-down-all: rollback all migrations (dev only)
migrate-down-all:
	$(MIGRATE_CMD) down-all

## migrate-version: show current migration version
migrate-version:
	$(MIGRATE_CMD) version

## migrate-create name=<name>: create a new migration file pair
## Usage: make migrate-create name=create_orders_table
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=<migration_name>"; \
		exit 1; \
	fi
	@NEXT=$$(ls migrations/*.sql 2>/dev/null | \
		grep -oE '^migrations/[0-9]+' | \
		sort -n | tail -1 | \
		grep -oE '[0-9]+' || echo "0"); \
	PADDED=$$(printf "%06d" $$((NEXT + 1))); \
	touch migrations/$${PADDED}_$(name).up.sql; \
	touch migrations/$${PADDED}_$(name).down.sql; \
	echo "✓ Created:"; \
	echo "  migrations/$${PADDED}_$(name).up.sql"; \
	echo "  migrations/$${PADDED}_$(name).down.sql"

# ── run commands ───────────────────────────────────────────────────────────
.PHONY: run build fresh help

## run: start the api server
run:
	go run cmd/api/main.go

## build: build the api binary
build:
	go build -o bin/$(BINARY_NAME) cmd/api/main.go

## fresh: reset db and re-run all migrations (dev only)
fresh:
	$(MIGRATE_CMD) down-all
	$(MIGRATE_CMD) up

## tidy: clean up go.mod and go.sum
tidy:
	go mod tidy

## help: list all available commands
help:
	@grep -E '^## ' Makefile | sed 's/## //'''