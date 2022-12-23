SQLBOILER_VERSION=v4.14.0
GOOSE_VERSION=v3.7.0
PG_WAIT_VERSION=v0.1.3
GOLANGCI_LINT_VERSION=v1.50.1

LOCAL_BIN=$(CURDIR)/bin
MAKE_PATH=$(LOCAL_BIN):/bin:/usr/bin:/usr/local/bin

SQLBOILER_BIN=$(LOCAL_BIN)/sqlboiler
SQLBOILER_DRIVER_BIN=$(LOCAL_BIN)/sqlboiler-psql
GOOSE_BIN=$(LOCAL_BIN)/goose
PG_WAIT_BIN=$(LOCAL_BIN)/pg-wait
GOLANGCI_LINT_BIN=$(LOCAL_BIN)/golangci-lint

POSTGRES_DSN=postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable

.PHONY: all
all: go-install

.PHONY: go-install
go-install: sqlboiler-install migrate-install

.PHONY: sqlboiler-install
sqlboiler-install:
	go run ./cmd/go-install github.com/volatiletech/sqlboiler/v4@$(SQLBOILER_VERSION) $(SQLBOILER_BIN) && \
	go run ./cmd/go-install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@$(SQLBOILER_VERSION) $(SQLBOILER_DRIVER_BIN)

.PHONY: goose-install
goose-install:
	go run ./cmd/go-install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) $(GOOSE_BIN)

.PHONY: pg-wait-install
pg-wait-install:
	go run ./cmd/go-install github.com/partyzanex/pg-wait/cmd/pg-wait@$(PG_WAIT_VERSION) $(PG_WAIT_BIN)

.PHONY: golangci-lint-install
golangci-lint-install:
	go run ./cmd/go-install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) $(GOLANGCI_LINT_BIN)

.PHONY: local-db-up
local-db-up: local-db-down
	docker-compose up -d postgresql

.PHONY: local-db-down
local-db-down:
	docker-compose stop postgresql

.PHONY: migration-up
migration-up: pg-wait-install goose-install local-db-up
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	$(GOOSE_BIN) -dir $(CURDIR)/db/migrations/postgres -table goadmin_migrations postgres $(POSTGRES_DSN) up

.PHONY: migration-down
migration-down: pg-wait-install goose-install local-db-up
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	$(GOOSE_BIN) -dir $(CURDIR)/db/migrations/postgres -table goadmin_migrations postgres $(POSTGRES_DSN) down

.PHONY: sqlboiler-gen
sqlboiler-gen: local-db-up migration-up
	cd $(CURDIR)/db && PATH=$(MAKE_PATH) $(SQLBOILER_BIN) psql

.PHONY: create-default-user
create-default-user: migration-up
	go run $(CURDIR)/cmd/goadmin-users --dsn=$(POSTGRES_DSN) \
	--login="admin@example.com" --password="123456" --name="Admin" --role="owner"

.PHONY: run-example
run-example: create-default-user
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	cd $(CURDIR)/example && \
	PG_DSN=$(POSTGRES_DSN) go run main.go

.PHONY: lint
lint:
	$(GOLANGCI_LINT_BIN) run

.PHONY: test
test: migration-up
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	TEST_PG=$(POSTGRES_DSN) go test -race -v -count=1 -tags 'integration' ./...

