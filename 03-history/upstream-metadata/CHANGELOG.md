# Changelog

All notable changes to SSH Manager will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- [2026-05-28] Initial project scaffolding with monorepo structure
- [2026-05-28] Go backend with SSH/RDP/VNC gateway, auth, vault, audit
- [2026-05-28] Next.js 16 frontend with React 19 + TypeScript
- [2026-05-28] Tauri v2 desktop app (Rust + WebView)
- [2026-05-28] Flutter mobile app scaffold
- [2026-05-28] Rust SSH core library with FFI bindings
- [2026-05-28] Kubernetes manifests (10 resources)
- [2026-05-28] Terraform modules (5 environments)
- [2026-05-28] Docker multi-stage builds with distroless runtime
- [2026-05-28] CI/CD pipeline (GitHub Actions) with SAST, SCA, signing
- [2026-05-28] Integration test suite (Playwright + Go tests)
- [2026-05-28] Security documentation (threat model, secure coding guide)
- [2026-05-28] Design system, wireframes, interactions
- [2026-05-28] Monitoring stack (Prometheus, Grafana, Loki)

### Security
- [2026-05-28] STRIDE threat model analysis
- [2026-05-28] Secure coding guidelines
- [2026-05-28] Penetration test plan
- [2026-05-28] Vulnerability findings tracker
- [2026-05-28] Security test cases (30+)

### Infrastructure
- [2026-05-28] Docker Compose development environment
- [2026-05-28] Kubernetes production manifests
- [2026-05-28] Terraform infrastructure modules
- [2026-05-28] Multi-stage Docker builds (builder → scanner → runtime)
- [2026-05-28] Image signing with Cosign
- [2026-05-28] SBOM generation

## [0.1.0] - 2026-05-28

### Added
- Initial release
- Monorepo structure: apps/ (api, web, desktop, mobile) + packages/ (ssh-core, ui, types, config)
- Full documentation suite: pm/, design/, dev/, qa/, devops/, architecture/, research/
- 18 documentation files, 226+ source files
- Dev, QA, DevOps agent status tracking

### Notes
- Phase 0 (Foundation) completed
- Sprint 0 goal achieved: Project scaffolding, security baseline, CI/CD skeleton
- Integration tests ready but need live backend verification
