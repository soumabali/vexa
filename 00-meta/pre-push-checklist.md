# vexa — Pre-Push Checklist

> Checklist wajib sebelum push root atau application subtree.  
003e SSOT: Obsidian Vault `infra/vexa Workflow Runbook.md`.

## Sebelum Push Root Orchestrator

- [ ] Semua perubahan root context sudah di-add.
- [ ] Tidak ada file kredensial yang ikut di-commit (`.env`, `*.pem`, `*.key`).
- [ ] Session note sudah ditulis/akan ditulis setelah push.
- [ ] `git remote -v` hanya menampilkan:
  - `origin` → `https://github.com/soumabali/vexa-root.git`
  - `app-vexa` → `https://github.com/soumabali/vexa.git`
- [ ] Commit message mengikuti conventional commits.

## Sebelum Push Application Subtree

- [ ] Verification gates sudah hijau:
  - [ ] `cd 02-application/apps/api && go test ./...`
  - [ ] `cd 02-application/apps/api && go build ./...`
  - [ ] `cd 02-application/apps/web && npm run build`
  - [ ] Copy `02-application/apps/web/.next/static` → `02-application/apps/web/.next/standalone/.next/static`
  - [ ] E2E production **19/19 passing**
- [ ] Tidak ada `.git` di dalam `02-application/`.
- [ ] `02-application/` tidak mengandung file generated (`node_modules`, `.next/`, binaries).
- [ ] Plan yang dieksekusi sudah ada di `06-temp/plans/README.md` index.

## Sebelum Redeploy Production

- [ ] Semua verification gates hijau.
- [ ] Subtree application sudah di-push ke GitHub.
- [ ] Root orchestrator sudah di-push ke GitHub.
- [ ] Docker Compose production valid:
  - [ ] `docker compose -f docker-compose.prod.yml config` exit 0
- [ ] Direktori `/backups/vexa/db`, `/backups/vexa/wireguard`, `/backups/vexa/logs` ada.
- [ ] `.env` production sudah terverifikasi.

## Common Mistakes to Avoid

1. **Jangan push dari dalam `02-application/`** — tidak ada git context di sana.
2. **Jangan lupa `git subtree push` setelah `git push origin main`.**
3. **Jangan commit `.env` atau file auth Playwright.**
4. **Jangan lupa copy `.next/static` ke standalone setelah build.**
5. **Jangan lupa update `Progress.md` di Obsidian setelah session signifikan.**
6. **Jangan biarkan remote invalid mengendap di `.git/config`.**
7. **Jangan lupa mount `/backups` di production compose.**
