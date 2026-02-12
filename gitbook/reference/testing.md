# 🧪 Testing

Aureo VPN uses Go's built-in testing framework with race detection and coverage reporting. Tests are organized into unit and integration directories.

---

## Test Structure

```
aureo-vpn/
└── tests/
    ├── unit/
    │   └── auth_test.go              # Authentication service unit tests
    │
    └── integration/
        ├── vpn_flow_test.go          # End-to-end VPN connection flow
        └── earnings_flow_test.go     # Operator earnings pipeline
```

Additionally, security package tests live alongside their source files:

```
aureo-vpn/pkg/security/
├── crypto_test.go                    # Cryptographic primitives tests
├── jwt_test.go                       # JWT service tests
├── netshield_test.go                 # Network security tests
├── privacy_test.go                   # Privacy filter tests
└── validation_test.go               # Input validation tests
```

---

## Commands

### Backend (aureo-vpn)

```bash
cd aureo-vpn

# Run all tests with race detection and coverage
make test
# Equivalent to: go test -v -race -coverprofile=coverage.out ./...

# Run unit tests only
make test-unit
# Equivalent to: go test -v -race ./tests/unit/...

# Run integration tests only
make test-integration
# Equivalent to: go test -v -race ./tests/integration/...

# Run a single test by name
go test -v -race ./tests/unit/ -run TestName

# Generate HTML coverage report
make coverage
# Opens coverage.html in browser

# Run benchmarks
make bench
# Equivalent to: go test -bench=. -benchmem ./...
```

### Linting & Static Analysis

```bash
cd aureo-vpn

# Run golangci-lint
make lint
# Equivalent to: golangci-lint run ./...

# Run go vet
make vet
# Equivalent to: go vet ./...

# Format code
make fmt
# Equivalent to: go fmt ./...

# Security scan
make security-scan
# Equivalent to: gosec ./...
```

### Mobile App (aureo-app)

```bash
cd aureo-app

# Run ESLint
npm run lint
```

---

## Test Database

Tests use **SQLite in-memory** databases to avoid external dependencies:

- The test setup initializes a fresh SQLite database for each test or test suite
- `DB_PATH` environment variable can override the database path for integration tests
- Each test cleans up its data to avoid cross-test contamination
- GORM `AutoMigrate` runs before tests to ensure schema is current

### Example Test Setup Pattern

```go
func TestMain(m *testing.M) {
    // Initialize in-memory SQLite
    database.ConnectSQLite(":memory:")
    database.AutoMigrate()

    // Run tests
    code := m.Run()

    // Cleanup
    database.Close()
    os.Exit(code)
}
```

---

## Test Categories

### Unit Tests (`tests/unit/`)

Fast, isolated tests that mock external dependencies:

- **auth_test.go** -- Tests user registration, login, token generation, token verification, password validation, and error handling without a real database

### Integration Tests (`tests/integration/`)

End-to-end tests that exercise the full stack with a real (SQLite) database:

- **vpn_flow_test.go** -- Tests the complete VPN connection lifecycle: register user, create session, provision on node, transfer traffic, disconnect
- **earnings_flow_test.go** -- Tests the operator earnings pipeline: register operator, create node, serve traffic, accumulate earnings, process payout

### Security Tests (`pkg/security/`)

- **crypto_test.go** -- Tests AES-GCM, ChaCha20-Poly1305, key derivation, password hashing/verification, secure random generation, memory wiping
- **jwt_test.go** -- Tests token generation, validation, expiry, refresh rotation, family revocation, blacklisting
- **privacy_test.go** -- Tests IP anonymization, email redaction, log sanitization, data retention
- **validation_test.go** -- Tests email, username, password, UUID, IP, port, hostname, country code, protocol, wallet address validation, SQL injection detection, XSS detection
- **netshield_test.go** -- Tests network security features

---

## Flags

| Flag | Description |
|---|---|
| `-v` | Verbose output (show individual test names and results) |
| `-race` | Enable Go race detector (catches data races) |
| `-coverprofile=FILE` | Write coverage data to file |
| `-run PATTERN` | Run only tests matching the regex pattern |
| `-bench PATTERN` | Run benchmarks matching the pattern |
| `-benchmem` | Include memory allocation stats in benchmarks |
| `-count N` | Run each test N times (useful for flaky test detection) |
| `-timeout DURATION` | Test timeout (default 10m) |
