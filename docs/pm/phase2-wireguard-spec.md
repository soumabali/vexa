# Phase 2 — WireGuard Tunnel Support

**Feature ID:** FEAT-024
**Priority:** High
**Status:** In Progress
**Started:** 2026-06-06
**Target Complete:** 2026-06-13

---

## 1. Overview

Add WireGuard VPN tunnel support to vexa, allowing users to create encrypted tunnels to remote networks without opening additional ports. Integrates with existing host management, credential vault, and session recording.

---

## 2. Requirements

### 2.1 Functional
- Generate WireGuard key pairs (Curve25519) per user/host
- Create/manage WireGuard interfaces on server side
- Assign IP addresses from configurable subnets
- Configure peer endpoints dynamically
- Support multiple concurrent tunnels per user
- Auto-rotate keys on configurable schedule
- Export/import WireGuard config files (.conf)

### 2.2 Security
- Server never stores client private keys
- Keys generated client-side, public key sent to server
- Preshared key (PSK) support for extra layer
- Configurable allowed IPs per tunnel
- Audit log for all tunnel create/connect/rotate events
- Rate limit: max 10 tunnels per user

### 2.3 Integration Points
| System | Integration |
|--------|-------------|
| Host Service | Host record gets `wireguard_enabled` + `wireguard_port` |
| Vault Service | Tunnel credentials stored with same encryption as SSH keys |
| Audit Service | Tunnel events logged with same schema as session events |
| WebSocket | Real-time tunnel status (up/down/transfer stats) |
| Mobile/Desktop | Native WireGuard support or embedded daemon |

---

## 3. API Design

### 3.1 REST Endpoints

```
POST   /api/v1/tunnels              — Create tunnel
GET    /api/v1/tunnels              — List user tunnels
GET    /api/v1/tunnels/:id          — Get tunnel details
PATCH  /api/v1/tunnels/:id          — Update (rotate key, change allowed IPs)
DELETE /api/v1/tunnels/:id          — Remove tunnel
POST   /api/v1/tunnels/:id/rotate   — Rotate key pair
GET    /api/v1/tunnels/:id/config   — Download .conf file
POST   /api/v1/tunnels/:id/enable   — Enable interface
POST   /api/v1/tunnels/:id/disable  — Disable interface
GET    /api/v1/tunnels/:id/stats    — Traffic statistics (bytes in/out, last handshake)
```

### 3.2 WebSocket Events

```
→ tunnel:create {host_id, allowed_ips, port}
← tunnel:created {tunnel_id, public_key, assigned_ip, config}
→ tunnel:rotate {tunnel_id}
← tunnel:rotated {tunnel_id, new_public_key, new_config}
→ tunnel:destroy {tunnel_id}
← tunnel:destroyed {tunnel_id}
→ tunnel:stats {tunnel_id}
← tunnel:stats {tunnel_id, bytes_sent, bytes_rcvd, last_handshake}
```

---

## 4. Database Schema

```sql
CREATE TABLE wireguard_tunnels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    host_id UUID NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    interface_name VARCHAR(16) NOT NULL,  -- wg0, wg1, etc
    server_public_key TEXT NOT NULL,
    client_public_key TEXT NOT NULL,
    preshared_key TEXT,  -- encrypted in vault, nullable
    server_ip INET NOT NULL,  -- 10.200.200.1/24
    client_ip INET NOT NULL,  -- 10.200.200.2/24
    listen_port INT NOT NULL CHECK (listen_port BETWEEN 1 AND 65535),
    allowed_ips CIDR[] DEFAULT '{0.0.0.0/0}'::CIDR[],
    dns_servers INET[] DEFAULT '{}'::INET[],
    is_enabled BOOLEAN DEFAULT false,
    last_rotated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT unique_user_tunnel_per_host UNIQUE (user_id, host_id),
    CONSTRAINT unique_interface_name UNIQUE (interface_name)
);

CREATE INDEX idx_wg_tunnels_user ON wireguard_tunnels(user_id);
CREATE INDEX idx_wg_tunnels_host ON wireguard_tunnels(host_id);
CREATE INDEX idx_wg_tunnels_enabled ON wireguard_tunnels(is_enabled);
```

---

## 5. Architecture

### 5.1 Server Side (Go)

```
┌─────────────────────────────────────────────┐
│           WireGuard Service                  │
│  ┌────────────┐  ┌────────────┐  ┌────────┐│
│  │ Config Mgr │  │ Interface  │  │  Stats  ││
│  │ (DB CRUD)  │  │ (wg ctrl)  │  │Collector││
│  └─────┬──────┘  └─────┬──────┘  └───┬────┘│
│        │               │             │     │
│  ┌─────▼───────────────▼─────────────▼────┐│
│  │         WireGuard Controller            ││
│  │  (exec `wg set` / `wg-quick` / netlink)││
│  └─────────────────────────────────────────┘│
│                    │                        │
│  ┌─────────────────▼──────────────────────┐│
│  │         Linux Kernel (wireguard.ko)   ││
│  └────────────────────────────────────────┘│
└─────────────────────────────────────────────┘
```

**Key Components:**
- **Config Manager:** CRUD tunnel configs in DB, validate subnets, assign IPs
- **Interface Controller:** Create/destroy `wg<N>` interfaces, set peers, keys
- **Stats Collector:** Read `/proc/net/dev` or `wg show` for traffic counters
- **Rotation Scheduler:** Cron job to auto-rotate keys older than N days

### 5.2 Client Config Export

```ini
[Interface]
PrivateKey = <client_private_key_generated_client_side>
Address = 10.200.200.2/24
DNS = 1.1.1.1, 8.8.8.8

[Peer]
PublicKey = <server_public_key>
PresharedKey = <optional_psk>
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = server.example.com:51820
PersistentKeepalive = 25
```

### 5.3 Frontend Changes

| Screen | Changes |
|--------|---------|
| Host Detail | Add "Tunnel" tab with create/rotate/destroy |
| Tunnel List | New page, filter by host/status |
| Tunnel Detail | Config viewer, QR code, download .conf |
| Mobile | Share config to WireGuard app (deep link) |

---

## 6. Implementation Plan

### Sprint 2.1 (Days 1-2): PM + Design
- [ ] PM: Finalize spec, create task breakdown
- [ ] Designer: Wireframe tunnel UI (host tab + tunnel list + config modal)
- [ ] Designer: Design system updates (tunnel status badges, QR code component)

### Sprint 2.2 (Days 3-5): Backend Dev
- [ ] Dev: Database migration (wireguard_tunnels table)
- [ ] Dev: WireGuard service scaffold + interface controller
- [ ] Dev: REST API endpoints (CRUD + rotate + config)
- [ ] Dev: WebSocket events
- [ ] Dev: Key rotation scheduler
- [ ] Dev: Integration with Host/Vault/Audit services

### Sprint 2.3 (Days 6-7): Frontend Dev
- [ ] Dev: Host detail — Tunnel tab
- [ ] Dev: Tunnel list page
- [ ] Dev: Config export modal (.conf + QR)
- [ ] Dev: Mobile config sharing (deep link)

### Sprint 2.4 (Day 8): QA
- [ ] QA: Unit tests (service layer, key generation)
- [ ] QA: Integration tests (create tunnel, enable/disable)
- [ ] QA: Security tests (rate limits, authz)
- [ ] QA: E2E tests (full tunnel lifecycle)

### Sprint 2.5 (Day 9-10): DevOps
- [ ] DevOps: Docker — add `wireguard-tools` to container
- [ ] DevOps: K8s — privileged mode + hostNetwork for WG
- [ ] DevOps: Terraform — update for tunnel port ranges
- [ ] DevOps: CI — add WG kernel module tests

---

## 7. Security Considerations

| Threat | Mitigation |
|--------|------------|
| Private key exposure | Generated client-side, never sent to server |
| Key compromise | Auto-rotation every 90 days |
| Tunnel enumeration | UUID-based IDs, rate-limited list |
| Privilege escalation | WG service runs as unprivileged user with CAP_NET_ADMIN |
| DoS (max tunnels) | Rate limit: 10 tunnels/user, 1000 total per server |
| Config tampering | .conf files signed with HMAC (optional) |

---

## 8. Dependencies

- `wireguard-tools` (apt package) — server side
- `golang.zx2c4.com/wireguard/wgctrl` — Go netlink control
- `github.com/skip2/go-qrcode` — QR code generation
- `wireguard.ko` kernel module — server runtime requirement

---

## 9. Acceptance Criteria

- [ ] User can create tunnel from host detail page
- [ ] Server auto-creates `wg<N>` interface with valid config
- [ ] Client can download `.conf` and connect with any WG client
- [ ] Traffic flows through tunnel (verified with `ping` + `iperf`)
- [ ] Stats API returns real-time bytes in/out
- [ ] Key rotation preserves connectivity (graceful rekey)
- [ ] All 59 existing tests still pass (no regressions)
- [ ] New test coverage > 80% for wireguard package

---

**Next:** Spawn PM agent for Sprint 2.1 spec refinement → Designer for wireframes → Dev for backend implementation.
