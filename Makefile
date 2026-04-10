LOCAL_BIN=$(CURDIR)/bin
PG_DSN ?= postgres://goadmin:goadmin@localhost:5432/goadmin?sslmode=disable
JWT_SECRET ?= dev-secret-change-me

.PHONY: all
all: lint test

.PHONY: tools
tools:
	go install github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: env-up
env-up:
	docker compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker compose exec postgres pg_isready -U goadmin > /dev/null 2>&1; do sleep 1; done
	@echo "PostgreSQL is ready."

.PHONY: env-down
env-down:
	docker compose down

.PHONY: migration-up
migration-up:
	go run $(CURDIR)/cmd/goadmin-users migrate --dsn=$(PG_DSN) --direction=up

.PHONY: create-default-user
create-default-user:
	go run $(CURDIR)/cmd/goadmin-users create-user \
		--dsn=$(PG_DSN) \
		--login="admin@example.com" --password="Admin123" --name="Admin" --role="owner" # dev only — use GOADMIN_PASSWORD in production

.PHONY: run-example
run-example: env-up
	go run $(CURDIR)/cmd/goadmin-users migrate --dsn=$(PG_DSN) --direction=up
	-go run $(CURDIR)/cmd/goadmin-users create-user \
		--dsn=$(PG_DSN) \
		--login="admin@example.com" --password="Admin1234" --name="Admin" --role="root"
	cd $(CURDIR)/example && \
	PG_DSN=$(PG_DSN) JWT_SECRET=dev-secret-change-me go run main.go

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run -c .golangci.yml

.PHONY: test
test:
	go test -race -count=1 -tags=integration -v ./...

.PHONY: cover
cover:
	go test -race -count=1 -tags=integration -coverprofile=coverage.out ./... && \
	go tool cover -html=coverage.out -o coverage.html
