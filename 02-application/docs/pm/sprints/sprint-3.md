# Sprint 3 — Protocol Layer

> **Phase:** 2 | **Weeks:** 7–10 | **Duration:** 20 working days  
> **Sprint Goal:** Implement full protocol proxy layer with file manager, tunneling, and WebSocket terminal  
> **Story Points:** 89  > **Status:** Not Started  > **Depends On:** Sprint 2 (Gateway + Sessions)

---

## User Stories

### US-016: WebSocket Terminal
**As a** user, **I want** a browser-based terminal, **so that** I can access servers without desktop client.

**Acceptance Criteria:**
- [ ] WebSocket endpoint for bidirectional terminal data
- [ ] Support for SSH sessions (pty allocation)
- [ ] Terminal resize handling (SIGWINCH)
- [ ] Binary protocol for efficient data transfer
- [ ] Heartbeat/ping to detect stale connections
- [ ] Multi-tab support (multiple sessions per user)

**Definition of Done:**
- [ ] Terminal renders correctly with `xterm.js`
- [ ] Responds to resize events in < 50ms
- [ ] Survives network interruption (auto-reconnect)
- [ ] Memory usage stable with 50 concurrent terminals

**Story Points:** 13

---

### US-017: SFTP File Manager
**As a** user, **I want** a file manager for SFTP, **so that** I can transfer files securely.

**Acceptance Criteria:**
- [ ] Browse remote file system (list directories)
- [ ] Upload files via drag-and-drop
- [ ] Download files (single and batch)
- [ ] Create/delete directories
- [ ] File permissions display and modification
- [ ] Progress indicators for large transfers
- [ ] Resume interrupted transfers

**Definition of Done:**
- [ ] Can transfer 1GB file without corruption
- [ ] Directory listing for 10,000 files in < 2s
- [ ] Concurrent transfer limit enforced (5 per user)
- [ ] Audit log records all file operations

**Story Points:** 13

---

### US-018: Protocol Translation
**As a** user, **I want** seamless protocol switching, **so that** I can use different protocols without reconfiguring.

**Acceptance Criteria:**
- [ ] Unified connection API regardless of protocol
- [ ] Protocol-specific options (e.g., RDP resolution, VNC quality)
- [ ] Auto-detect protocol from port number (22→SSH, 3389→RDP, 5900→VNC)
- [ ] Protocol negotiation fallback
- [ ] Connection templates for common configurations

**Definition of Done:**
- [ ] Can switch from SSH to RDP without re-entering credentials
- [ ] Protocol detection accuracy > 95%
- [ ] Fallback works when auto-detection fails

**Story Points:** 8

---

### US-019: Tunneling Support
**As an** advanced user, **I want** port forwarding/tunneling, **so that** I can access services behind firewalls.

**Acceptance Criteria:**
- [ ] Local port forwarding (`-L` equivalent)
- [ ] Remote port forwarding (`-R` equivalent)
- [ ] Dynamic SOCKS proxy (`-D` equivalent)
- [ ] Tunnel management UI (list, create, delete)
- [ ] Tunnel health monitoring
- [ ] Automatic reconnection on failure

**Definition of Done:**
- [ ] Can access internal web service via tunnel
- [ ] Tunnel survives gateway restart
- [ ] Performance: < 5ms latency overhead
- [ ] Audit log records tunnel creation/usage

**Story Points:** 13

---

### US-020: Session Recording & Playback
**As an** administrator, **I want** session recordings, **so that** I can review activity for compliance.

**Acceptance Criteria:**
- [ ] Text recording for SSH sessions (asciinema format)
- [ ] Video recording for RDP/VNC sessions (WebM)
- [ ] Playback interface with speed control
- [ ] Search within recordings (text only)
- [ ] Export recordings (MP4, GIF)
- [ ] Automatic deletion after retention period

**Definition of Done:**
- [ ] Recording quality acceptable for audit review
- [ ] Playback synchronized with audit log timestamps
- [ ] Storage: 1 hour SSH = ~10MB, 1 hour RDP = ~100MB
- [ ] GDPR compliance: user can request deletion

**Story Points:** 13

---

### US-021: File Transfer Protocols
**As a** user, **I want** multiple file transfer options, **so that** I can choose the best method for my needs.

**Acceptance Criteria:**
- [ ] SFTP for SSH connections
- [ ] RDP drive redirection for RDP connections
- [ ] VNC clipboard sharing for VNC connections
- [ ] Drag-and-drop file transfer in web client
- [ ] Transfer queue with pause/resume
- [ ] Bandwidth limiting option

**Definition of Done:**
- [ ] All transfer methods tested with 1GB files
- [ ] Progress tracking accurate to ±1%
- [ ] Error recovery for interrupted transfers
- [ ] Audit log records all transfers

**Story Points:** 13

---

### US-022: Connection Health Monitoring
**As a** user, **I want** real-time connection health, **so that** I know when servers are unreachable.

**Acceptance Criteria:**
- [ ] Background health checks every 60s
- [ ] Visual indicators: green (reachable), yellow (slow), red (unreachable)
- [ ] Latency display (ping time)
- [ ] Last seen timestamp
- [ ] Health check history (24h graph)
- [ ] Alert on status change (configurable)

**Definition of Done:**
- [ ] Health check doesn't impact server performance
- [ ] Status updates within 5s of change
- [ ] Historical data accurate within 1s
- [ ] Alert delivery < 30s

**Story Points:** 8

---

## Dependencies

| Dependency | Source | Impact |
|------------|--------|--------|
| Sprint 2 completion | Previous sprint | Blocks protocol work |
| WebSocket library | `gorilla/websocket` | Blocks terminal |
| `xterm.js` frontend | npm package | Blocks terminal UI |
| SFTP library | `pkg/sftp` | Blocks file manager |
| Recording storage | S3/MinIO | Blocks session recording |

---

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation | Owner |
|------|-------------|--------|------------|-------|
| WebSocket scalability | High | High | Horizontal scaling with sticky sessions | DevOps |
| File transfer security | Medium | Critical | Scan uploads, sandbox download processing | Security |
| Recording storage costs | Medium | High | Compression, tiered storage, auto-cleanup | DevOps |
| Protocol compatibility | Medium | Medium | Test matrix with common server versions | QA |

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| WebSocket library | `gorilla/websocket` | Most popular Go WebSocket library |
| Terminal emulator | `xterm.js` | VS Code's terminal, well-maintained |
| SFTP library | `pkg/sftp` | Pure Go, actively maintained |
| Recording format | asciinema for SSH, WebM for RDP/VNC | Open standards, compressible |
| Storage backend | MinIO (S3-compatible) | Self-hosted, cost-effective |

---

## Sprint Review Checklist

- [ ] WebSocket terminal tested with 50 concurrent users
- [ ] SFTP file transfer tested with 1GB files
- [ ] Protocol translation accuracy > 95%
- [ ] Tunneling verified with real network topology
- [ ] Session recording playback reviewed by QA
- [ ] Performance tests pass (p95 < 100ms)
- [ ] Security review of file transfer completed
- [ ] Sprint retrospective completed

---

## Sprint Retrospective Notes

*To be filled after sprint completion*

**What went well:**
- 

**What could be improved:**
- 

**Action items for next sprint:**
- 
