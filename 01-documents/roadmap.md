# vexa Roadmap

> Single source of truth for project direction. Mirror of Obsidian Vault note.

## Repositories

- **Application (public):** https://github.com/soumabali/vexa
- **Orchestrator root (private):** https://github.com/soumabali/vexa-root

## Active Phase

### Phase 1 — Open-Source Cleanup ✅
- Rebrand to `vexa — Complete SSH Manager`
- Rename internal packages `@ssh-manager/*` → `@vexa/*`
- Update DB defaults, sandbox identifiers, docs
- Remove legacy infra/docker and infra/github-actions.yml

## Next Phases

| Phase | Goal | Key Deliverables |
|-------|------|-----------------|
| P2 | Core SSH Features | Real SSH terminal verify, SFTP manager, SSH key integration, session logs |
| P3 | Security & Enterprise | TOTP MFA real, credential sharing, audit trail, RBAC |
| P4 | Tunnels Polish | Live stats, WG config download, enable/disable UI |
| P5 | Desktop & Mobile | Tauri/Flutter login + host list |
| P6 | Developer Experience | `make dev`, GitHub Actions CI, contribution guide |

## Workflow

1. Ame writes plan in `06-temp/plans/`.
2. Claude Code executes inside `02-application/`.
3. Ame runs verification gates.
4. Independent Claude Code review for non-trivial changes.
5. Ame commits, pushes, and writes session note.
