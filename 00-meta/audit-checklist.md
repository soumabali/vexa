# vexa — Audit Checklist

> Checklist untuk audit rutin workflow, dokumentasi, dan struktur repo.  
> Jalankan sebelum milestone besar atau setiap 2 minggu.

## Struktur Repo

- [ ] `02-application/` adalah git subtree, tidak punya `.git` sendiri.
- [ ] Root `.git/config` hanya punya remote valid:
  - `origin` → `soumabali/vexa-root`
  - `app-vexa` → `soumabali/vexa`
- [ ] `02-application/` di-track oleh root git (terlihat di `git status` saat ada perubahan).

## Dokumentasi Root

- [ ] `README.md` root up-to-date (link ke app repo, root repo, Obsidian, git structure).
- [ ] `CLAUDE.md` root up-to-date (role split, workflow, verification gates, skills).
- [ ] `00-meta/` lengkap: `urls.md`, `credentials-index.md`, `skills.md`, `git-structure.md`, `pre-push-checklist.md`.
- [ ] `01-documents/roadmap.md` adalah pointer ke Obsidian SSOT.
- [ ] `06-temp/plans/README.md` mencakup semua plan aktif dan arsip.
- [ ] `03-history/sessions/README.md` mencakup semua session notes.

## Dokumentasi Obsidian Vault

- [ ] `vexa — Project Index.md` up-to-date (commit hashes, links, git commands).
- [ ] `vexa Progress.md` tidak ada duplikasi section/header.
- [ ] `vexa Rules.md` penomoran section benar (tidak ada nomor duplikat).
- [ ] `vexa Workflow Runbook.md` mencakup Plan → Dispatch → Execute → Verify → Commit → Document.
- [ ] `vexa Claude Code Setup.md` mencakup push/pull subtree, skill mapping, verification gates.
- [ ] `vexa Runbook.md` mencakup backup volume, backup commands, troubleshooting.

## Application Context

- [ ] `02-application/CLAUDE.md` up-to-date (subtree note, skill mapping, verification gates).
- [ ] `02-application/.claude/settings.json` valid JSON, punya playwright MCP.
- [ ] `02-application/.claude/rules/hermes-skills.md` konsisten dengan root skill mapping.
- [ ] `02-application/README.md` open-source contributor version.

## CI/CD & Scripts

- [ ] Tidak ada workflow yang merujuk folder/file yang sudah tidak ada (K8s, Terraform, Helm).
- [ ] `scripts/rollback.sh` tidak merujuk K8s/Terraform.
- [ ] `docker-compose.prod.yml` valid dan mount `/backups`.

## Skills & MCP

- [ ] `~/.claude/skills/superpowers/` ada.
- [ ] `~/.claude/skills/caveman/` ada.
- [ ] `~/.claude/skills/graphify/` ada.
- [ ] `~/.claude/settings.json` punya playwright MCP.

## Backup & Infrastructure

- [ ] `/backups/vexa/db` ada dan writeable.
- [ ] `/backups/vexa/wireguard` ada dan writeable.
- [ ] `/backups/vexa/logs` ada dan writeable.
- [ ] Direktori backup di-mount ke container production.

## Automation / Prevention

- [ ] Jalankan `scripts/audit-vexa.sh` — harus exit 0.
- [ ] Pre-push checklist sudah di-review sebelum setiap push.
