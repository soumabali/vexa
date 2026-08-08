# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.0.x   | ✅ Active support |
| < 1.0   | ❌ Unsupported |

## Reporting a Vulnerability

**Please do NOT open public issues for security vulnerabilities.**

Report vulnerabilities via email: **sudharmika@gmail.com**

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Impact assessment (what an attacker could do)
- Suggested fix (if you have one)
- Affected version(s)

### Response Timeline

| Phase | Target |
|-------|--------|
| Acknowledgment | Within 48 hours |
| Triage & confirmation | Within 5 business days |
| Fix published | Within 7 days of confirmation |
| CVE requested | If applicable |

We will keep you informed throughout the process and credit you in the release notes (unless you prefer to remain anonymous).

## Security Features

vexa includes defense-in-depth:

- **Authentication:** Argon2id password hashing (OWASP recommendation), JWT access + refresh tokens, WebAuthn/FIDO2 passkey support, TOTP two-factor authentication
- **Encryption:** AES-256-GCM credential encryption at rest with unlockable master vault
- **Network:** Rate limiting, brute-force protection, CORS origin enforcement, TLS support for production
- **Audit:** Comprehensive audit logging for all critical actions
- **Pipeline:** Pre-commit SAST (semgrep), secret scanning (trufflehog), dependency scanning (Snyk/CodeQL)

## Acknowledgments

We maintain a hall of fame for security researchers who responsibly disclose issues. Thank you for helping keep vexa secure.

---

*Last updated: 2026-08-08*
