# vexa Workflow — Target 10/10

> Flow kerja ideal untuk project vexa yang memberikan rating 10/10:  
003e otomatis, terverifikasi, tanpa kebocoran, dan semua agent patuh.

---

## Flowchart: Ideal Workflow

```text
┌─────────────────┐
│ 0. INTAKE       │  Dhar atau trigger otomatis meminta task
│    (Ame)        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ 1. AUDIT        │────▶│ ❌ FAIL         │
│    (Ame)        │     │ Jalankan audit  │
│ audit-vexa.sh   │     │ perbaiki dulu   │
└────────┬────────┘     └─────────────────┘
         │ ✅ PASS
         ▼
┌─────────────────┐
│ 2. PLAN         │  Tulis plan: 06-temp/plans/YYYYMMDD-<slug>.md
│    (Ame)        │  Scope, files, acceptance criteria, rollback
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 3. DISPATCH     │  Claude Code dengan compliance prompt
│    (Ame)        │  --output-format json dengan schema
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ 4. COMPLIANCE   │────▶│ ❌ FAIL         │
│    (Claude/Tool)│     │ Hentikan task,  │
│ 10-point check  │     │ laporkan Ame    │
└────────┬────────┘     └─────────────────┘
         │ ✅ PASS
         ▼
┌─────────────────┐
│ 5. EXECUTE      │  Claude Code coding
│    (Claude Code)│  Skill stack: superpowers, caveman, graphify
└────────┬────────┘     (+ playwright MCP untuk E2E)
         │
         ▼
┌─────────────────┐
│ 6. VERIFY       │  go test, go build, npm build, E2E
│    (Ame)        │  Copy static → standalone
└────────┬────────┘
         │ ✅ PASS
         ▼
┌─────────────────┐     ┌─────────────────┐
│ 7. REVIEW       │────▶│ ❌ CHANGES      │
│    (Claude Code)│     │ Perbaiki ulang  │
│ Pre-commit      │     │ loop ke step 5  │
└────────┬────────┘     └─────────────────┘
         │ ✅ PASS
         ▼
┌─────────────────┐
│ 8. STAGE CLEAN  │  Cek git status, buang generated files
│    (Ame)        │  next-env.d.ts, .auth/, .next/, dll
└────────┬────────┘
         │ ✅ CLEAN
         ▼
┌─────────────────┐
│ 9. COMMIT APP   │  Dari root, commit ke subtree
│    (Ame)        │  git subtree push --prefix=02-application
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 10. PUSH ROOT   │  git push origin main → vexa-root
│    (Ame)        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 11. DEPLOY      │  docker compose -f docker-compose.prod.yml up -d
│    (Ame)        │  Hanya jika perubahan deployment file
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 12. E2E PROD    │  SKIP_WEBSERVER=true, BASE_URL production
│    (Ame)        │  Target: 19/19 passing
└────────┬────────┘
         │ ✅ 19/19
         ▼
┌─────────────────┐
│ 13. DOCUMENT    │  Session note + Obsidian update
│    (Ame)        │  new-session.sh otomatis
└─────────────────┘
```

---

## Gate pada Setiap Step

| Step | Gate | Tool/File | Apa yang Dicek |
|------|------|-----------|----------------|
| 1 | Audit Gate | `audit-vexa.sh` | Struktur, remote, skills, backup, context files |
| 4 | Compliance Gate | `claude-compliance-check.sh` | 10-point confirmation dari Claude Code |
| 6 | Verification Gate | go test, go build, npm build | Code quality & build |
| 7 | Review Gate | Claude Code pre-commit review | Security, quality, best practice |
| 8 | Stage Clean Gate | `git status --short` | Tidak ada generated files |
| 9 | Subtree Gate | `git subtree push` | Application repo updated |
| 10 | Root Push Gate | `git push origin main` | Root repo updated |
| 12 | Production Gate | Playwright E2E | 19/19 passing |

---

## Perbedaan Rating 8/10 vs 10/10

| Aspek | 8/10 Sekarang | 10/10 Target |
|-------|---------------|--------------|
| Compliance | Self-reported | Tool-validated (JSON schema) |
| Audit | Manual script | Pre-push hook + CI gate |
| Response Claude | Free text | Structured JSON 10-point confirmation |
| Backup | Volume di-mount | Automation harian (cron/container) |
| Session note | Manual write | `new-session.sh` generator |
| Progress tracking | Manual update | `vexa-status.sh` dashboard otomatis |
| Error handling | Ad-hoc | Error Recovery Playbook |
| Skill verification | Check file exists | Parse Claude response + runtime test |

---

## Acceptance Criteria untuk 10/10

- [ ] `scripts/claude-compliance-check.sh` exit 0 setelah dry-run.
- [ ] `scripts/audit-vexa.sh` mencakup compliance check.
- [ ] Pre-push hook terpasang dan memblokir push yang gagal audit.
- [ ] `00-meta/claude-response-schema.json` ada dan digunakan.
- [ ] Backup automation berjalan (minimal PostgreSQL harian).
- [ ] `scripts/vexa-status.sh` menampilkan snapshot health/repo.
- [ ] `scripts/new-session.sh` generator otomatis.
- [ ] Obsidian `vexa Error Recovery Playbook.md` tersedia.

---

## Files & Links

- Active Plan: `06-temp/plans/20260622-workflow-improvement-10of10.md`
- Session Note: `03-history/sessions/2026-06-22-workflow-dryrun-test.md`
- Audit Script: `scripts/audit-vexa.sh`
- Hard Rules: `AGENTS.md`
- Skill Mapping: `00-meta/skills.md`
- Obsidian: `infra/vexa Workflow Runbook.md`
- Obsidian: `infra/vexa Rules.md`
