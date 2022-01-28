# Contributing 

### 🚧 Building and running
- Copy example.env as .env and change the values according to your needs.
- Build `go build`
- (Optional) Manually initialize development database: `goose -dir db/migrations sqlite3 ./data/data.sqlite up`
- Run `backend` (`backend.exe` on Windows)

### 💾 Database migrations
- Migrations: `goose -dir db/migrations sqlite3 ./data.sqlite up`
- Automatic models and types: `jet -dsn="file:///full/path/to/data.sqlite" -path=types` based on the db schema

### 🔬 Testing
- Test: `go test ./... -v  -coverprofile "coverage.out"` to test
- Show coverage report: `go tool cover -html "coverage.out"`

### 📝 Generating docs
- Run `godoc -http=localhost:8080`
- Go to `http://localhost:8080/pkg/#thirdparty`

## Requirements
- Go 1.7+
- SQLite3
- Docker (optional)
