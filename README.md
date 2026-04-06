# go-admin-bootstrap

Bootstrap library for building Go admin panels with Echo v4, PostgreSQL (uptrace/bun), and Jet templates.

## Features

- User management (CRUD) with role-based access control (owner, root, user)
- JWT access tokens + refresh tokens in PostgreSQL
- Structured logging via `log/slog`
- CSRF protection on all state-changing routes (POST only; GET requests to destructive endpoints return 405)
- Rate limiting on login endpoint
- Secure cookies (HttpOnly, SameSite=Strict, Secure in production)
- Embedded assets (JS, CSS, views) served from `embed.FS` in production
- Goose migrations with embedded SQL
- CLI adapters for cobra and urfave/cli v3
- Graceful shutdown, health check endpoint
- Password hashing with argon2id (bcrypt backward-compatible)

## Quick Start

```go
package main

import (
    "context"
    "database/sql"
    "log/slog"
    "os"
    "os/signal"

    "github.com/uptrace/bun"
    "github.com/uptrace/bun/dialect/pgdialect"
    "github.com/uptrace/bun/driver/pgdriver"

    goadmin "github.com/partyzanex/go-admin-bootstrap"
    "github.com/partyzanex/go-admin-bootstrap/repository/postgres"
    "github.com/partyzanex/go-admin-bootstrap/usecase"
)

func main() {
    sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(os.Getenv("PG_DSN"))))
    db := bun.NewDB(sqldb, pgdialect.New())
    defer db.Close()

    userRepo := postgres.NewUserRepository(db)
    tokenRepo := postgres.NewTokenRepository(db)

    admin, _ := goadmin.New(&goadmin.Config{
        Host:      "localhost",
        Port:      9900,
        BaseURL:   "http://localhost:9900/admin",
        JWTSecret: []byte(os.Getenv("JWT_SECRET")),
        DBConfig:  goadmin.DBConfig{DB: db},
        UserCase:  usecase.NewUserCase(userRepo, tokenRepo),
    })

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    admin.Serve(ctx)
}
```

See `example/main.go` for a complete example with middleware and logging.

## Project Structure

```
├── app.go, config.go, auth.go, ...   Core library
├── usecase/                           Business logic
├── repository/postgres/               Bun-based data layer
├── db/migrations/postgres/            Goose SQL migrations
├── widgets/                           Pagination, breadcrumbs
├── views/                             Embedded Jet templates
├── assets/                            Embedded JS, CSS, favicon
├── pkg/commands/                      Framework-agnostic CLI commands
├── adapters/
│   ├── cobracli/                      Cobra adapter
│   └── urfavecli/                     urfave/cli v3 adapter
├── cmd/
│   ├── goadmin-users/                 User management CLI (cobra)
│   ├── admin-cobra/                   Example CLI (cobra)
│   └── admin-cli/                     Example CLI (urfave/cli v3)
└── example/                           Example web application
```

## CLI Tools

### Create User (cobra)

Password is resolved in priority order: `--password` flag → `GOADMIN_PASSWORD` env → interactive prompt → piped stdin.
Use the flag only in dev/CI environments; prefer the env var or prompt in production.

```bash
# Interactive prompt (production-safe — password is not visible in process list or shell history)
go run ./cmd/goadmin-users create-user \
    --dsn="postgres://user:pass@localhost:5432/db?sslmode=disable" \
    --login="admin@example.com" \
    --name="Admin" \
    --role="owner"

# Environment variable (CI/CD)
GOADMIN_PASSWORD="Admin123" go run ./cmd/goadmin-users create-user \
    --dsn="postgres://user:pass@localhost:5432/db?sslmode=disable" \
    --login="admin@example.com" \
    --name="Admin" \
    --role="owner"

# --password flag (dev shortcut only)
go run ./cmd/goadmin-users create-user \
    --dsn="postgres://user:pass@localhost:5432/db?sslmode=disable" \
    --login="admin@example.com" \
    --password="Admin123" \
    --name="Admin" \
    --role="owner"
```

### Run Migrations

```bash
go run ./cmd/goadmin-users migrate --dsn="..." --direction=up
```

## Development

```bash
make tools           # Install goose and pg-wait
make local-db-up     # Start PostgreSQL via docker compose
make migration-up    # Run migrations
make create-default-user  # Create admin user
make run-example     # Start example application
make test            # Run unit tests
make cover           # Run all tests (requires PostgreSQL) with coverage report
make lint            # Run golangci-lint v2
```

## Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| Host | string | no | "" | Listen host |
| Port | uint16 | yes | - | Listen port |
| BaseURL | string | yes | - | Full base URL |
| JWTSecret | []byte | yes | - | JWT signing secret |
| AccessCookieName | string | no | "auth_token" | Cookie name prefix |
| AccessTokenTTL | Duration | no | 15m | JWT access token TTL |
| RefreshTokenTTL | Duration | no | 30d | Refresh token TTL |
| DevMode | bool | no | false | Development mode |
| DBConfig.DB | *bun.DB | yes | - | Database connection |
| DBConfig.MigrationsTable | string | no | "goadmin_migrations" | Goose table name |
| UserCase | UserUseCase | yes | - | User use case implementation |
| Logger | *slog.Logger | no | slog.Default() | Structured logger |

Options via `goadmin.WithMiddleware(...)` and `goadmin.WithAssets(...)`.

## Security

- **CSRF**: All admin routes use Echo's CSRF middleware. State-changing operations (logout, delete) are POST-only — GET requests return 405. Tokens are validated via cookie + header/form double-submit pattern.
- **Authentication**: JWT access tokens in HttpOnly cookies. Refresh tokens stored in PostgreSQL and invalidated on logout.
- **Rate limiting**: Login endpoint is rate-limited to prevent brute-force attacks.
- **Roles**: `owner` and `root` users can manage other users. `user` role has read-only dashboard access. Users cannot delete themselves.
