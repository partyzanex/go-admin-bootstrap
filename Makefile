LOCAL_BIN=$(CURDIR)/bin

.PHONY: all
all: lint test

.PHONY: create-default-user
create-default-user:
	go run $(CURDIR)/cmd/goadmin-users create-user \
		--dsn=$(PG_DSN) \
		--login="admin@example.com" --password="Admin123" --name="Admin" --role="owner" # dev only — use GOADMIN_PASSWORD in production

.PHONY: run-example
run-example:
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
