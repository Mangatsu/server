# Contributing 

### 🚧 Building
- Build `go build`
- Initialize development database: `goose -dir db/migrations sqlite3 ./data/data.sqlite up`
- Run `backend` on Linux or `backend.exe` on Windows

### 💾 Database migrations
```
goose -dir db/migrations sqlite3 ./data.sqlite up
jet -dsn="file:///full/path/to/data.sqlite" -path=types  
```

### 🔬 Testing
- Test: `go test ./... -v  -coverprofile "coverage.out"` to test
- Show coverage report: `go tool cover -html "coverage.out"`

### 📝 Generating docs
- Coming soon

## Requirements
- Go 1.7+
- SQLite3
- Docker (optional)
