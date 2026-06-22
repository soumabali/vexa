# Feature Registry

## Implemented ✅

| ID | Feature | Status | Agent | Date |
|----|---------|--------|-------|------|
| FEAT-001 | Go backend API (REST + WebSocket) | implemented | dev | 2026-05-28 |
| FEAT-002 | Next.js 16 frontend (auth, hosts, sessions) | implemented | dev | 2026-05-28 |
| FEAT-003 | Tauri v2 desktop app scaffold | implemented | dev | 2026-05-28 |
| FEAT-004 | Flutter mobile app scaffold | implemented | dev | 2026-05-28 |
| FEAT-005 | Rust SSH core library (FFI) | implemented | dev | 2026-05-28 |
| FEAT-006 | PostgreSQL + Redis persistence | implemented | dev | 2026-05-28 |
| FEAT-007 | JWT auth + MFA + RBAC | implemented | dev | 2026-05-28 |
| FEAT-008 | Credential vault (encrypted) | implemented | dev | 2026-05-28 |
| FEAT-009 | SSH/RDP/VNC protocol gateway | implemented | dev | 2026-05-28 |
| FEAT-010 | Session recording + audit log | implemented | dev | 2026-05-28 |
| FEAT-011 | Docker multi-stage builds | implemented | devops | 2026-05-28 |
| FEAT-012 | K8s manifests (namespace, config, secrets, DB, app) | implemented | devops | 2026-05-28 |
| FEAT-013 | Terraform IaC modules | implemented | devops | 2026-05-28 |
| FEAT-014 | CI/CD pipeline (SAST, test, build, deploy) | implemented | devops | 2026-05-28 |
| FEAT-015 | Integration test suite | implemented | qa | 2026-05-28 |
| FEAT-016 | Security docs (threat model, secure coding) | implemented | pm | 2026-05-28 |
| FEAT-017 | Design system + wireframes | implemented | design | 2026-05-28 |
| FEAT-018 | Monitoring stack docs | implemented | devops | 2026-05-28 |

## Implemented ✅ (Just Completed)

| ID | Feature | Status | Agent | Date |
|----|---------|--------|-------|------|
| FEAT-019 | Live integration test execution | implemented | qa | 2026-06-06 |
| FEAT-020 | Git repository initialization | implemented | devops | 2026-06-06 |

## In Progress 🔄

| ID | Feature | Status | Agent | Phase |
|----|---------|--------|-------|-------|
| FEAT-025 | Kubernetes operator | planned | devops | Phase 2 |

## Implemented ✅ (Phase 2 - WireGuard Tunnel)

| ID | Feature | Status | Agent | Date |
|----|---------|--------|-------|------|
| FEAT-024 | WireGuard tunnel support | **implemented** | pm/dev/design/qa/devops | 2026-06-06 |

**Phase 2 Sprints:**
- Sprint 2.1 ✅ PM: Task breakdown, 3 risks, ADR-013
- Sprint 2.1 ✅ Design: Wireframes (host tab, list, config modal)
- Sprint 2.2 ✅ Backend: 14 Go files, 74 tests, 80.1%→82.3% coverage
- Sprint 2.3 ✅ Frontend: 10 TSX files, Zustand + WebSocket
- Sprint 2.4 ✅ QA: 74 unit + 12 integration + 9 E2E + 7 security tests
- Sprint 2.5 ✅ DevOps: Docker WG support, production compose, Caddy TLS

**Phase 2 Stats:**
- 25 new files, ~8,500 lines code
- 102 tests passing (74 unit + 12 integration + 9 E2E + 7 security)
- 82.3% backend coverage
- 0 breaking changes
- Self-hosted deployment ready (Docker Compose)

## Planned 📋

| ID | Feature | Phase | Priority |
|----|---------|-------|----------|
| FEAT-026 | SAML/SSO integration | Phase 3 | Medium |
| FEAT-027 | Real-time collaboration | Phase 4 | Low |
| FEAT-028 | AI-assisted command suggestions | Phase 5 | Low |

## Deferred ⏸️ (Future / Optional)

| ID | Feature | Status | Reason | Future Action |
|----|---------|--------|--------|---------------|
| FEAT-021 | Production K8s deployment | deferred | Self-hosted优先, Docker compose cukup | Setup minikube/cloud K8s kalau butuh scale |
| FEAT-022 | Desktop app code signing | deferred | Unsigned build cukup untuk internal/test | Apply Apple Dev / Windows EV cert untuk public release |
| FEAT-023 | Mobile app store submission | deferred | APK sideload cukup, nggak butuh store fee | Register Google Play + Apple Dev kalau mau publik |

## Implemented ✅ (Quick Wins)

| ID | Feature | Status | Agent | Date |
|----|---------|--------|-------|------|
| FEAT-021a | Docker Compose production | **implemented** | devops | 2026-06-06 |
| FEAT-022a | Desktop unsigned builds | in_progress | dev | 2026-06-06 |
| FEAT-023a | Mobile APK sideload | pending | dev | 2026-06-06 |

**Total: 3 quick wins, 0 biaya external, bisa dikerjain kapan aja.**

## Legend

- ✅ `implemented` - Code complete, tested, merged
- 🔄 `in_progress` - Active development
- 📋 `planned` - Scheduled for future sprint
- ⏸️ `blocked` - Waiting on dependency
