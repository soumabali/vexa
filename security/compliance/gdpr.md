# GDPR Compliance

## Overview

This document outlines GDPR compliance measures for vexa.

## Data Controller Information

| Field | Value |
|-------|-------|
| Controller | vexa Inc. |
| DPO | privacy@vexa.io |
| Address | [Company Address] |
| Jurisdiction | EU (GDPR applies to all users) |

## Legal Basis

### Article 6 - Lawfulness of Processing

| Purpose | Legal Basis | Consent Required |
|---------|------------|-----------------|
| User authentication | Contract performance | No |
| Session management | Contract performance | No |
| Connection storage | Contract performance | No |
| Credential vault | Contract performance | No |
| Security monitoring | Legitimate interest | No |
| Analytics | Legitimate interest | No |
| Marketing | Consent | Yes |
| Product improvements | Consent | Yes |

## Data Categories

### Personal Data Processed

| Category | Data | Retention |
|----------|------|-----------|
| Identity | Email, display name | Account lifetime |
| Authentication | WebAuthn credentials, API keys | Account lifetime |
| Technical | IP address, user agent | 90 days |
| Usage | Session logs, activity | 90 days |
| Preferences | Settings, theme | Account lifetime |

### Special Categories

| Category | Status | Reason |
|----------|--------|--------|
| Biometric data | No | WebAuthn uses local processing |
| Health data | No | Not collected |
| Financial data | No | Not collected |

## Data Subject Rights

### Article 12-14 - Transparency

| Requirement | Implementation |
|-------------|---------------|
| Privacy notice | `apps/web/app/privacy/page.tsx` |
| Cookie notice | Cookie consent banner |
| Data collection notice | Registration form |

### Article 15 - Right of Access

| Requirement | Implementation |
|-------------|---------------|
| Self-service export | Settings → Export data |
| Manual request | privacy@vexa.io |
| Response time | 30 days |

### Article 16 - Right to Rectification

| Requirement | Implementation |
|-------------|---------------|
| Self-service | Settings → Profile |
| Manual request | privacy@vexa.io |

### Article 17 - Right to Erasure

| Requirement | Implementation |
|-------------|---------------|
| Self-service | Settings → Delete account |
| Manual request | privacy@vexa.io |
| Data removal | Cascading delete |

### Article 18 - Right to Restriction

| Requirement | Implementation |
|-------------|---------------|
| Account suspension | Settings → Suspend |
| Processing restriction | Data anonymization |

### Article 20 - Right to Data Portability

| Requirement | Implementation |
|-------------|---------------|
| Export format | JSON, CSV |
| Self-service | Settings → Export |

### Article 21 - Right to Object

| Requirement | Implementation |
|-------------|---------------|
| Marketing opt-out | Unsubscribe link |
| Profiling | No profiling performed |

### Article 22 - Automated Decision Making

| Requirement | Implementation |
|-------------|---------------|
| No automated decisions | No profiling or automated decisions |

## Technical Measures

### Article 25 - Data Protection by Design

| Measure | Implementation |
|---------|---------------|
| Privacy by design | Architecture review |
| Default settings | Minimal data collection |
| Data minimization | Only necessary data |
| Purpose limitation | Clear data use |

### Article 32 - Security

| Measure | Implementation |
|---------|---------------|
| Encryption at rest | AES-256-GCM |
| Encryption in transit | TLS 1.3 |
| Access control | RBAC |
| Logging | Audit logs |
| Testing | Regular pentests |

### Article 33 - Breach Notification

| Requirement | Implementation |
|-------------|---------------|
| Detection | Monitoring, alerts |
| Assessment | Severity evaluation |
| Supervisory authority | 72-hour notification |
| Data subjects | Without undue delay |

### Article 34 - Data Subject Notification

| Requirement | Implementation |
|-------------|---------------|
| High risk breaches | Direct notification |
| Communication method | Email, in-app |
| Content | Nature, measures, contact |

## Data Processing Agreement

### Article 28 - Processor

| Field | Value |
|-------|-------|
| Processor | vexa Inc. |
| Sub-processors | Cloud provider, payment processor |
| Security measures | See Article 32 |
| Audit rights | Annual audit |
| Data location | EU/US (with SCCs) |

## International Transfers

### Article 45-49 - Transfers

| Transfer | Mechanism |
|----------|-----------|
| EU → US | Standard Contractual Clauses |
| EU → UK | Adequacy decision |
| EU → Other | SCCs + DPA |

## Data Retention

| Data Type | Retention Period | Deletion Method |
|-----------|-----------------|----------------|
| Account data | Account lifetime | Cascading delete |
| Session logs | 90 days | Automated purge |
| Audit logs | 1 year | Automated purge |
| Backups | 30 days | Encrypted, separate keys |
| Analytics | 1 year | Anonymization |

## DPIA (Data Protection Impact Assessment)

### High-Risk Processing

| Processing | Risk Level | DPIA Required |
|------------|-----------|-------------|
| Biometric auth | High | Yes |
| Credential storage | High | Yes |
| Session recording | Medium | Yes |
| Basic auth | Low | No |

## Records of Processing

| Activity | Purpose | Data Subjects | Recipients | Retention |
|----------|---------|--------------|-----------|-----------|
| User registration | Account creation | Users | Internal | Account lifetime |
| Authentication | Access control | Users | Internal | Account lifetime |
| Session management | Service provision | Users | Internal | 90 days |
| Credential storage | Feature provision | Users | Internal | Account lifetime |
| Security monitoring | Security | Users | Security team | 90 days |
| Analytics | Improvement | Users | Analytics team | 1 year |

## Cookie Policy

| Cookie | Purpose | Type | Duration |
|--------|---------|------|----------|
| session | Authentication | Necessary | Session |
| preferences | Settings | Functional | 1 year |
| analytics | Analytics | Performance | 1 year |
| marketing | Marketing | Advertising | 1 year |

## Consent Management

### Consent Requirements

| Activity | Required | Method | Withdrawal |
|----------|----------|--------|-----------|
| Marketing | Yes | Checkbox | Settings or email |
| Analytics | Yes | Banner | Cookie settings |
| Product feedback | Yes | Banner | Cookie settings |

### Consent Record

| Field | Value |
|-------|-------|
| Timestamp | ISO 8601 |
| Version | Policy version |
| Method | Checkbox, banner |
| Withdrawal | Available anytime |

## Contact

### Data Protection Officer

| Method | Contact |
|--------|---------|
| Email | privacy@vexa.io |
| Address | [Company Address] |
| Response | 30 days |

### Supervisory Authority

| Jurisdiction | Authority |
|--------------|-----------|
| EU | Local DPA |
| Germany | BfDI |
| France | CNIL |
| UK | ICO |

## Next Steps

1. Complete DPIA
2. Implement consent management
3. Create data subject request workflow
4. Set up breach notification procedure
5. Train team on GDPR
6. Conduct privacy audit
7. Update privacy notice
8. Register with supervisory authority (if required)
