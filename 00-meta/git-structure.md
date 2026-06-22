# vexa — Git Structure

> Cara repo root dan application repo dihubungkan.  
> SSOT: Obsidian Vault `infra/vexa — Project Index.md`.

## Repositories

| Repo | URL | Purpose |
|------|-----|---------|
| Application | https://github.com/soumabali/vexa | Production code only |
| Root orchestrator | https://github.com/soumabali/vexa-root | Hermes+Claude context, plans, session notes |

## Local Layout

```text
/home/ubuntu/projects/vexa/          ← root git repo (main)
├── .git/                             ← single git root
├── 00-meta/                          ← SSOT pointers
├── 01-documents/                     ← mirrored docs
├── 02-application/                   ← git subtree (folder biasa, bukan repo terpisah)
├── 03-history/sessions/              ← session notes
├── 06-temp/plans/                    ← plans
├── .claude/
├── README.md
└── CLAUDE.md
```

**Catatan penting:** `02-application/` adalah **git subtree**, bukan repo lokal terpisah. Dia tidak punya `.git` sendiri.

## Push Workflow

### Commit perubahan di root

Semua perubahan di root maupun di `02-application/` di-commit dari root:

```bash
cd /home/ubuntu/projects/vexa
git add -A
git commit -m "type: description"
```

### Push root orchestrator

```bash
cd /home/ubuntu/projects/vexa
git push origin main
```

### Push application subtree

```bash
cd /home/ubuntu/projects/vexa
git subtree push --prefix=02-application https://github.com/soumabali/vexa.git main
```

### Push both

```bash
cd /home/ubuntu/projects/vexa
git push origin main
git subtree push --prefix=02-application https://github.com/soumabali/vexa.git main
```

## Why Subtree?

- Root tetap berisi plan, session notes, docs, dan pointer ke production code.
- Application repo tetap bersih hanya berisi code production.
- Kontributor open source clone `soumabali/vexa` tanpa melihat orchestrator context.

## Rules

- Jangan membuat `.git` baru di dalam `02-application/`.
- Jangan push dari dalam `02-application/` (tidak ada git context di sana).
- Selalu commit dari root, lalu push root + subtree.
