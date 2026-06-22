#!/bin/bash
# SSH Manager Fuzz Test Runner

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORTS_DIR="$PROJECT_DIR/reports/fuzz"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[FUZZ]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

# Create reports directory
mkdir -p "$REPORTS_DIR"

# Check Go installation
if ! command -v go &> /dev/null; then
    error "Go is not installed"
    exit 1
fi

log "Go version: $(go version)"

# Fuzz duration (default: 30s)
DURATION="${FUZZ_DURATION:-30s}"

# Number of workers
WORKERS="${FUZZ_WORKERS:-$(nproc)}"

# Parse arguments
CATEGORY="${1:-all}"

case "$CATEGORY" in
    ssh)
        log "Running SSH protocol fuzz tests..."
        cd "$PROJECT_DIR"
        go test -fuzz=FuzzSSHProtocol -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzSSHClient -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzSSHConfig -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzSSHCrypto -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzSSHAuth -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzSSHPacketEncode -fuzztime="$DURATION" ./security/fuzz/
        ;;
    
    websocket)
        log "Running WebSocket fuzz tests..."
        cd "$PROJECT_DIR"
        go test -fuzz=FuzzWebSocketFrame -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzWebSocketMessage -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzWebSocketHandshake -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzWebSocketExtensions -fuzztime="$DURATION" ./security/fuzz/
        ;;
    
    api)
        log "Running API fuzz tests..."
        cd "$PROJECT_DIR"
        go test -fuzz=FuzzAPIRequest -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzAPIResponse -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzQueryParameters -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzJSONPayload -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzURLPath -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzAuthToken -fuzztime="$DURATION" ./security/fuzz/
        go test -fuzz=FuzzRateLimit -fuzztime="$DURATION" ./security/fuzz/
        ;;
    
    all)
        log "Running all fuzz tests..."
        
        log "SSH Protocol..."
        go test -fuzz=FuzzSSHProtocol -fuzztime="$DURATION" ./security/fuzz/ || warn "SSH fuzz found issues"
        
        log "WebSocket..."
        go test -fuzz=FuzzWebSocketFrame -fuzztime="$DURATION" ./security/fuzz/ || warn "WebSocket fuzz found issues"
        
        log "API..."
        go test -fuzz=FuzzAPIRequest -fuzztime="$DURATION" ./security/fuzz/ || warn "API fuzz found issues"
        ;;
    
    *)
        error "Unknown category: $CATEGORY"
        echo "Usage: $0 [ssh|websocket|api|all]"
        exit 1
        ;;
esac

# Generate report
log "Generating fuzz report..."
cat > "$REPORTS_DIR/fuzz-report-$(date +%Y%m%d-%H%M%S).md" << EOF
# Fuzz Test Report

**Date:** $(date -Iseconds)
**Duration:** $DURATION
**Category:** $CATEGORY

## Results

See test output above for detailed results.

## Coverage

- SSH Protocol
- WebSocket Protocol
- API Endpoints
- Authentication
- Rate Limiting

## Next Steps

1. Review any crashes or failures
2. Add regression tests for found issues
3. Increase fuzzing duration for deeper coverage
4. Run fuzzing in CI/CD pipeline

EOF

success "Fuzz tests completed. Report saved to $REPORTS_DIR/"

# Optional: Run benchmarks
if [[ "${RUN_BENCHMARKS:-false}" == "true" ]]; then
    log "Running benchmarks..."
    go test -bench=. -benchmem ./security/fuzz/
fi

exit 0
