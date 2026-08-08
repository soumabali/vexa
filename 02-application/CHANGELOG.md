# Changelog

All notable changes to vexa are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] — 2026-08-08

### Added

#### Security Foundation (P1)
- Argon2id password hashing (OWASP recommendation)
- JWT access + refresh token authentication
- Encrypted TOTP/WebAuthn metadata storage
- Redis-backed MFA session management
- CORS origin enforcement (`ALLOWED_ORIGINS`)
- Comprehensive audit logging

#### Open Source Cleanup (P2)
- Rebrand to `github.com/soumabali/vexa`
- Consistent application metadata across web, desktop, and mobile
- Root orchestrator with subtree push to application repo
- Removed stale K8s/Terraform documentation

#### Documentation (P3)
- Contributor onboarding guide
- Self-hosted quick start with Docker Compose
- Architecture overview
- Development setup instructions
- Security disclosure policy
- MIT license

#### Production Hardening (P4)
- Log rotation for API and web containers
- PostgreSQL automated backups (daily, with retention)
- WireGuard key rotation workflow
- Health check endpoint expansion
- Container-level monitoring

#### TOTP MFA (P5)
- Server-side TOTP secret generation
- TOTP verification during login flow
- One-time recovery codes with regenerate support
- UI for enable/disable TOTP in settings
- Backup codes display with copy-to-clipboard

#### Credential Sharing / Teams (P6)
- Team and organization concept
- Vault credential access control (read-only vs full access)
- Audit trail for credential access
- Team membership management
- Migration framework for teams tables

#### WireGuard Live Stats (P7)
- Endpoint `/tunnels/:id/stats` with real-time data
- UI displaying handshake and transfer statistics
- Auto-refresh with Page Visibility API

#### SSH Terminal E2E Tests (P8)
- E2E specs that actually type SSH commands
- Mock SSH server for deterministic testing
- Output verification from terminal sessions
- Integration tests using `golang.org/x/crypto/ssh`

#### Host Detail Enhancement (P9)
- Activity log per host
- Connection history graph
- Health check from backend via SSH
- Tags and labels for host organization
- HostStatsCard and HostMetadataPanel components

#### Settings Polish (P10)
- Active sessions list with remote revocation
- Session kill/reject endpoint
- WebAuthn passkey management (rename, delete)
- Recovery codes regeneration dialog
- Backup codes one-time display UI

### Security

- Argon2id password hashing (OWASP recommendation)
- AES-256-GCM credential encryption at rest with unlockable master vault
- Rate limiting and brute-force protection
- WebAuthn/FIDO2 passkey support
- TOTP two-factor authentication with backup codes
- Comprehensive audit logging for all critical actions
- CORS origin enforcement
- Pre-commit SAST (semgrep), secret scanning (trufflehog), dependency scanning (Snyk/CodeQL)

---

[1.0.0]: https://github.com/soumabali/vexa/releases/tag/v1.0.0
