# Contributing

## git

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) when making commits.

## Setup

### 🚧 Building and running

- Copy example.env as .env and change the values according to your needs.
- Build `go build`
- (Optional) Manually initialize development database: `goose -dir pkg/db/migrations sqlite3 ./data/mangatsu.sqlite up`
- Run `backend` (`backend.exe` on Windows)

### 💾 Database migrations

- Automatically run when the server is launched. Can be disabled by setting `MTSU_DB_MIGRATIONS=false` in `.env`.
- Manually: `goose -dir pkg/db/migrations sqlite3 ./PATH/TO/mangatsu.sqlite <up|down|status>`
    - To use goose as a tool: `go install github.com/pressly/goose/v3/cmd/goose@latest`
- Automatic models and types: `jet -dsn="file:///full/path/to/data.sqlite" -path=types` based on the db schema
    - To use install jet as a tool: `go install github.com/go-jet/jet/v2/cmd/jet@latest`

### 🔬 Testing

- Test: `go test ./... -v -coverprofile "coverage.out"`
- Show coverage report: `go tool cover -html "coverage.out"`

### 📝 Generating docs

- Run `godoc -http=localhost:8080`
- Go to `http://localhost:8080/pkg/#thirdparty`

## Requirements

- Go 1.21+
- SQLite3
- Docker (optional)
