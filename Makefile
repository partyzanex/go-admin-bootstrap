GOOSE_VERSION=v3.27.0
PG_WAIT_VERSION=v0.1.3

LOCAL_BIN=$(CURDIR)/bin
MAKE_PATH=$(LOCAL_BIN):/bin:/usr/bin:/usr/local/bin

GOOSE_BIN=$(LOCAL_BIN)/goose
PG_WAIT_BIN=$(LOCAL_BIN)/pg-wait

POSTGRES_DSN=postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable

.PHONY: all
all: tools

.PHONY: tools
tools: goose-install pg-wait-install

.PHONY: goose-install
goose-install:
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

.PHONY: pg-wait-install
pg-wait-install:
	GOBIN=$(LOCAL_BIN) go install github.com/partyzanex/pg-wait/cmd/pg-wait@$(PG_WAIT_VERSION)

.PHONY: local-db-up
local-db-up: local-db-down
	docker compose up -d postgresql

.PHONY: local-db-down
local-db-down:
	docker compose stop postgresql

.PHONY: migration-up
migration-up: pg-wait-install goose-install local-db-up
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	$(GOOSE_BIN) -dir $(CURDIR)/db/migrations/postgres -table goadmin_migrations postgres $(POSTGRES_DSN) up

.PHONY: migration-down
migration-down: pg-wait-install goose-install local-db-up
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	$(GOOSE_BIN) -dir $(CURDIR)/db/migrations/postgres -table goadmin_migrations postgres $(POSTGRES_DSN) down

.PHONY: create-default-user
create-default-user: migration-up
	go run $(CURDIR)/cmd/goadmin-users create-user \
		--dsn=$(POSTGRES_DSN) \
		--login="admin@example.com" --password="Admin123" --name="Admin" --role="owner" # dev only — use GOADMIN_PASSWORD in production

.PHONY: run-example
run-example: create-default-user
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	cd $(CURDIR)/example && \
	PG_DSN=$(POSTGRES_DSN) JWT_SECRET=dev-secret-change-me go run main.go

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run -c .golangci.yml

.PHONY: test
test: migration-up
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	TEST_PG=$(POSTGRES_DSN) go test -race -v -count=1 -tags 'integration' ./...

.PHONY: cover
cover: migration-up
	$(PG_WAIT_BIN) -d $(POSTGRES_DSN) && \
	TEST_PG=$(POSTGRES_DSN) go test -race -count=1 -tags 'integration' -coverprofile=coverage.out ./... && \
	go tool cover -html=coverage.out -o coverage.html
