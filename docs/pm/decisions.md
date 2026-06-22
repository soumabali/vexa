# vexa — Architecture Decision Records (ADR)

> **Agent:** Project Manager + Tech Lead  
> **Status:** Generated 2026-05-28  
> **Version:** 0.1.0

---

## ADR-001: Go Backend

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Backend language selection  
**Decision:** Use Go (Golang) for the backend API  
**Rationale:**
- Excellent concurrency with goroutines (critical for handling thousands of SSH connections)
- Fast compilation and startup times
- Strong standard library for networking and cryptography
- Native cross-compilation (Linux, Windows, macOS)
- Memory safety without garbage collection pauses affecting real-time connections
- Large ecosystem for SSH (`golang.org/x/crypto/ssh`)
- Docker-friendly (small images, fast builds)

**Alternatives Considered:**
- **Node.js:** Event loop can block on CPU-intensive crypto operations; less predictable performance
- **Python:** GIL limits concurrency; slower for protocol proxying
- **Rust:** Excellent but steeper learning curve; development velocity slower than Go
- **Java:** Heavy runtime, slower startup, larger Docker images

**Consequences:**
- ✅ Fast, concurrent protocol handling
- ✅ Small Docker images (~20MB)
- ✅ Easy to hire Go developers
- ⚠️ Less mature web framework ecosystem compared to Node.js
- ⚠️ Frontend developers may need to learn Go for full-stack work

---

## ADR-002: Next.js Frontend

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Web frontend framework selection  
**Decision:** Use Next.js 16 with React 19 and TypeScript  
**Rationale:**
- Server Components reduce JavaScript bundle size
- App Router for better performance and SEO
- Built-in API routes for simple backend-for-frontend
- Excellent developer experience (fast refresh, TypeScript)
- Large ecosystem (shadcn/ui, Tailwind CSS)
- Static export capability for CDN deployment
- Image optimization and font optimization built-in

**Alternatives Considered:**
- **Vue 3:** Good but smaller ecosystem for enterprise components
- **Svelte:** Fast but smaller talent pool
- **Plain React:** Missing Next.js optimizations and conventions
- **Angular:** Heavy, opinionated, slower development velocity

**Consequences:**
- ✅ Fast page loads with SSR/SSG
- ✅ Excellent TypeScript support
- ✅ Rich component ecosystem
- ⚠️ Server Components learning curve
- ⚠️ Vercel-centric (but works self-hosted)

---

## ADR-003: Tauri Desktop

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Desktop application framework selection  
**Decision:** Use Tauri v2 (Rust + WebView)  
**Rationale:**
- Smaller bundle size (~3MB vs 100MB+ for Electron)
- Better security (Rust memory safety, OS-native WebView)
- Better performance (no bundled Chromium)
- Native system integration (menu bar, notifications, auto-updater)
- Cross-platform (Windows, macOS, Linux)
- Can reuse web frontend code

**Alternatives Considered:**
- **Electron:** Mature but large bundle, security concerns, high memory usage
- **Flutter Desktop:** Single codebase but less native feel, larger size
- **Native (Qt/WPF):** Best performance but no code sharing with web
- **React Native Desktop:** Less mature, limited native APIs

**Consequences:**
- ✅ Small bundle size (~3-5MB)
- ✅ Low memory usage (~100MB vs 300MB Electron)
- ✅ Security-first design
- ✅ Code sharing with web frontend
- ⚠️ Rust learning curve for custom native modules
- ⚠️ WebView differences across platforms

---

## ADR-004: Flutter Mobile

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Mobile application framework selection  
**Decision:** Use Flutter 3.x with Dart  
**Rationale:**
- Single codebase for iOS and Android
- Native performance (compiles to ARM code)
- Rich widget library (Material Design + Cupertino)
- Hot reload for fast development
- Growing ecosystem and community
- Good for MVP with limited mobile team

**Alternatives Considered:**
- **React Native:** Good ecosystem but performance issues with complex UIs
- **Native (Kotlin/Swift):** Best performance but double development effort
- **Ionic/Capacitor:** Web-based, performance concerns
- **.NET MAUI:** Microsoft ecosystem lock-in

**Consequences:**
- ✅ Single codebase = faster development
- ✅ Native performance
- ✅ Consistent UI across platforms
- ⚠️ Dart learning curve
- ⚠️ FFI complexity for native SSH performance
- ⚠️ Larger app size (~20MB vs 10MB native)

---

## ADR-005: PostgreSQL Database

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Primary database selection  
**Decision:** Use PostgreSQL 16  
**Rationale:**
- ACID compliance (critical for financial/audit data)
- Excellent JSON support (for flexible metadata)
- Full-text search (for audit log queries)
- Row-level security
- Extensions (PostGIS, LTREE for host groups)
- Mature replication (streaming, logical)
- Strong Go driver (`lib/pq`, `pgx`)

**Alternatives Considered:**
- **MySQL 8:** Good but less feature-rich for JSON and extensions
- **MariaDB:** Similar to MySQL, fewer enterprise features
- **SQLite:** Good for embedded/desktop but not for multi-user server
- **MongoDB:** Flexible schema but lacks ACID for complex transactions
- **CockroachDB:** Distributed SQL but overkill for current scale

**Consequences:**
- ✅ ACID transactions for audit log integrity
- ✅ Rich JSON support for flexible metadata
- ✅ Excellent Go driver support
- ✅ LTREE extension for hierarchical host groups
- ⚠️ Vertical scaling limits (mitigated by read replicas)
- ⚠️ More complex setup than MySQL

---

## ADR-006: Rust SSH Core

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Low-level SSH protocol implementation  
**Decision:** Use Rust with `libssh2` bindings for SSH core library  
**Rationale:**
- Memory safety without GC (critical for long-running connections)
- Zero-cost abstractions (performance comparable to C)
- Excellent FFI with Go (via CGO) and Dart (via FFI)
- `libssh2` is mature and widely used
- Can compile to WASM for web terminal (future)
- Async/await support (`tokio`)

**Alternatives Considered:**
- **Pure Go (`golang.org/x/crypto/ssh`):** Easier integration but less feature-rich
- **C libssh:** Fast but manual memory management, unsafe
- **Rust `openssh`:** Pure Rust but less mature than `libssh2`
- **C++ libssh2:** Good but harder to integrate with Go/Dart

**Consequences:**
- ✅ Memory safety for long-running connections
- ✅ Excellent performance
- ✅ FFI with Go and Dart
- ⚠️ Rust learning curve for team
- ⚠️ Build complexity (Rust + Go toolchain)
- ⚠️ Cross-compilation challenges for mobile

---

## ADR-007: Redis Cache

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Caching and session storage  
**Decision:** Use Redis 7 for caching and session storage  
**Rationale:**
- In-memory performance (sub-millisecond latency)
- TTL support for automatic session expiry
- Pub/sub for real-time notifications
- Atomic operations (counters, rate limiting)
- Persistence options (RDB, AOF)
- Cluster mode for horizontal scaling

**Alternatives Considered:**
- **Memcached:** Simpler but lacks persistence and pub/sub
- **KeyDB:** Faster fork of Redis but less mature
- **Dragonfly:** Modern alternative but less ecosystem
- **In-memory Go map:** Fast but not shared across instances

**Consequences:**
- ✅ Sub-millisecond session lookups
- ✅ Automatic TTL expiry
- ✅ Pub/sub for real-time features
- ⚠️ Additional infrastructure component
- ⚠️ Data loss risk without persistence (mitigated by AOF)

---

## ADR-008: JWT Authentication

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Authentication mechanism  
**Decision:** Use JWT with RS256 signing and refresh token rotation  
**Rationale:**
- Stateless (no server-side session storage needed)
- RS256 allows key rotation without client changes
- Short-lived access tokens (15 min) + long-lived refresh tokens (7 days)
- Industry standard with mature libraries

**Alternatives Considered:**
- **Session cookies:** Stateful, requires sticky sessions
- **OAuth2/OIDC:** Overkill for self-hosted, adds complexity
- **API keys:** Simple but no expiry, hard to revoke
- **mTLS:** Excellent security but complex client setup

**Consequences:**
- ✅ Stateless, scalable
- ✅ Key rotation support
- ✅ Short-lived tokens limit exposure
- ⚠️ Token size (mitigated by compact claims)
- ⚠️ Revocation complexity (mitigated by blacklist in Redis)

---

## ADR-009: Docker Multi-Stage Builds

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Container build strategy  
**Decision:** Use multi-stage Docker builds with distroless runtime  
**Rationale:**
- Smaller final images (builder → scanner → runtime)
- Security scanning in CI pipeline
- Distroless runtime reduces attack surface
- Reproducible builds

**Alternatives Considered:**
- **Single-stage builds:** Simpler but larger images
- **Alpine Linux:** Small but musl libc issues
- **Ubuntu/Debian:** Familiar but larger attack surface

**Consequences:**
- ✅ Smaller images (~20MB vs 200MB)
- ✅ Reduced attack surface
- ✅ Security scanning integrated
- ⚠️ Build time increases (mitigated by layer caching)
- ⚠️ Debugging harder in distroless (mitigated by debug images)

---

## ADR-010: GitHub Actions CI/CD

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** CI/CD platform selection  
**Decision:** Use GitHub Actions with self-hosted runners for heavy jobs  
**Rationale:**
- Native GitHub integration
- Large marketplace of actions
- Matrix builds for multiple platforms
- Free minutes for open source
- Self-hosted runners for ARM builds and heavy tests

**Alternatives Considered:**
- **GitLab CI:** Good but requires GitLab hosting
- **Jenkins:** Flexible but maintenance overhead
- **CircleCI:** Good but limited free tier
- **Drone CI:** Lightweight but smaller ecosystem

**Consequences:**
- ✅ Native GitHub integration
- ✅ Large action marketplace
- ✅ Free tier sufficient for current scale
- ⚠️ Public minutes limited (mitigated by self-hosted runners)
- ⚠️ Vendor lock-in (mitigated by portable workflow files)

---

## ADR-011: Zustand State Management

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Frontend state management  
**Decision:** Use Zustand for global state, React Query for server state  
**Rationale:**
- Zustand: Lightweight, TypeScript-friendly, minimal boilerplate
- React Query: Caching, background refetch, optimistic updates
- Separation of concerns (client vs server state)
- No provider wrapping needed

**Alternatives Considered:**
- **Redux:** Mature but verbose, more boilerplate
- **MobX:** Powerful but complex for simple apps
- **Context API:** Built-in but performance issues with frequent updates
- **Recoil/Jotai:** Atom-based but newer, less ecosystem

**Consequences:**
- ✅ Lightweight (~1KB)
- ✅ TypeScript support
- ✅ Minimal boilerplate
- ✅ React Query handles server state elegantly
- ⚠️ Less devtools than Redux (mitigated by Zustand devtools)

---

## ADR-012: Prometheus + Grafana Monitoring

**Status:** Accepted  
**Date:** 2026-05-28  
**Context:** Monitoring and observability  
**Decision:** Use Prometheus for metrics, Grafana for dashboards, Loki for logs  
**Rationale:**
- Industry standard for cloud-native monitoring
- Pull-based metrics (reliable, scalable)
- Grafana: Rich visualization, alerting
- Loki: Log aggregation without full-text index overhead
- Easy integration with Kubernetes

**Alternatives Considered:**
- **DataDog:** Managed but expensive
- **New Relic:** Managed but vendor lock-in
- **InfluxDB:** Good but less ecosystem
- **ELK Stack:** Powerful but resource-intensive

**Consequences:**
- ✅ Open source, no licensing costs
- ✅ Kubernetes-native
- ✅ Rich alerting capabilities
- ⚠️ Self-hosted maintenance (mitigated by managed Prometheus if needed)
- ⚠️ Storage costs for long retention (mitigated by downsampling)

---

## Decision Log Summary

| ADR | Decision | Status | Date |
|-----|----------|--------|------|
| ADR-001 | Go backend | Accepted | 2026-05-28 |
| ADR-002 | Next.js frontend | Accepted | 2026-05-28 |
| ADR-003 | Tauri desktop | Accepted | 2026-05-28 |
| ADR-004 | Flutter mobile | Accepted | 2026-05-28 |
| ADR-005 | PostgreSQL database | Accepted | 2026-05-28 |
| ADR-006 | Rust SSH core | Accepted | 2026-05-28 |
| ADR-007 | Redis cache | Accepted | 2026-05-28 |
| ADR-008 | JWT authentication | Accepted | 2026-05-28 |
| ADR-009 | Docker multi-stage | Accepted | 2026-05-28 |
| ADR-010 | GitHub Actions CI/CD | Accepted | 2026-05-28 |
| ADR-011 | Zustand state management | Accepted | 2026-05-28 |
| ADR-012 | Prometheus + Grafana | Accepted | 2026-05-28 |
| ADR-013 | WireGuard kernel module + wgctrl | Accepted | 2026-06-06 |

---

---

## ADR-013: WireGuard via Kernel Module (Primary) + Userspace Fallback

**Status:** Accepted
**Date:** 2026-06-06
**Context:** WireGuard tunnel implementation for Phase 2 (FEAT-024)
**Decision:** Use kernel `wireguard.ko` module as primary implementation with `wgctrl` Go library for netlink control, and `wireguard-go` userspace implementation as fallback for containers/K8s

**Rationale:**
- Kernel module provides best performance (zero-copy, kernel-space routing)
- `wgctrl` (golang.zx2c4.com/wireguard/wgctrl) provides native Go interface to netlink API
- Docker containers can use kernel module with `CAP_NET_ADMIN`
- K8s environments may not have kernel module —`wireguard-go` userspace implementation serves as fallback
- Key generation uses Curve25519 via `golang.org/x/crypto/curve25519` (already in dependency tree)
- Config export uses standard INI-format .conf files (interoperable with all WireGuard clients)
- QR code generation via `github.com/skip2/go-qrcode` for mobile config sharing

**Alternatives Considered:**
- **Userspace only (wireguard-go):** Consistent across all environments but ~15% performance overhead
- **Subprocess `wg-quick`:** Simpler but slower, harder to get real-time stats
- **Custom UDP tunnel:** More flexible but reinvents the wheel, no WireGuard client compatibility
- **OpenVPN:** Overhead of TLS handshake, CA management, larger codebase

**Consequences:**
- ✅ Best performance with kernel module
- ✅ Native Go control via netlink
- ✅ Full WireGuard client compatibility (standard .conf format)
- ✅ Container-friendly with fallback path
- ⚠️ Kernel module dependency = documentation requirement
- ⚠️ Two code paths to maintain (kernel + userspace)
- ⚠️ Privileged container mode needed for kernel module access

---

## Rejected Decisions

| Decision | Reason for Rejection |
|----------|---------------------|
| Electron for desktop | Large bundle size, security concerns |
| React Native for mobile | Performance issues, smaller ecosystem |
| MongoDB primary | Lacks ACID for audit log integrity |
| MySQL primary | Less JSON/extension support than PostgreSQL |
| OAuth2/OIDC | Overkill for self-hosted MVP |
