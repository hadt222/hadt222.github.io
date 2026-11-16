# REST API Go + PostgreSQL

Task Management API with Go standard `net/http`, `database/sql`, pgx v5, PostgreSQL 17, and tests.

## Requirements

- Go 1.27+
- Docker with Docker Compose

## Install

```sh
go mod tidy
docker compose up -d --wait
docker compose exec -T db psql -U taskuser -d taskdb -f /migrations/001_create_tasks.sql
```

Environment for Linux/macOS Bash:

```sh
export DATABASE_URL='postgres://taskuser:taskpass@localhost:5432/taskdb?sslmode=disable'
go run ./cmd/api
```

Windows PowerShell equivalent:

```powershell
$env:DATABASE_URL='postgres://taskuser:taskpass@localhost:5432/taskdb?sslmode=disable'
go run ./cmd/api
```

Use http://localhost:8080/health to verify server.

## Endpoints

| Method | URL | Effect |
|---|---|---|
| GET | /health | Process health |
| POST | /tasks | Create task |
| GET | /tasks | List tasks |
| GET | /tasks/{id} | Get task |
| PUT | /tasks/{id} | Replace task fields |
| DELETE | /tasks/{id} | Delete task |

## Unit testing

```sh
go test ./...
go test -race ./...
```

## PostgreSQL integration testing

Start dedicated test database (port 5433) and apply migration:

```sh
docker compose --profile test up -d --wait
docker compose exec -T db-test psql -U taskuser -d taskdb_test -f /migrations/001_create_tasks.sql
```

Bash:

```sh
export TEST_DATABASE_URL='postgres://taskuser:taskpass@localhost:5433/taskdb_test?sslmode=disable'
go test -v ./...
```

PowerShell:

```powershell
$env:TEST_DATABASE_URL='postgres://taskuser:taskpass@localhost:5433/taskdb_test?sslmode=disable'
go test -v ./...
```

Integration test skips without TEST_DATABASE_URL; it uses a dedicated database and cleans up the row it creates.

## Example requests

```sh
curl http://localhost:8080/health
curl -i -X POST http://localhost:8080/tasks -H 'Content-Type: application/json' -d '{"title":"Learn Go","description":"Build API","completed":false}'
curl http://localhost:8080/tasks
curl http://localhost:8080/tasks/1
curl -i -X PUT http://localhost:8080/tasks/1 -H 'Content-Type: application/json' -d '{"title":"Learn Go","description":"Done","completed":true}'
curl -i -X DELETE http://localhost:8080/tasks/1
```

These commands use bash-compatible quoting. On PowerShell, prefer `curl.exe` and adapt quoting or use Postman.

## Stop

```sh
docker compose --profile test down
```

To also delete database contents, run `docker compose --profile test down -v` (DESTRUCTIVE).

## Notes

The database URLs and passwords are for *local development only*. Never commit production secrets. This project intentionally omits authentication and external deployment. `/health` only confirms the process is responding, not DB readiness.
