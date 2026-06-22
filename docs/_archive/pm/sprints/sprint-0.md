# Sprint 0 — Foundation

> **Phase:** 0 | **Weeks:** 1–2 | **Duration:** 10 working days  
> **Sprint Goal:** Establish secure, scalable project foundation with CI/CD, monorepo structure, and team onboarding  
> **Story Points:** 55  
> **Status:** Not Started

---

## User Stories

### US-001: Monorepo Scaffolding
**As a** developer, **I want** a well-organized monorepo structure, **so that** all team members can work efficiently across backend, frontend, desktop, and mobile.

**Acceptance Criteria:**
- [ ] Directory structure follows `/apps/{api,web,desktop,mobile}`, `/packages/{ssh-core,ui,types,config}`, `/docs/`, `/infra/`
- [ ] Each app has its own `package.json`, `go.mod`, `Cargo.toml`, or `pubspec.yaml`
- [ ] Root `README.md` explains structure and getting started
- [ ] Root `.gitignore` ignores build artifacts, secrets, IDE files
- [ ] Makefile with common commands: `test`, `build`, `lint`, `docker-up`

**Definition of Done:**
- [ ] New team member can clone and run `make docker-up` successfully
- [ ] All apps build without errors
- [ ] CI pipeline passes on first commit

**Story Points:** 8

---

### US-002: CI/CD Pipeline Skeleton
**As a** DevOps engineer, **I want** automated build and test pipelines, **so that** every code change is validated before merge.

**Acceptance Criteria:**
- [ ] GitHub Actions workflow for Go backend (lint, test, build)
- [ ] GitHub Actions workflow for Next.js frontend (lint, test, build)
- [ ] GitHub Actions workflow for Rust desktop (lint, test, build)
- [ ] GitHub Actions workflow for Flutter mobile (lint, test, build)
- [ ] Docker multi-stage builds for backend and web
- [ ] Security scanning: `gosec`, `npm audit`, `cargo audit`
- [ ] Artifact storage for build outputs

**Definition of Done:**
- [ ] All workflows run on PR creation
- [ ] Failed workflow blocks merge
- [ ] Docker images pushed to registry on main branch

**Story Points:** 13

---

### US-003: Development Environment
**As a** developer, **I want** a consistent local development environment, **so that** "it works on my machine" is eliminated.

**Acceptance Criteria:**
- [ ] `docker-compose.yml` with PostgreSQL 16, Redis 7, Vault (dev mode)
- [ ] `.env.example` with all required environment variables
- [ ] Hot reload for Go backend (`air` or `fresh`)
- [ ] Hot reload for Next.js (`pnpm dev`)
- [ ] Hot reload for Tauri (`cargo tauri dev`)
- [ ] Database migrations run automatically on container start
- [ ] Seed data available for development

**Definition of Done:**
- [ ] `docker compose up` starts all services in < 60 seconds
- [ ] Developer can make code changes and see results within 5 seconds
- [ ] Documentation in `docs/dev/getting-started.md` is accurate

**Story Points:** 8

---

### US-004: Security Baseline
**As a** security engineer, **I want** foundational security controls, **so that** the project starts secure by default.

**Acceptance Criteria:**
- [ ] Secret scanning in CI (detect API keys, passwords in commits)
- [ ] Dependency vulnerability scanning (Snyk, Dependabot, or similar)
- [ ] Pre-commit hooks: `gofmt`, `eslint`, `prettier`, `rustfmt`
- [ ] Security headers middleware template
- [ ] CORS configuration for development
- [ ] Rate limiting stub (in-memory, configurable)

**Definition of Done:**
- [ ] Commit with hardcoded secret is rejected by pre-commit hook
- [ ] CI fails if `npm audit` finds critical vulnerabilities
- [ ] Security documentation in `docs/security/BASELINE.md`

**Story Points:** 13

---

### US-005: Team Onboarding Documentation
**As a** new team member, **I want** comprehensive onboarding docs, **so that** I can be productive within my first day.

**Acceptance Criteria:**
- [ ] `docs/dev/getting-started.md` with step-by-step setup
- [ ] `docs/dev/CONTRIBUTING.md` with coding standards and PR process
- [ ] `docs/dev/ARCHITECTURE.md` with system overview diagrams
- [ ] `docs/dev/DEPENDENCIES.md` with technology choices and rationale
- [ ] Troubleshooting guide for common issues
- [ ] IDE setup guide (VS Code, GoLand, IntelliJ)

**Definition of Done:**
- [ ] New hire completes setup in < 2 hours without asking for help
- [ ] Documentation is reviewed and approved by at least 2 team members
- [ ] All links in documentation are valid

**Story Points:** 13

---

## Dependencies

| Dependency | Source | Impact |
|------------|--------|--------|
| GitHub repository access | DevOps | Blocks CI/CD setup |
| Docker Hub / registry access | DevOps | Blocks image publishing |
| Domain name for staging | DevOps | Blocks SSL/TLS testing |
| Team member accounts | PM | Blocks access provisioning |

---

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation | Owner |
|------|-------------|--------|------------|-------|
| Docker performance issues on Windows/Mac | Medium | High | Provide native development fallback instructions | DevOps |
| Team unfamiliar with monorepo tools | Medium | Medium | Schedule 2-hour monorepo workshop in Week 1 | Tech Lead |
| CI minutes quota exceeded | Low | Medium | Use self-hosted runners for heavy jobs | DevOps |
| Secret leak in initial commits | Medium | Critical | Enable secret scanning before first push | Security |

---

## Sprint Review Checklist

- [ ] Monorepo structure reviewed by all team leads
- [ ] CI pipelines green for all apps
- [ ] Development environment tested by 3+ team members
- [ ] Security baseline approved by security engineer
- [ ] Onboarding docs tested by new hire (or simulated)
- [ ] Sprint retrospective scheduled and completed

---

## Sprint Retrospective Notes

*To be filled after sprint completion*

**What went well:**
- 

**What could be improved:**
- 

**Action items for next sprint:**
- 
