# vexa — Security Runbook

> **Agent:** DevOps + Security Engineer  
> **Status:** Generated 2026-05-28  
> **Version:** 0.1.0  
> **Classification:** Confidential

---

## 1. Incident Response

### 1.1 Incident Severity Levels

| Level | Description | Response Time | Escalation |
|-------|-------------|---------------|------------|
| **Critical** | Active breach, data loss, system down | 15 minutes | CEO, CTO, Legal |
| **High** | Potential breach, critical vulnerability | 1 hour | CTO, Security Lead |
| **Medium** | Security event, policy violation | 4 hours | Security Lead |
| **Low** | Minor issue, false positive | 24 hours | DevOps Lead |

### 1.2 Incident Response Flow

```
1. DETECT
   └── Alert fires (monitoring, user report, audit)
   
2. TRIAGE
   └── Assess severity
   └── Notify on-call engineer
   
3. CONTAIN
   └── Isolate affected systems
   └── Block malicious IPs
   └── Revoke compromised credentials
   
4. ERADICATE
   └── Remove attacker access
   └── Patch vulnerability
   └── Clean compromised data
   
5. RECOVER
   └── Restore from clean backup
   └── Verify system integrity
   └── Resume normal operations
   
6. POST-INCIDENT
   └── Document timeline
   └── Conduct post-mortem
   └── Implement preventive measures
```

### 1.3 Emergency Contacts

| Role | Contact | Phone |
|------|---------|-------|
| Security Lead | security@vexa.local | +1-XXX-XXX-XXXX |
| CTO | cto@vexa.local | +1-XXX-XXX-XXXX |
| On-Call Engineer | oncall@vexa.local | PagerDuty |
| Legal | legal@vexa.local | +1-XXX-XXX-XXXX |

---

## 2. Security Monitoring

### 2.1 Critical Alerts

| Alert | Condition | Action |
|-------|-----------|--------|
| **Brute Force** | > 5 failed logins in 1 min | Block IP, alert security |
| **Privilege Escalation** | Admin action from non-admin | Alert, audit review |
| **Data Exfiltration** | > 1GB outbound in 1 min | Block, investigate |
| **Unauthorized Access** | Login from new country | Require MFA, alert user |
| **Audit Log Tampering** | Integrity check fails | Critical alert, investigation |
| **Certificate Expiry** | < 7 days until expiry | Renew immediately |
| **Vulnerability Detected** | CVSS > 7.0 | Patch within 24h |

### 2.2 Monitoring Dashboards

- **Security Overview:** `https://grafana.vexa.local/d/security`
- **Audit Log:** `https://grafana.vexa.local/d/audit`
- **Threat Detection:** `https://grafana.vexa.local/d/threats`
- **Vulnerability Scan:** `https://grafana.vexa.local/d/vulns`

---

## 3. Vulnerability Management

### 3.1 Scanning Schedule

| Scan Type | Frequency | Tool |
|-----------|-----------|------|
| Dependency scan | Daily | Snyk, Dependabot |
| Container scan | Every build | Trivy |
| Infrastructure scan | Weekly | ScoutSuite |
| Penetration test | Quarterly | External firm |
| Code review | Every PR | SonarQube |

### 3.2 Vulnerability Response SLA

| CVSS Score | Severity | Response Time | Fix Time |
|------------|----------|---------------|----------|
| 9.0–10.0 | Critical | 1 hour | 24 hours |
| 7.0–8.9 | High | 4 hours | 7 days |
| 4.0–6.9 | Medium | 24 hours | 30 days |
| 0.1–3.9 | Low | 7 days | 90 days |

### 3.3 Patch Process

```bash
# 1. Identify affected systems
kubectl get pods -l app=vexa

# 2. Create patch branch
git checkout -b security/patch-CVE-2026-XXXX

# 3. Update dependency
cd apps/api && go get package@patched-version

# 4. Test
make test

# 5. Deploy to staging
make deploy-staging

# 6. Verify
make smoke-test

# 7. Deploy to production (canary)
make deploy-canary

# 8. Monitor
make monitor

# 9. Full rollout
make deploy-production
```

---

## 4. Access Control

### 4.1 Production Access

| Role | Access | Approval Required |
|------|--------|-------------------|
| DevOps Lead | Full | None |
| Security Lead | Read + Audit | CTO |
| Backend Lead | Read + Debug | DevOps Lead |
| Developer | None | DevOps Lead |

### 4.2 Emergency Access

```bash
# Break-glass access (requires approval)
ssh breakglass@vexa.prod
# Authenticates via hardware token + peer approval

# All break-glass sessions are:
# - Recorded (screen + keystrokes)
# - Alerted to security team
# - Limited to 2 hours
# - Requires post-incident review
```

### 4.3 Credential Rotation

| Credential Type | Rotation Frequency | Method |
|-----------------|-------------------|--------|
| Database password | 90 days | Terraform + Vault |
| API keys | 90 days | Vault dynamic secrets |
| TLS certificates | 60 days | cert-manager |
| SSH keys | 180 days | Vault SSH CA |
| JWT signing keys | 180 days | Automated rotation |

---

## 5. Backup & Recovery

### 5.1 Backup Schedule

| Data | Frequency | Retention | Location |
|------|-----------|-----------|----------|
| PostgreSQL | Hourly (incremental) | 7 days | S3 (encrypted) |
| PostgreSQL | Daily (full) | 30 days | S3 (encrypted) |
| PostgreSQL | Weekly | 1 year | Glacier |
| Redis | Daily (RDB) | 7 days | S3 |
| Audit logs | Real-time | 7 years | S3 + Glacier |
| Config files | On change | 1 year | Git |

### 5.2 Recovery Procedures

#### Database Recovery

```bash
# Point-in-time recovery
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance vexa-prod \
  --target-db-instance vexa-recovery \
  --restore-time 2026-05-28T12:00:00Z

# Verify data integrity
psql -h vexa-recovery -c "SELECT COUNT(*) FROM users;"

# Promote to primary
aws rds promote-read-replica \
  --db-instance-identifier vexa-recovery
```

#### Full System Recovery

```bash
# 1. Provision new infrastructure
terraform apply -var="environment=disaster-recovery"

# 2. Restore database
make restore-database BACKUP_DATE=2026-05-28

# 3. Restore configuration
make restore-config

# 4. Verify health
make health-check

# 5. Update DNS
make update-dns

# RTO: 1 hour
# RPO: 15 minutes
```

---

## 6. Compliance

### 6.1 GDPR Compliance

| Requirement | Implementation |
|-------------|---------------|
| Data Processing Agreement | `docs/legal/DPA.md` |
| Right to Access | API endpoint: `GET /users/me/export` |
| Right to Erasure | API endpoint: `DELETE /users/me` |
| Data Portability | JSON export |
| Breach Notification | 72 hours to authorities |

### 6.2 SOC 2 Controls

| Control | Evidence | Frequency |
|---------|----------|-----------|
| Access reviews | Quarterly access review report | Quarterly |
| Penetration tests | Pen test report | Annually |
| Vulnerability scans | Scan results | Monthly |
| Change management | PR approvals, deployment logs | Continuous |
| Backup testing | Recovery test results | Quarterly |

### 6.3 Audit Trail

All security-relevant events are logged:

```json
{
  "timestamp": "2026-05-28T12:00:00Z",
  "event_type": "security_alert",
  "severity": "high",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "ip_address": "192.168.1.100",
  "action": "failed_login",
  "details": {
    "attempts": 5,
    "time_window": "60s"
  },
  "integrity_hash": "sha256:abc123..."
}
```

---

## 7. Security Tools Reference

### 7.1 Scanning Tools

| Tool | Purpose | Command |
|------|---------|---------|
| Trivy | Container scanning | `trivy image vexa:latest` |
| Snyk | Dependency scanning | `snyk test` |
| Gosec | Go security | `gosec ./...` |
| OWASP ZAP | Web app scanning | `zap-baseline.py -t http://localhost:8080` |
| ScoutSuite | Cloud security | `scoutSuite aws` |

### 7.2 Monitoring Tools

| Tool | Purpose | URL |
|------|---------|-----|
| Prometheus | Metrics | `http://prometheus.vexa.local` |
| Grafana | Dashboards | `http://grafana.vexa.local` |
| Loki | Logs | `http://loki.vexa.local` |
| Jaeger | Tracing | `http://jaeger.vexa.local` |
| Sentry | Errors | `http://sentry.vexa.local` |

### 7.3 Incident Response Tools

| Tool | Purpose | Command |
|------|---------|---------|
| Wireshark | Packet capture | `tshark -i eth0 -w incident.pcap` |
| Volatility | Memory forensics | `volatility -f memory.dmp linux_pslist` |
| ELK | Log analysis | `kibana.vexa.local` |

---

## 8. Runbooks

### 8.1 Credential Vault Compromise

```bash
# 1. Lock all vaults
kubectl exec deploy/vexa -- /app/lock-all-vaults

# 2. Rotate master keys
kubectl exec deploy/vexa -- /app/rotate-master-keys

# 3. Force password reset for all users
kubectl exec deploy/vexa -- /app/force-password-reset

# 4. Audit review
make audit-review START=2026-05-28T00:00:00Z

# 5. Notify users
make notify-users TEMPLATE=vault-compromise
```

### 8.2 DDoS Attack

```bash
# 1. Activate Cloudflare DDoS protection
cloudflare ddos protect vexa.local

# 2. Rate limit at nginx
kubectl apply -f k8s/rate-limit-emergency.yaml

# 3. Block malicious IPs
iptables -A INPUT -s MALICIOUS_IP -j DROP

# 4. Scale up
kubectl scale deployment vexa --replicas=20

# 5. Monitor
watch -n 1 'kubectl top pods'
```

### 8.3 Database Breach

```bash
# 1. Isolate database
aws rds modify-db-instance \
  --db-instance-identifier vexa-prod \
  --vpc-security-group-ids sg-isolated

# 2. Create forensic snapshot
aws rds create-db-snapshot \
  --db-instance-identifier vexa-prod \
  --db-snapshot-identifier forensic-2026-05-28

# 3. Audit access logs
make audit-database-access

# 4. Reset all credentials
make rotate-all-credentials

# 5. Notify affected users
make notify-breach
```

---

## 9. Security Checklists

### 9.1 Pre-Deployment Security Checklist

- [ ] All dependencies scanned (no critical/high vulnerabilities)
- [ ] Container image scanned (Trivy clean)
- [ ] Secrets scan clean (no hardcoded credentials)
- [ ] TLS certificates valid (> 30 days)
- [ ] Security headers configured
- [ ] Rate limiting enabled
- [ ] Audit logging enabled
- [ ] Backup verified
- [ ] Rollback plan documented
- [ ] On-call schedule active

### 9.2 Quarterly Security Review

- [ ] Access review (remove terminated employees)
- [ ] Certificate expiry check
- [ ] Password policy compliance
- [ ] MFA enrollment rate (> 90%)
- [ ] Penetration test review
- [ ] Vulnerability scan review
- [ ] Incident response drill
- [ ] Backup recovery test
- [ ] Disaster recovery test
- [ ] Security training completion

---

## 10. Contact Information

### Security Team

| Name | Role | Email | PagerDuty |
|------|------|-------|-----------|
| Security Lead | Security Engineering | security@vexa.local | security-oncall |
| DevOps Lead | Infrastructure | devops@vexa.local | devops-oncall |
| Backend Lead | Application Security | backend@vexa.local | backend-oncall |

### External Contacts

| Service | Contact | Escalation |
|---------|---------|------------|
| Cloud Provider | support@aws.com | Account Manager |
| CDN | support@cloudflare.com | Enterprise Support |
| Certificate Authority | support@digicert.com | Priority Support |
| Security Firm | contact@security-firm.com | Principal Consultant |

---

## Appendix A: Security Incident Report Template

```markdown
# Security Incident Report

**Incident ID:** INC-2026-XXXX
**Date:** 2026-05-28
**Severity:** Critical/High/Medium/Low
**Reporter:** Name

## Summary
Brief description of the incident

## Timeline
- 12:00 - Event detected
- 12:05 - Initial response
- 12:30 - Containment completed
- 13:00 - Investigation started

## Impact
- Affected systems:
- Data exposure:
- User impact:

## Root Cause
Detailed analysis of cause

## Resolution
Steps taken to resolve

## Lessons Learned
What went well / what could be improved

## Preventive Measures
Actions to prevent recurrence

## Sign-off
- [ ] Security Lead
- [ ] CTO
- [ ] Legal (if required)
```

## Appendix B: Security Metrics Dashboard

| Metric | Target | Current | Trend |
|--------|--------|---------|-------|
| Vulnerability patch time (critical) | < 24h | 18h | ↓ |
| MFA enrollment rate | > 90% | 87% | ↑ |
| Security incidents | < 1/month | 2 | → |
| Failed login rate | < 1% | 0.5% | ↓ |
| Audit log coverage | 100% | 100% | → |
