# Getting Started — vexa

Welcome! This guide will help you set up the development environment for vexa.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Git | ≥ 2.30 | Version control |
| Go | ≥ 1.21 | Backend services |
| Node.js | ≥ 18.x | Frontend / tooling |
| Rust | ≥ 1.75 | Native extensions |
| Dart | ≥ 3.0 | Flutter desktop app |
| Python | ≥ 3.9 | Scripts / pre-commit |

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/soumabali/vexa.git
cd vexa

# 2. Run the automated setup
./scripts/setup-hooks.sh

# 3. Verify everything works
pre-commit run --all-files
```

---

## Pre-commit Security Hooks

We use **pre-commit** hooks to automatically scan code for security issues before every commit. These hooks are **required** for all contributors.

### What Gets Scanned

| Hook | Tool | Languages | Purpose |
|------|------|-----------|---------|
| Secret Detection | gitleaks | All | Hardcoded secrets, API keys, tokens |
| Secret Detection | trufflehog | All | Verified secrets in git history |
| SAST | semgrep | All | Static analysis security testing |
| Go Security | gosec | Go | Go-specific vulnerabilities (CWE, OWASP) |
| JS/TS Security | eslint-plugin-security | JS/TS/TSX | Dangerous patterns (eval, innerHTML, etc.) |
| Custom Checker | check-suspicious.sh | All | Dangerous code patterns review |
| Rust Format | rustfmt | Rust | Code formatting |
| Dart Analysis | dart analyze | Dart | Static analysis & type checking |

### Performance Targets

- **Total runtime:** < 30 seconds for typical commits
- **Large files:** Automatically skipped (> 1MB)
- **Network:** Not required during commit (all tools cached locally)

### Setup Commands

```bash
# Install hooks (run once)
pre-commit install

# Run manually on all files
pre-commit run --all-files

# Run only secret scanners
pre-commit run gitleaks trufflehog

# Run only Go hooks
pre-commit run gofmt goimports gosec

# Update hooks to latest versions
pre-commit autoupdate

# Skip hooks in emergency (avoid if possible)
git commit --no-verify -m "your message"
```

### Tool Installation (Manual)

If the automated script doesn't work for your system:

```bash
# Pre-commit framework
pip install pre-commit        # or: pipx install pre-commit
brew install pre-commit       # macOS

# Secret scanning
brew install gitleaks
go install github.com/gitleaks/gitleaks/v8@latest

brew install trufflesecurity/trufflehog/trufflehog
go install github.com/trufflesecurity/trufflehog/v3@latest

# SAST
pip install semgrep           # or: brew install semgrep

# Go security
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Go formatting
go install golang.org/x/tools/cmd/goimports@latest

# JS/TS security
npm install -g eslint eslint-plugin-security
npm install -g @typescript-eslint/parser @typescript-eslint/eslint-plugin

# Rust
rustup component add rustfmt

# Dart (included with Flutter)
flutter doctor
```

### Troubleshooting

| Problem | Solution |
|---------|----------|
| Hook takes too long | Run `pre-commit run <hook-id>` individually to identify slow hook |
| gitleaks false positive | Add to `.gitleaksignore` or adjust `.gitleaks.toml` |
| semgrep timeout | Reduce `--timeout` in `.pre-commit-config.yaml` |
| eslint not found | Run `npm install` in project root or install globally |
| "command not found" after install | Add Go bin to PATH: `export PATH="$PATH:$(go env GOPATH)/bin"` |
| Skip large files | Already configured: `--maxkb=1024` and `--max-target-bytes=1048576` |

### Configuration Files

- `.pre-commit-config.yaml` — Main hook configuration
- `apps/web/.eslintrc.json` — ESLint config with security plugins
- `apps/web/.eslintignore` — ESLint ignore patterns
- `apps/api/.golangci.yml` — golangci-lint config with security linters
- `apps/api/.golangci-ignore` — golangci-lint ignore patterns
- `apps/desktop/clippy.toml` — Clippy security configuration
- `apps/desktop/scripts/clippy-check.sh` — Clippy check script
- `apps/mobile/analysis_options.yaml` — Flutter analysis configuration
- `scripts/setup-hooks.sh` — Automated setup script
- `scripts/check-suspicious.sh` — Custom pattern checker
- `.status/precommit.status` — Installation status tracker

---

## Lint Commands

This project uses security-focused linting for all tech stacks.

### Web (ESLint)
```bash
cd apps/web

# Install security plugins
npm install eslint-plugin-security eslint-plugin-no-secrets eslint-plugin-no-unsanitized eslint-plugin-scanjs-rules eslint-plugin-jsx-a11y

# Check
npm run lint

# Auto-fix
npm run lint:fix

# CI mode (strict, no warnings)
npm run lint:ci

# Generate JSON report
npm run lint:report
```

### API (golangci-lint)
```bash
cd apps/api

# Check
make lint

# Auto-fix
make lint-fix

# Fast mode (quick feedback)
make lint-fast

# CI mode
make lint-ci

# Only new changes
make lint-new

# Diff against main branch
make lint-diff
```

### Desktop (Clippy)
```bash
cd apps/desktop

# Check
./scripts/clippy-check.sh

# Auto-fix
./scripts/clippy-check.sh --fix

# CI mode (strict)
./scripts/clippy-check.sh --ci

# All targets with warnings as errors
./scripts/clippy-check.sh --all-targets --deny-warnings

# Help
./scripts/clippy-check.sh --help
```

### Mobile (Flutter)
```bash
cd apps/mobile

# Analyze
flutter analyze

# Treat warnings as errors
flutter analyze --fatal-warnings

# Treat infos as errors
flutter analyze --fatal-infos
```

---

## Development Workflow

```bash
# Before committing, always run:
pre-commit run

# Or let it run automatically on git commit

# If hooks fail:
# 1. Review the output
# 2. Fix the issues
# 3. Stage your changes: git add .
# 4. Commit again: git commit -m "..."
```

---

## Project Structure

```
vexa/
├── .pre-commit-config.yaml    # Pre-commit hooks
├── .eslintrc.security.js      # Security linting rules
├── scripts/
│   ├── setup-hooks.sh         # Hook installer
│   └── check-suspicious.sh    # Custom security checker
├── docs/
│   └── dev/
│       └── getting-started.md # This file
├── .status/
│   └── precommit.status       # Hook installation status
├── src/                       # Source code
└── ...
```

---

### Mobile (Flutter)
```bash
cd apps/mobile

# Analyze
flutter analyze

# Treat warnings as errors
flutter analyze --fatal-warnings

# Treat infos as errors
flutter analyze --fatal-infos
```

## Need Help?

- Read the full [security guide](security.md)
- Check [troubleshooting](#troubleshooting) above
- Open an issue with the `setup` label
