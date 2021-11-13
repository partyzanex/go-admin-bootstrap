SQLBOILER_VERSION=v4.7.1
MIGRATE_VERSION=v4.15.1
PG_WAIT_VERSION=v0.1.3

LOCAL_BIN=$(CURDIR)/bin
MAKE_PATH=$(LOCAL_BIN):/bin:/usr/bin:/usr/local/bin

SQLBOILER_BIN=$(LOCAL_BIN)/sqlboiler
SQLBOILER_DRIVER_BIN=$(LOCAL_BIN)/sqlboiler-psql
MIGRATE_BIN=$(LOCAL_BIN)/migrate
PG_WAIT_BIN=$(LOCAL_BIN)/pg-wait

POSTGRES_DSN=postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable

.PHONY: all
all: go-install

.PHONY: go-install
go-install: sqlboiler-install migrate-install

.PHONY: sqlboiler-install
sqlboiler-install:
	go run ./cmd/go-install github.com/volatiletech/sqlboiler/v4@$(SQLBOILER_VERSION) $(SQLBOILER_BIN) && \
	go run ./cmd/go-install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@$(SQLBOILER_VERSION) $(SQLBOILER_DRIVER_BIN)

.PHONY: migrate-install
migrate-install:
	go run ./cmd/go-install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION) $(MIGRATE_BIN)

.PHONY: pg-wait-install
pg-wait-install:
	go run ./cmd/go-install github.com/partyzanex/pg-wait/cmd/pg-wait@$(PG_WAIT_VERSION) $(PG_WAIT_BIN)

.PHONY: local-db-up
local-db-up: local-db-down
	docker-compose up -d postgresql

.PHONY: local-db-down
local-db-down:
	docker-compose stop postgresql

.PHONY: migration-up
migration-up: pg-wait-install migrate-install local-db-up
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	$(MIGRATE_BIN) -source file://$(CURDIR)/db/migrations/postgres -database $(POSTGRES_DSN) up

.PHONY: migration-down
migration-down: pg-wait-install migrate-install local-db-up
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	$(MIGRATE_BIN) -source file://$(CURDIR)/db/migrations/postgres -database $(POSTGRES_DSN) down -all

.PHONY: sqlboiler-gen
sqlboiler-gen: local-db-up migration-up
	cd $(CURDIR)/db && PATH=$(MAKE_PATH) $(SQLBOILER_BIN) psql


