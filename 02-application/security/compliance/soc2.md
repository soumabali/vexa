# SOC 2 Compliance Mapping

## Overview

This document maps vexa controls to SOC 2 Trust Service Criteria.

## Trust Service Criteria

### CC1.0 - Control Environment

| Control | Implementation | Evidence |
|---------|---------------|----------|
| CC1.1 - Integrity and Ethical Values | Code of Conduct, Security Policy | `docs/policies/code-of-conduct.md` |
| CC1.2 - Board Independence | N/A (startup) | - |
| CC1.3 - Management Override | Approval workflows | GitHub branch protection |
| CC1.4 - Accountability | Role-based access control | RBAC implementation |
| CC1.5 - Authority and Responsibility | Org chart, job descriptions | `docs/org-chart.md` |

### CC2.0 - Communication and Information

| Control | Implementation | Evidence |
|---------|---------------|----------|
| CC2.1 - Information Systems | Logging, monitoring | Audit logs, metrics |
| CC2.2 - Internal Communication | Slack, email, meetings | Meeting notes |
| CC2.3 - External Communication | Status page, support | `status.vexa.io` |

### CC3.0 - Risk Assessment

| Control | Implementation | Evidence |
|---------|---------------|----------|
| CC3.1 - Risk Identification | Risk register | `security/risk-register.md` |
| CC3.2 - Fraud Risk | Fraud detection | Rate limiting, anomaly detection |
| CC3.3 - Technology Risk | Dependency scanning | Snyk, Dependabot |
| CC3.4 - Risk Mitigation | Security controls | WAF, rate limiting, encryption |

### CC4.0 - Monitoring Activities

| Control | Implementation | Evidence |
|---------|---------------|----------|
| CC4.1 - Monitoring Activities | Metrics, alerts | Prometheus, PagerDuty |
| CC4.2 - Internal Audit | Security audits | Quarterly pentests |

### CC5.0 - Control Activities

| Control | Implementation | Evidence |
|---------|---------------|----------|
| CC5.1 - Control Selection | Risk-based controls | Control matrix |
| CC5.2 - General IT Controls | Access control, change management | GitOps, RBAC |
| CC5.3 - Application Controls | Input validation, output encoding | WAF, sanitization |

### CC6.0 - Logical and Physical Access

| Control | Implementation | Evidence |
|---------|---------------|----------|
| CC6.1 - Logical Access Security | Authentication, authorization | WebAuthn, RBAC |
| CC6.2 - Access Removal | Offboarding process | `docs/offboarding.md` |
| CC6.3 - Access Modification | Role changes | Audit logs |
| CC6.4 - Access Reviews | Quarterly access reviews | `security/access-reviews/` |
| CC6.5 - MFA | WebAuthn, TOTP | `apps/web/app/settings/security/` |
| CC6.6 - Encryption | TLS 1.3, AES-256-GCM | `apps/api/internal/security/hardening.go` |
| CC6.7 - Encryption Keys | Vault, KMS | HashiCorp Vault |

### CC7.0 - System Operations

| Control | Implementation | Evidence |
|---------|---------------|----------|
| CC7.1 - Detection | Intrusion detection | Falco, audit logs |
| CC7.2 - Incident Response | IR plan | `docs/incident-response.md` |
| CC7.3 - Mitigation | Patch management | Dependabot |
| CC7.4 - Backup | Backup strategy | `docs/backup.md` |
| CC7.5 - Recovery | Disaster recovery | `docs/disaster-recovery.md` |

### CC8.0 - Change Management

| Control | Implementation | Evidence |
|---------|---------------|----------|
| CC8.1 - Change Authorization | PR reviews | GitHub PRs |
| CC8.2 - Testing | CI/CD testing | GitHub Actions |
| CC8.3 - Implementation | GitOps | ArgoCD |
| CC8.4 - Rollback | Automated rollback | `apps/api/internal/health/` |

### A1.0 - Availability

| Control | Implementation | Evidence |
|---------|---------------|----------|
| A1.1 - Uptime | 99.9% SLA | Status page |
| A1.2 - Monitoring | Health checks | `apps/api/internal/health/` |
| A1.3 - Capacity | Auto-scaling | HPA, VPA |

### C1.0 - Confidentiality

| Control | Implementation | Evidence |
|---------|---------------|----------|
| C1.1 - Data Classification | Classification policy | `docs/data-classification.md` |
| C1.2 - Encryption | At-rest, in-transit | TLS 1.3, Vault |
| C1.3 - Access Control | Need-to-know | RBAC |

### PI1.0 - Processing Integrity

| Control | Implementation | Evidence |
|---------|---------------|----------|
| PI1.1 - Input Processing | Validation | `apps/api/internal/validation/` |
| PI1.2 - Output Processing | Encoding | `apps/web/lib/utils.ts` |

### P1.0 - Privacy

| Control | Implementation | Evidence |
|---------|---------------|----------|
| P1.1 - Notice | Privacy policy | `apps/web/app/privacy/` |
| P1.2 - Choice | Consent management | Cookie consent |
| P1.3 - Collection | Data minimization | Only collect necessary data |
| P1.4 - Use | Purpose limitation | `docs/privacy.md` |
| P1.5 - Retention | Retention policy | `docs/retention.md` |
| P1.6 - Disclosure | Third-party sharing | `docs/third-parties.md` |
| P1.7 - Quality | Data accuracy | User profile updates |
| P1.8 - Security | Security measures | All security controls |
| P1.9 - Monitoring | Privacy monitoring | Audit logs |

## Evidence Collection

### Automated Evidence

| Source | Evidence | Frequency |
|--------|----------|-----------|
| GitHub | PR reviews, code changes | Continuous |
| CI/CD | Test results, deployments | Per build |
| Monitoring | Uptime, alerts | Continuous |
| Security | Scan results, pentests | Quarterly |

### Manual Evidence

| Source | Evidence | Frequency |
|--------|----------|-----------|
| Management | Risk assessments | Quarterly |
| HR | Access reviews | Quarterly |
| Legal | Contracts, policies | As needed |

## Auditor Questions

### Sample Questions

1. How do you authenticate users?
2. How do you authorize access to resources?
3. How do you log and monitor access?
4. How do you encrypt data?
5. How do you manage changes?
6. How do you respond to incidents?
7. How do you ensure availability?
8. How do you protect privacy?

### Expected Answers

See corresponding sections above.

## Gap Analysis

### Current Gaps

| Gap | Priority | Remediation | Timeline |
|-----|----------|-------------|----------|
| Formal risk register | High | Create `security/risk-register.md` | Q1 |
| Quarterly access reviews | High | Implement access review process | Q1 |
| Formal incident response | High | Create `docs/incident-response.md` | Q1 |
| Disaster recovery testing | Medium | DR drill | Q2 |
| Third-party assessments | Medium | Vendor security reviews | Q2 |

## Next Steps

1. Create risk register
2. Implement access review process
3. Create incident response plan
4. Conduct DR drill
5. Schedule quarterly pentests
6. Document all policies
7. Train team on SOC 2
8. Engage auditor
