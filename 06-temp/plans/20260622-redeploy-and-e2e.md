# Redeploy Production + E2E Regression Plan

> Project: vexa  
> Date: 2026-06-22  
> Path: `/home/ubuntu/projects/vexa/06-temp/plans/20260622-redeploy-and-e2e.md`

## Goal

Redeploy aplikasi vexa dari repo baru `/home/ubuntu/projects/vexa/02-application` ke production server, lalu jalankan full Playwright E2E regression suite sampai 19/19 passing.

## Context

- Root orchestrator: `/home/ubuntu/projects/vexa`
- Application repo: `/home/ubuntu/projects/vexa/02-application`
- GitHub app repo: `https://github.com/soumabali/vexa`
- GitHub root repo: `https://github.com/soumabali/vexa-root`
- Production domain: `https://vexa.nexigo.my.id`
- Production API: `https://api-vexa.nexigo.my.id`
- Server IP: `43.156.128.55`
- Existing containers (masih berjalan): `vexa-traefik`, `vexa-web`, `vexa-api`, `vexa-postgres`, `vexa-redis`
- Verification gate sebelumnya: go test/build passing, Next.js build passing
- Production E2E target: 19/19 passing

## Approach

1. Pastikan `.env` production ada di `02-application/` dan valid.
2. Build ulang image web dari source terbaru.
3. Build ulang image API dari source terbaru.
4. Redeploy via `docker compose -f docker-compose.traefik.yml -f docker-compose.prod.yml up -d --build`.
5. Health check containers dan endpoints.
6. Jalankan E2E production suite dengan `SKIP_WEBSERVER=true`.
7. Jika gagal, capture error, perbaiki, dan ulang.
8. Commit perubahan di app repo (kalau ada) dan push ke GitHub.
9. Update root subtree dan push root repo.

## Task Breakdown

### Task 1: Verify production .env

**File:** `02-application/.env`

Pastikan variabel berikut ada dan benar:
- `DOMAIN=vexa.nexigo.my.id`
- `API_DOMAIN=api-vexa.nexigo.my.id`
- `ACME_EMAIL=sudhar.denpasar@gmail.com`
- `DB_USER=vexa`, `DB_PASSWORD=[REDACTED]`, `DB_NAME=vexa`
- `JWT_SECRET`, `JWT_REFRESH_SECRET`, `ENCRYPTION_KEY` terisi
- `ALLOWED_ORIGINS=https://vexa.nexigo.my.id,https://api-vexa.nexigo.my.id`
- `NEXT_PUBLIC_API_URL=https://api-vexa.nexigo.my.id`
- `NEXT_PUBLIC_WS_URL=wss://api-vexa.nexigo.my.id`

**Verify:** `cat .env | grep -E 'DOMAIN|ACME_EMAIL|DB_USER|NEXT_PUBLIC'`

### Task 2: Build web production image

**Dir:** `02-application/apps/web`

```bash
cd /home/ubuntu/projects/vexa/02-application/apps/web
npm install --legacy-peer-deps
npm run build
cp -r .next/static .next/standalone/.next/static
```

**Verify:** `.next/standalone/server.js` exists dan `.next/standalone/.next/static/` terisi.

### Task 3: Build API binary/image

**Dir:** `02-application/apps/api`

```bash
cd /home/ubuntu/projects/vexa/02-application/apps/api
go test ./...
go build ./...
```

**Verify:** exit code 0.

### Task 4: Redeploy production stack

**Dir:** `02-application/`

```bash
cd /home/ubuntu/projects/vexa/02-application
docker compose -f docker-compose.traefik.yml -f docker-compose.prod.yml up -d --build web api
```

Jika perlu full recreate: tambahkan `--force-recreate` untuk web dan api saja.

**Verify:** `docker compose ps` menunjukkan semua container healthy.

### Task 5: Health check endpoints

```bash
curl -s https://api-vexa.nexigo.my.id/health/ready | head
curl -s https://vexa.nexigo.my.id/login | grep -i vexa
```

### Task 6: Run E2E production suite

**Dir:** `02-application/tests/e2e`

```bash
cd /home/ubuntu/projects/vexa/02-application/tests/e2e
SKIP_WEBSERVER=true \
BASE_URL=https://vexa.nexigo.my.id \
API_BASE_URL=https://api-vexa.nexigo.my.id \
npx playwright test --workers=1
```

**Expected:** 19 passed, 0 failed.

### Task 7: Handle failures (if any)

Jika E2E gagal:
1. Baca `playwright-report/` dan `test-results/`.
2. Jalankan debug script production dari skill `production-e2e-regression`.
3. Capture `pageerror` dan console errors.
4. Perbaiki di `apps/web/src/components/...` atau backend.
5. Rebuild + redeploy + re-run E2E.

### Task 8: Commit and push changes

Jika ada perubahan code setelah fix:

```bash
cd /home/ubuntu/projects/vexa/02-application
git add -A
git commit -m "fix(prod): <description>"
git push origin main
```

Lalu update root subtree:

```bash
cd /home/ubuntu/projects/vexa
git subtree pull --prefix=02-application https://github.com/soumabali/vexa.git main --squash
git push origin main
```

### Task 9: Update Obsidian Vault

Ame (Hermes) akan update:
- `infra/vexa.md` dengan status redeploy dan E2E result.
- `03-history/sessions/` dengan session note hari ini.

## Constraints

- Jangan hapus database postgres/redis kecuali benar-benar perlu.
- Jangan expose Traefik dashboard ke publik (port 8081).
- Semua kredensial tetap `[REDACTED]` di log.
- Jika butuh token GitHub untuk push, ambil dari Obsidian Vault dan reset remote URL setelah push.

## Verification Checklist

- [ ] `.env` production valid
- [ ] Web build sukses + static copied
- [ ] API go test/build passing
- [ ] Containers healthy
- [ ] Health endpoints OK
- [ ] E2E 19/19 passing
- [ ] App repo commit + push
- [ ] Root subtree updated + push
- [ ] Obsidian vault updated
