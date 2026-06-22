# vexa — Testing Guide

## Test Structure

```
apps/api/tests/
├── auth_test.go              # Authentication tests
├── vault_test.go             # Encryption/vault tests
├── crypto_test.go            # Crypto utilities tests
├── security_regression_test.go  # Security regression
├── api/
│   └── auth_test.go          # API endpoint tests
├── sftp/
│   ├── service_test.go       # SFTP service tests
│   ├── path_test.go          # Path validation tests
│   └── websocket_test.go     # WebSocket tests
├── tunnel/
│   └── manager_test.go       # Tunnel manager tests
└── team/
    └── service_test.go       # Team service tests
```

## Running Tests

### Go Backend Tests
```bash
cd apps/api

# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/auth/...

# Run with verbose output
go test ./... -v

# Run security tests only
go test ./tests/security_regression_test.go -v
```

### Frontend Tests
```bash
cd apps/web

# Run all tests
pnpm test

# Run in watch mode
pnpm test:watch

# Run with coverage
pnpm test:coverage
```

### Rust Tests
```bash
cd packages/ssh-core

# Run all tests
cargo test

# Run with output
cargo test -- --nocapture
```

## Test Coverage

### Backend Coverage Report
```bash
cd apps/api
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Thresholds
| Package | Target |
|---------|--------|
| Auth | 90% |
| Vault | 95% |
| Gateway | 85% |
| API Handlers | 80% |

## Manual Testing Checklist

### Authentication
- [ ] Register new user
- [ ] Login with email/password
- [ ] MFA setup (TOTP)
- [ ] MFA login flow
- [ ] Password reset
- [ ] Email verification
- [ ] Session expiry
- [ ] Logout

### Hosts Management
- [ ] Add SSH host
- [ ] Add RDP host
- [ ] Add VNC host
- [ ] Edit host details
- [ ] Delete host
- [ ] Test connection
- [ ] Group hosts by tags

### Terminal
- [ ] Open SSH session
- [ ] Execute commands
- [ ] Session recording
- [ ] Multiple tabs
- [ ] Copy/paste
- [ ] Resize terminal

### File Manager (SFTP)
- [ ] Browse directories
- [ ] Upload file
- [ ] Download file
- [ ] Delete file
- [ ] Create directory
- [ ] Drag and drop

### Security
- [ ] Audit log entries
- [ ] Session timeout
- [ ] Rate limiting
- [ ] Password strength
- [ ] Encryption at rest

## Security Testing

### OWASP ZAP
```bash
# Run ZAP baseline scan
docker run -t owasp/zap2docker-stable zap-baseline.py \
  -t http://localhost:8080
```

### Go Vulnerability Scanner
```bash
cd apps/api
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### Dependency Audit
```bash
cd apps/web
pnpm audit

cd apps/api
go list -json -deps | nancy sleuth
```

## E2E Testing (Planned)

### Playwright Setup
```bash
cd apps/web
pnpm exec playwright install
pnpm exec playwright test
```

### E2E Test Scenarios
1. User registration flow
2. Login with MFA
3. Add and connect to host
4. Execute command in terminal
5. Upload/download file
6. Audit log verification

## Performance Testing

### Load Test
```bash
# Using k6
k6 run --vus 100 --duration 30s tests/load/auth.js
```

### Benchmarks
```bash
cd apps/api
go test -bench=. ./internal/crypto/
```

## CI/CD Test Pipeline

```yaml
# .github/workflows/test.yml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.25'
    
    - name: Run Go tests
      run: cd apps/api && go test ./... -race -cover
    
    - name: Setup Node
      uses: actions/setup-node@v4
      with:
        node-version: '22'
    
    - name: Run frontend tests
      run: cd apps/web && pnpm test
```

## Debugging Tests

### Go
```bash
# Run specific test with debug
go test ./internal/auth -run TestLogin -v

# Use delve debugger
dlv test ./internal/auth
```

### Frontend
```bash
# Debug mode
pnpm test --debug

# Specific test file
pnpm test auth.test.tsx
```

## Test Data

### Fixtures
Located in `apps/api/tests/fixtures/`:
- `users.json` — Test user accounts
- `hosts.json` — Test host configurations
- `credentials.json` — Test credentials (encrypted)

## Known Test Issues

| Issue | Status | Workaround |
|-------|--------|------------|
| SFTP tests need running server | Open | Use mock server |
| MFA tests need TOTP sync | Open | Use fixed seed |
| WebSocket tests flaky | Open | Increase timeouts |
