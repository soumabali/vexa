# vexa — Risk Register

> **Agent:** Project Manager + Risk Management  
> **Status:** Generated 2026-05-28  
> **Version:** 0.1.0  
> **Last Review:** 2026-05-28

---

## Risk Matrix

| Probability \ Impact | Low (1) | Medium (2) | High (3) | Critical (4) |
|---------------------|---------|------------|----------|--------------|
| **Very High (5)** | 5 | 10 | 15 | 20 |
| **High (4)** | 4 | 8 | 12 | 16 |
| **Medium (3)** | 3 | 6 | 9 | 12 |
| **Low (2)** | 2 | 4 | 6 | 8 |
| **Very Low (1)** | 1 | 2 | 3 | 4 |

**Risk Score = Probability × Impact**

- **Critical (≥15):** Immediate action required
- **High (10–14):** Action plan needed within 1 week
- **Medium (6–9):** Monitor closely, action plan within 1 month
- **Low (≤5):** Accept and monitor

---

## Active Risks

### RISK-001: Credential Vault Compromise
| Field | Value |
|-------|-------|
| **ID** | RISK-001 |
| **Category** | Security |
| **Description** | Master password brute force or key extraction from memory leads to credential vault compromise |
| **Probability** | 3 (Medium) |
| **Impact** | 4 (Critical) |
| **Score** | 12 (High) |
| **Status** | Active |
| **Owner** | Security Lead |
| **Mitigation** | Argon2id with high memory cost; vault auto-lock; memory encryption; hardware security module (HSM) integration |
| **Contingency** | Incident response plan; user notification; forced password reset; audit log review |
| **Last Updated** | 2026-05-28 |

---

### RISK-002: Protocol Proxy Vulnerability
| Field | Value |
|-------|-------|
| **ID** | RISK-002 |
| **Category** | Security |
| **Description** | SSH/RDP/VNC proxy contains vulnerability allowing lateral movement or privilege escalation |
| **Probability** | 3 (Medium) |
| **Impact** | 4 (Critical) |
| **Score** | 12 (High) |
| **Status** | Active |
| **Owner** | Security Lead |
| **Mitigation** | Sandboxed subprocesses; seccomp/AppArmor; network namespace isolation; regular security scanning; penetration testing |
| **Contingency** | Immediate patch deployment; WAF rules; isolation of affected proxies |
| **Last Updated** | 2026-05-28 |

---

### RISK-003: WebSocket Scalability Bottleneck
| Field | Value |
|-------|-------|
| **ID** | RISK-003 |
| **Category** | Performance |
| **Description** | WebSocket terminal connections overload gateway server under high load |
| **Probability** | 4 (High) |
| **Impact** | 3 (High) |
| **Score** | 12 (High) |
| **Status** | Active |
| **Owner** | Backend Lead |
| **Mitigation** | Horizontal scaling with sticky sessions; connection pooling; load balancer with WebSocket support; rate limiting |
| **Contingency** | Circuit breaker; graceful degradation to HTTP polling; auto-scaling triggers |
| **Last Updated** | 2026-05-28 |

---

### RISK-004: Audit Log Tampering
| Field | Value |
|-------|-------|
| **ID** | RISK-004 |
| **Category** | Compliance |
| **Description** | Malicious actor modifies or deletes audit log entries to hide tracks |
| **Probability** | 2 (Low) |
| **Impact** | 4 (Critical) |
| **Score** | 8 (Medium) |
| **Status** | Active |
| **Owner** | Security Lead |
| **Mitigation** | Immutable log storage (WORM); hash chain integrity; append-only permissions; remote log shipping; digital signatures |
| **Contingency** | Forensic analysis; external audit; legal notification |
| **Last Updated** | 2026-05-28 |

---

### RISK-005: Mobile App Store Rejection
| Field | Value |
|-------|-------|
| **ID** | RISK-005 |
| **Category** | Business |
| **Description** | Apple App Store or Google Play Store rejects app due to SSH/terminal functionality |
| **Probability** | 3 (Medium) |
| **Impact** | 3 (High) |
| **Score** | 9 (Medium) |
| **Status** | Active |
| **Owner** | Product Manager |
| **Mitigation** | Early submission for review; comply with store guidelines; provide clear app description; enterprise distribution fallback |
| **Contingency** | Enterprise distribution (TestFlight, MDM); PWA fallback; direct APK distribution |
| **Last Updated** | 2026-05-28 |

---

### RISK-006: Desktop Code Signing Issues
| Field | Value |
|-------|-------|
| **ID** | RISK-006 |
| **Category** | Technical |
| **Description** | Apple/Microsoft code signing certificates delayed or rejected, blocking distribution |
| **Probability** | 3 (Medium) |
| **Impact** | 3 (High) |
| **Score** | 9 (Medium) |
| **Status** | Active |
| **Owner** | DevOps Lead |
| **Mitigation** | Early certificate application; notarization for macOS; EV code signing for Windows; CI/CD automation |
| **Contingency** | Unsigned builds for testing; self-signed certificates for internal use; delayed release |
| **Last Updated** | 2026-05-28 |

---

### RISK-007: Performance Regression Under Load
| Field | Value |
|-------|-------|
| **ID** | RISK-007 |
| **Category** | Performance |
| **Description** | System performance degrades significantly with 1000+ concurrent users |
| **Probability** | 4 (High) |
| **Impact** | 2 (Medium) |
| **Score** | 8 (Medium) |
| **Status** | Active |
| **Owner** | Backend Lead |
| **Mitigation** | Load testing at 2x expected capacity; connection pooling; caching; database query optimization; auto-scaling |
| **Contingency** | Graceful degradation; queue-based processing; temporary rate limiting; capacity alerts |
| **Last Updated** | 2026-05-28 |

---

### RISK-008: Dependency Vulnerability
| Field | Value |
|-------|-------|
| **ID** | RISK-008 |
| **Category** | Security |
| **Description** | Critical vulnerability discovered in Go, Rust, or Node.js dependency |
| **Probability** | 4 (High) |
| **Impact** | 2 (Medium) |
| **Score** | 8 (Medium) |
| **Status** | Active |
| **Owner** | Security Lead |
| **Mitigation** | Automated dependency scanning (Snyk, Dependabot); SBOM tracking; rapid patch process; minimal dependencies |
| **Contingency** | Immediate patching; temporary WAF rules; emergency deployment; vulnerability disclosure |
| **Last Updated** | 2026-05-28 |

---

### RISK-009: Data Loss During Migration
| Field | Value |
|-------|-------|
| **ID** | RISK-009 |
| **Category** | Technical |
| **Description** | Database migration fails or corrupts data during deployment |
| **Probability** | 2 (Low) |
| **Impact** | 4 (Critical) |
| **Score** | 8 (Medium) |
| **Status** | Active |
| **Owner** | Backend Lead |
| **Mitigation** | Migration testing on copy of production; backward-compatible changes; rollback scripts; backups before migration |
| **Contingency** | Database restore from backup; rollback to previous version; data reconciliation scripts |
| **Last Updated** | 2026-05-28 |

---

### RISK-010: Cross-Platform Compatibility Issues
| Field | Value |
|-------|-------|
| **ID** | RISK-010 |
| **Category** | Technical |
| **Description** | Desktop or mobile app behaves differently on specific OS versions or devices |
| **Probability** | 4 (High) |
| **Impact** | 2 (Medium) |
| **Score** | 8 (Medium) |
| **Status** | Active |
| **Owner** | QA Lead |
| **Mitigation** | Extensive device/OS matrix testing; automated E2E tests; beta testing program; feature flags for risky features |
| **Contingency** | Hotfix release; feature flag disable; workaround documentation; customer support escalation |
| **Last Updated** | 2026-05-28 |

---

### RISK-011: Team Member Departure
| Field | Value |
|-------|-------|
| **ID** | RISK-011 |
| **Category** | Business |
| **Description** | Key team member leaves during critical development phase |
| **Probability** | 3 (Medium) |
| **Impact** | 2 (Medium) |
| **Score** | 6 (Medium) |
| **Status** | Active |
| **Owner** | Project Manager |
| **Mitigation** | Knowledge documentation; pair programming; cross-training; bus factor > 1 for critical areas |
| **Contingency** | Contractor augmentation; feature deferral; extended timeline; recruitment acceleration |
| **Last Updated** | 2026-05-28 |

---

### RISK-012: WireGuard Kernel Module Unavailable

| Field | Value |
|-------|-------|
| **ID** | RISK-012 |
| **Category** | Technical |
| **Description** | Server lacks `wireguard.ko` kernel module (e.g., on older kernels, container environments, or cloud VMs without WG support) |
| **Probability** | 3 (Medium) |
| **Impact** | 4 (Critical) |
| **Score** | 12 (High) |
| **Status** | Active |
| **Owner** | DevOps Lead |
| **Mitigation** | Document kernel requirements; add `modprobe wireguard` check to health probes; provide fallback userspace WG impl (wireguard-go) for containers |
| **Contingency** | Switch to userspace WireGuard (wireguard-go); defer feature on non-WG hosts; provide clear error message to users |
| **Last Updated** | 2026-06-06 |

---

### RISK-013: WireGuard Key Rotation Connectivity Loss

| Field | Value |
|-------|-------|
| **ID** | RISK-013 |
| **Category** | Technical |
| **Description** | Key rotation drops active WireGuard connections, causing user frustration or data loss |
| **Probability** | 2 (Low) |
| **Impact** | 3 (High) |
| **Score** | 6 (Medium) |
| **Status** | Active |
| **Owner** | Backend Lead |
| **Mitigation** | Implement graceful rekey (hold old key briefly, apply new key without full interface restart); test rotation under active traffic |
| **Contingency** | Manual key rollback procedure; notify users of scheduled rotation windows; reconnect logic in clients |
| **Last Updated** | 2026-06-06 |

---

### RISK-014: WireGuard Subnet IP Exhaustion

| Field | Value |
|-------|-------|
| **ID** | RISK-014 |
| **Category** | Technical |
| **Description** | Configurable WireGuard subnet runs out of available IPs due to too many tunnels |
| **Probability** | 2 (Low) |
| **Impact** | 2 (Medium) |
| **Score** | 4 (Low) |
| **Status** | Active |
| **Owner** | Backend Lead |
| **Mitigation** | Configurable subnet size (default /24 = 253 IPs); rate limit of 10 tunnels/user + 1000 total per server; monitor IP pool usage with alert |
| **Contingency** | Expand subnet (e.g., /23 → /22); prune stale tunnels; add multi-subnet support |
| **Last Updated** | 2026-06-06 |

---

### RISK-015: Budget Overrun
| Field | Value |
|-------|-------|
| **ID** | RISK-015 |
| **Category** | Business |
| **Description** | Infrastructure or tooling costs exceed budget |
| **Probability** | 2 (Low) |
| **Impact** | 3 (High) |
| **Score** | 6 (Medium) |
| **Status** | Active |
| **Owner** | Project Manager |
| **Mitigation** | Cost monitoring dashboards; reserved instances; self-hosted runners; resource quotas; monthly budget reviews |
| **Contingency** | Resource optimization; feature deferral; cloud migration to cheaper provider; open-source alternatives |
| **Last Updated** | 2026-05-28 |

---

## Mitigated Risks

| ID | Risk | Mitigation Applied | Status |
|----|------|-------------------|--------|
| RISK-101 | Secret leak in CI | Pre-commit hooks, secret scanning, Vault integration | Mitigated |
| RISK-102 | Docker image vulnerabilities | Multi-stage builds, distroless runtime, Trivy scanning | Mitigated |
| RISK-103 | Single point of failure | HA PostgreSQL, Redis cluster, load balancer | Mitigated |

---

## Risk Monitoring

### Weekly Review Checklist
- [ ] Review new risks identified during sprint
- [ ] Update risk scores based on current data
- [ ] Verify mitigation actions are on track
- [ ] Escalate critical risks to stakeholders
- [ ] Document risk decisions in sprint retrospective

### Monthly Review Checklist
- [ ] Comprehensive risk register review
- [ ] Probability/impact reassessment
- [ ] Contingency plan testing
- [ ] Insurance/contract review
- [ ] Stakeholder risk communication

---

## Risk Trends

```
Score Trend (last 5 weeks):

Week 1: ████████████████████ (avg: 10.2)
Week 2: ██████████████████ (avg: 9.8)
Week 3: ████████████████ (avg: 8.5)
Week 4: ██████████████ (avg: 7.2)
Week 5: ████████████████ (avg: 8.8)

Trend: → Stable (new WG risks added, mitigations active)
```

---

## Appendix: Risk Assessment Methodology

### Probability Scale
| Score | Description |
|-------|-------------|
| 1 | Very Low (< 5% chance) |
| 2 | Low (5–20% chance) |
| 3 | Medium (20–50% chance) |
| 4 | High (50–80% chance) |
| 5 | Very High (> 80% chance) |

### Impact Scale
| Score | Description |
|-------|-------------|
| 1 | Low (minor inconvenience) |
| 2 | Medium (feature delay, < $10K cost) |
| 3 | High (sprint delay, $10K–$50K cost) |
| 4 | Critical (release delay, > $50K cost, reputation damage) |
