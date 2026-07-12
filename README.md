# own_test

Small Go service that serves:

- a static home page at `/`
- a GraphQL API at `/graphql`
- GraphiQL at `/graphiql`
- a websocket echo endpoint at `/echo`

## Requirements

- Go 1.25.12 or newer
- Docker Compose, only if you want to run the full container stack locally

## Setup

Install Go dependencies:

```sh
cd go-simple
go mod download
```

## Run

Run the Go service directly:

```sh
cd go-simple
go run .
```

The service listens on `http://localhost:8000`.

Example GraphQL query:

```sh
curl -X POST http://localhost:8000/graphql \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ jobs { id position company skillsRequired } }"}'
```

Run the container stack when Docker Compose is available:

```sh
docker compose up --build
```

Nginx exposes the app on `http://localhost`.

## Test

Run the test suite:

```sh
cd go-simple
go test ./...
```

Run tests with coverage and enforce the project threshold:

```sh
cd go-simple
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -func=coverage.out | awk '/^total:/ { if ($3+0 < 93) exit 1 }'
```

## Environment

The Go application does not require environment variables. It reads `data.json`
and `hello.html` from its working directory.

The Docker Compose PostgreSQL service reads development defaults from
`conf/postgres/conf`. Replace those values with deployment-specific credentials
outside local development.
