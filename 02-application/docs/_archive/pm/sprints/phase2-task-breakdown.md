# Phase 2 — WireGuard Tunnel Support: Task Breakdown

> **Feature ID:** FEAT-024
> **Priority:** High
> **Target Complete:** 2026-06-13
> **Status:** Active — Sprint 2.1 (PM + Design)

---

## Sprint 2.1 — PM + Design (Days 1–2)

**Sprint Goal:** Finalize spec, create task breakdown, wireframe tunnel UI
**Status:** In Progress

### PM-001: Spec Refinement & Task Breakdown
**As a** project manager, **I want** a finalized spec and task breakdown, **so that** the team has clear requirements.

**Acceptance Criteria:**
- [ ] Phase 2 spec reviewed against main spec.md for consistency
- [ ] All integration points documented (Host, Vault, Audit, WebSocket)
- [ ] Task breakdown created with sprints, user stories, story points
- [ ] Risk register updated with WireGuard-specific risks
- [ ] FEATURES.md updated with current status

**Story Points:** 3

### DES-001: Tunnel UI Wireframes (Host Tab)
**As a** designer, **I want** wireframes for the host detail Tunnel tab, **so that** devs have a visual reference.

**Acceptance Criteria:**
- [ ] Host detail page with "Tunnel" tab alongside existing tabs
- [ ] Create tunnel form (allowed IPs, port, DNS)
- [ ] Tunnel status badge (up/down/error)
- [ ] Enable/disable toggle
- [ ] Rotate key button
- [ ] Delete tunnel with confirmation

**Story Points:** 5

### DES-002: Tunnel List Page + Config Modal Wireframes
**As a** designer, **I want** wireframes for the tunnel list and config export, **so that** all tunnel flows are covered.

**Acceptance Criteria:**
- [ ] Tunnel list page with filter by host/status
- [ ] Tunnel detail view showing config, stats, timestamps
- [ ] Config export modal (.conf viewer + download)
- [ ] QR code display for mobile import
- [ ] Mobile deep link sharing wireframe

**Story Points:** 5

### DES-003: Design System Updates
**As a** designer, **I want** design system updates for tunnel components, **so that** UI is consistent.

**Acceptance Criteria:**
- [ ] Tunnel status badges (wireguard-up, wireguard-down, wireguard-error variants)
- [ ] QR code component spec
- [ ] Config copy/download button styles
- [ ] Stats display component (bytes in/out, last handshake)

**Story Points:** 3

**Sprint 2.1 Total:** 16 story points

---

## Sprint 2.2 — Backend Dev (Days 3–5)

**Sprint Goal:** Implement WireGuard service, database, API, WebSocket, key rotation
**Status:** Planned

### DEV-001: Database Migration — wireguard_tunnels
**As a** backend developer, **I want** the wireguard_tunnels table, **so that** tunnel configs persist.

**Acceptance Criteria:**
- [ ] Migration creates `wireguard_tunnels` table (UUID PK, user_id FK, host_id FK, interface_name, server_public_key, client_public_key, preshared_key, server_ip, client_ip, listen_port, allowed_ips CIDR[], dns_servers INET[], is_enabled, last_rotated_at, timestamps)
- [ ] Unique constraint on (user_id, host_id)
- [ ] Unique constraint on interface_name
- [ ] Indexes on user_id, host_id, is_enabled
- [ ] Migration is reversible (rollback support)
- [ ] Seed data for development

**Story Points:** 5

### DEV-002: WireGuard Service Scaffold + Interface Controller
**As a** backend developer, **I want** a WireGuard service, **so that** the server can create/manage WG interfaces.

**Acceptance Criteria:**
- [ ] `WireGuardService` struct with Config Manager, Interface Controller, Stats Collector, Rotation Scheduler components
- [ ] Config Manager: CRUD tunnel configs in DB, subnet validation, IP assignment from configurable pool
- [ ] Interface Controller: create/destroy `wg<N>` interfaces via `wg` command or `wgctrl` netlink, set peers/keys
- [ ] Stats Collector: read `wg show` or `/proc/net/dev` for traffic counters
- [ ] Rotation Scheduler: cron job for auto-rotation (configurable interval, default 90 days)
- [ ] Service runs as unprivileged user with CAP_NET_ADMIN
- [ ] Proper error handling and logging

**Story Points:** 13

### DEV-003: REST API Endpoints — Tunnel CRUD + Rotate + Config
**As an** API consumer, **I want** REST endpoints for tunnels, **so that** clients can manage tunnels.

**Acceptance Criteria:**
- [ ] `POST /api/v1/tunnels` — Create tunnel (generates server keypair, assigns IP, creates interface)
- [ ] `GET /api/v1/tunnels` — List user tunnels (filterable by host_id, is_enabled)
- [ ] `GET /api/v1/tunnels/:id` — Get tunnel details
- [ ] `PATCH /api/v1/tunnels/:id` — Update (allowed IPs, DNS, port)
- [ ] `DELETE /api/v1/tunnels/:id` — Remove tunnel (destroys interface)
- [ ] `POST /api/v1/tunnels/:id/rotate` — Rotate server keypair (generates new keys, updates peer)
- [ ] `GET /api/v1/tunnels/:id/config` — Download .conf file
- [ ] `POST /api/v1/tunnels/:id/enable` — Enable interface
- [ ] `POST /api/v1/tunnels/:id/disable` — Disable interface
- [ ] `GET /api/v1/tunnels/:id/stats` — Traffic statistics
- [ ] Rate limiting: max 10 tunnels per user
- [ ] All endpoints require auth + appropriate authorization

**Story Points:** 13

### DEV-004: WebSocket Events — Tunnel Lifecycle
**As a** frontend developer, **I want** real-time tunnel events, **so that** UI stays in sync.

**Acceptance Criteria:**
- [ ] `tunnel:create` / `tunnel:created` event pair
- [ ] `tunnel:rotate` / `tunnel:rotated` event pair
- [ ] `tunnel:destroy` / `tunnel:destroyed` event pair
- [ ] `tunnel:stats` event (bytes_sent, bytes_rcvd, last_handshake)
- [ ] Events follow existing WebSocket protocol conventions
- [ ] Events include tunnel_id for client-side routing

**Story Points:** 5

### DEV-005: Integration with Host, Vault, Audit Services
**As a** system architect, **I want** tunnel service integrated with existing services, **so that** system consistency is maintained.

**Acceptance Criteria:**
- [ ] Host Service: `wireguard_enabled` boolean + `wireguard_port` fields on host record
- [ ] Vault Service: Tunnel private keys stored using same client-side encryption as SSH keys
- [ ] Audit Service: All tunnel create/connect/rotate/delete events logged with same schema as session events
- [ ] Tunnel create/delete events trigger host record updates

**Story Points:** 8

**Sprint 2.2 Total:** 44 story points

---

## Sprint 2.3 — Frontend Dev (Days 6–7)

**Sprint Goal:** Implement tunnel UI in web app
**Status:** Planned

### FE-001: Host Detail — Tunnel Tab
**As a** user, **I want** a Tunnel tab on the host detail page, **so that** I can manage tunnels per host.

**Acceptance Criteria:**
- [ ] "Tunnel" tab visible on host detail page (when WireGuard is available)
- [ ] Create tunnel form (allowed IPs input, port, DNS servers)
- [ ] Tunnel status display (up/down/error badge)
- [ ] Enable/disable toggle
- [ ] Rotate key button with confirmation
- [ ] Delete tunnel button with confirmation dialog
- [ ] Traffic stats display (bytes in/out, last handshake time)

**Story Points:** 8

### FE-002: Tunnel List Page
**As a** user, **I want** a tunnel list page, **so that** I can see all my tunnels.

**Acceptance Criteria:**
- [ ] New route: `/tunnels`
- [ ] Tunnel list with columns: host, status, IP, port, created, actions
- [ ] Filter by host name
- [ ] Filter by status (enabled/disabled)
- [ ] Click to view tunnel details
- [ ] Empty state when no tunnels exist
- [ ] Loading skeleton

**Story Points:** 5

### FE-003: Config Export Modal (.conf + QR)
**As a** user, **I want** to download tunnel config, **so that** I can connect my WireGuard client.

**Acceptance Criteria:**
- [ ] Config viewer modal showing .conf file content
- [ ] Copy to clipboard button
- [ ] Download .conf file button
- [ ] QR code display for mobile scanning
- [ ] QR code generated server-side (or via library)
- [ ] Config preview before download

**Story Points:** 5

### FE-004: Mobile Config Sharing (Deep Link)
**As a** mobile user, **I want** to share tunnel config to WireGuard app, **so that** I can connect from my phone.

**Acceptance Criteria:**
- [ ] "Open in WireGuard" deep link on mobile
- [ ] Config shared via `wireguard://` URI scheme or .conf export
- [ ] Works on iOS (WireGuard app registered scheme)
- [ ] Works on Android (WireGuard app registered scheme)
- [ ] Fallback to QR code when deep link unavailable

**Story Points:** 3

**Sprint 2.3 Total:** 21 story points

---

## Sprint 2.4 — QA (Day 8)

**Sprint Goal:** Comprehensive testing of WireGuard feature
**Status:** Planned

### QA-001: Unit Tests — Service Layer + Key Generation
**As a** QA engineer, **I want** unit tests for the WireGuard service, **so that** core logic is verified.

**Acceptance Criteria:**
- [ ] Key generation tests (Curve25519, valid formats)
- [ ] Config Manager CRUD tests
- [ ] IP allocation tests (pool exhaustion, overlap prevention)
- [ ] Subnet validation tests
- [ ] Rotation scheduler tests
- [ ] Config export format tests (.conf syntax valid)
- [ ] Rate limit enforcement tests

**Story Points:** 8

### QA-002: Integration Tests — Tunnel Lifecycle
**As a** QA engineer, **I want** integration tests for tunnel create/enable/disable/delete, **so that** the full lifecycle works.

**Acceptance Criteria:**
- [ ] Create tunnel → verify interface created on server
- [ ] Enable/disable tunnel → verify interface state
- [ ] Rotate keys → verify new keys applied
- [ ] Delete tunnel → verify interface destroyed, DB cleaned
- [ ] Config download → verify .conf is valid WireGuard format
- [ ] Concurrent tunnel creation (up to 10 per user)
- [ ] 11th tunnel → rate limit error

**Story Points:** 8

### QA-003: Security Tests — Authz, Rate Limits, Key Handling
**As a** security engineer, **I want** security tests for tunnels, **so that** no authz gaps exist.

**Acceptance Criteria:**
- [ ] User A cannot access User B's tunnels (verify per-tunnel authz)
- [ ] Unauthenticated requests return 401
- [ ] Non-admin cannot access admin tunnel endpoints
- [ ] Rate limit: 10 tunnels per user enforced
- [ ] Rate limit: 1000 total tunnels per server enforced
- [ ] Private keys never exposed in API responses
- [ ] Audit log entries created for all tunnel lifecycle events

**Story Points:** 5

### QA-004: E2E Tests — Full Tunnel Lifecycle
**As a** QA engineer, **I want** end-to-end tests, **so that** the complete user flow works.

**Acceptance Criteria:**
- [ ] User creates tunnel from host detail page (UI)
- [ ] Server creates `wg<N>` interface (verified via `wg show`)
- [ ] Client downloads .conf and connects with WireGuard client (container-based)
- [ ] Traffic flows through tunnel (verified with ping + iperf)
- [ ] Stats API returns real-time bytes in/out
- [ ] Key rotation preserves connectivity (graceful rekey, no packet loss)
- [ ] All 59 existing tests still pass (no regressions)
- [ ] New test coverage > 80% for wireguard package

**Story Points:** 13

**Sprint 2.4 Total:** 34 story points

---

## Sprint 2.5 — DevOps (Days 9–10)

**Sprint Goal:** Container/infra changes for WireGuard support
**Status:** Planned

### OPS-001: Docker — Add wireguard-tools
**As a** DevOps engineer, **I want** wireguard-tools in the container, **so that** WG commands work.

**Acceptance Criteria:**
- [ ] `wireguard-tools` installed in Docker image
- [ ] `linux-headers` or kernel module checked at runtime
- [ ] Container has `CAP_NET_ADMIN` capability
- [ ] Container can create `wg<N>` interfaces
- [ ] Health check verifies WG kernel module availability

**Story Points:** 3

### OPS-002: K8s — Privileged Mode + hostNetwork for WG
**As a** DevOps engineer, **I want** K8s manifests updated, **so that** WireGuard works in cluster.

**Acceptance Criteria:**
- [ ] Pod security context with `CAP_NET_ADMIN`
- [ ] hostNetwork or dedicated network interface for WG traffic
- [ ] Node-level kernel module check in init container
- [ ] DaemonSet or sidecar pattern for WG interface management
- [ ] Updated resource requests/limits for tunnel traffic

**Story Points:** 5

### OPS-003: Terraform — Tunnel Port Ranges
**As a** DevOps engineer, **I want** Terraform updated, **so that** firewall rules allow WG ports.

**Acceptance Criteria:**
- [ ] Configurable WireGuard port range (default: 51820-51830)
- [ ] Security group rules for UDP port range
- [ ] Variable for WG subnet allocation (default: 10.200.200.0/24)
- [ ] Output tunnel endpoint IP/hostname

**Story Points:** 3

### OPS-004: CI — Add WG Kernel Module Tests
**As a** DevOps engineer, **I want** CI to validate WG availability, **so that** we catch kernel issues early.

**Acceptance Criteria:**
- [ ] CI runner checks `modprobe wireguard` succeeds
- [ ] GitHub Actions workflow creates test tunnel and verifies connectivity
- [ ] Matrix test across kernel versions (if applicable)
- [ ] Fail CI if WG kernel module unavailable
- [ ] Documentation for WG kernel requirements

**Story Points:** 3

**Sprint 2.5 Total:** 14 story points

---

## Summary

| Sprint | Focus | Story Points | Dependencies |
|--------|-------|-------------|--------------|
| 2.1 | PM + Design | 16 | None |
| 2.2 | Backend Dev | 44 | Sprint 2.1 (wireframes) |
| 2.3 | Frontend Dev | 21 | Sprint 2.2 (API) |
| 2.4 | QA | 34 | Sprint 2.2 + 2.3 |
| 2.5 | DevOps | 14 | Sprint 2.2 (WG service) |

**Phase 2 Total:** 129 story points
**Target:** 10 working days (2026-06-06 to 2026-06-13)

---

## Dependencies & Critical Path

```
Sprint 2.1 ───────────────────── Sprint 2.4 ───┐
    │                                            │
    ▼                                            ▼
Sprint 2.2 ─────────────── Sprint 2.5 ────► Release
    │
    ▼
Sprint 2.3 ───────────────┘
```

**Critical Path:** 2.1 → 2.2 → 2.3 → 2.4 (frontend waits on API)
**Parallelizable:** 2.4 (QA can start on backend tests before frontend is done), 2.5 (DevOps can start once WG service is stable)
