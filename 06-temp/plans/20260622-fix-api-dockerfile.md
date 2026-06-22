# Fix API Dockerfile WireGuard Tools

## Problem

Docker build for `02-application/apps/api/Dockerfile` fails at stage `api stage-5`:

```
COPY --from=wg-tools /usr/share/wireguard-tools /usr/share/wireguard-tools
failed to calculate checksum: "/usr/share/wireguard-tools": not found
```

## Goal

Fix Dockerfile agar build sukses tanpa mengubah runtime behavior (WG tetap support, non-root, healthcheck, distroless base).

## Tasks

### Task 1: Inspect debian:12-slim wireguard-tools package

Cari path yang benar dari `wireguard-tools` di `debian:12-slim`.

```bash
docker run --rm debian:12-slim sh -c 'apt-get update && apt-get install -y --no-install-recommends wireguard-tools && dpkg -L wireguard-tools | sort'
```

### Task 2: Update Dockerfile

File: `02-application/apps/api/Dockerfile`

- Hapus baris `COPY --from=wg-tools /usr/share/wireguard-tools /usr/share/wireguard-tools`.
- Jika `/usr/share/wireguard-tools` tidak ada, ganti dengan menyalin file template/config yang diperlukan saja, atau biarkan jika runtime cukup `wg` dan `wg-quick`.
- Pastikan stage `wg-tools` tetap hanya satu.
- Pastikan tidak ada dua `FROM` production stage (saat ini ada `FROM gcr.io/distroless` lalu `FROM debian:12-slim`, lalu lagi `FROM gcr.io/distroless`). Hapus duplikat yang pertama.

### Task 3: Verify build

```bash
cd /home/ubuntu/projects/vexa/02-application/apps/api
DOCKER_BUILDKIT=1 docker build -t vexa-api-test -f Dockerfile .
```

### Task 4: Redeploy production

```bash
cd /home/ubuntu/projects/vexa/02-application
sudo docker compose -f docker-compose.prod.yml up -d --build api web
```

## Constraints

- Do not change Go source code.
- Keep non-root user.
- Keep privileged + NET_ADMIN via docker-compose (already configured).
- Do not remove wg/wg-quick/ip/iptables.
