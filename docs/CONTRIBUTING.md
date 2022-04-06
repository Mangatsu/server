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
- Manual migrations
  - SQLite3: `goose -dir pkg/db/migrations/sqlite sqlite3 ./PATH/TO/mangatsu.sqlite <up|down|status>`
  - PostgreSQL: `goose -dir pkg/db/migrations/psql postgres "user=USER password=PASSWORD dbname=DATABASE_NAME sslmode=<disable|enabled>" <up|down|status>`
  - MySQL/MariaDB: `goose -dir pkg/db/migrations/mysql mysql "USER:PASSWORD@/DATABASE_NAME?parseTime=true&multiStatements=true" <up|down|status>`
  - More examples on [Goose's GitHub](https://github.com/pressly/goose#usage)
- Automatic models and types: `jet -dsn="file:///full/path/to/data.sqlite" -path=types` based on the db schema

### 🔬 Testing
- Test: `go test ./... -v  -coverprofile "coverage.out"`
- Show coverage report: `go tool cover -html "coverage.out"`

### 📝 Generating docs
- Run `godoc -http=localhost:8080`
- Go to `http://localhost:8080/pkg/#thirdparty`

## Requirements
- Go 1.18+
- SQLite3
- Docker (optional)
