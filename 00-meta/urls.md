# vexa — URLs & Ports

> Single source of truth untuk endpoint, port, dan domain.  
> Mirror dari Obsidian Vault: `infra/vexa — Project Index.md`.

## Production

| Service | URL | Catatan |
|---------|-----|---------|
| Web app | https://vexa.nexigo.my.id | Next.js via Docker + Traefik |
| API | https://api-vexa.nexigo.my.id | Go/Gin via Docker + Traefik |
| Traefik dashboard | https://traefik.nexigo.my.id/dashboard/ | Basic auth via `.env` |

## Server

| Atribut | Nilai |
|---------|-------|
| Public IP | 43.156.128.55 |
| Provider | Tencent Cloud |
| OS | Ubuntu 22.04 LTS |
| ACME email | sudhar.denpasar@gmail.com |

## Port Mapping

| Port | Protocol | Service | Publik |
|------|----------|---------|--------|
| 80 | TCP | Traefik HTTP → HTTPS redirect | Ya |
| 443 | TCP | Traefik HTTPS | Ya |
| 8081 | TCP | Traefik dashboard | Ya (HTTPS) |
| 22 | TCP | SSH server host | Ya |
| 51820–51830 | UDP | WireGuard tunnels | Ya |
| 3000 | TCP | Web Next.js internal | Tidak |
| 8080 | TCP | API Go internal | Tidak |
| 5432 | TCP | PostgreSQL internal | Tidak |
| 6379 | TCP | Redis internal | Tidak |

## Repositories

| Repo | URL | Scope |
|------|-----|-------|
| Application (public) | https://github.com/soumabali/vexa | Production code |
| Orchestrator root (private) | https://github.com/soumabali/vexa-root | Plans, session logs, docs |

## Local Paths

| Path | Fungsi |
|------|--------|
| `/home/ubuntu/projects/vexa` | Root orchestrator workspace |
| `/home/ubuntu/projects/vexa/02-application` | Application working copy (git subtree) |
| `/home/ubuntu/projects/vexa-backup-20260622-112822` | Backup state lama (dipertahankan sementara) |
| `/backups/vexa/db` | Backup harian PostgreSQL |
| `/backups/vexa/wireguard` | Backup harian WireGuard configs |
| `/backups/vexa/logs` | Backup harian API logs |

## Automation Scripts

| Script | Fungsi |
|--------|--------|
| `scripts/audit-vexa.sh` | Audit workflow, context, skills, backup |
| `scripts/claude-compliance-check.sh` | Validasi structured response Claude Code |
| `scripts/pre-push.sh` | Git pre-push hook |
| `scripts/vexa-status.sh` | Dashboard status snapshot |
| `scripts/new-session.sh` | Generator session note |
| `02-application/scripts/backup.sh` | Backup runner DB, WG, logs |
| `02-application/scripts/install-backup-cron.sh` | Install cron backup harian |
