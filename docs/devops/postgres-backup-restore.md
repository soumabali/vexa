# PostgreSQL Backup & Restore — vexa

> P4 #2 — automated daily + weekly backups with retention.

## Overview

`scripts/backup.sh` performs `pg_dump | gzip` against the vexa PostgreSQL
database, writes the artifact under `backups/postgres/`, verifies gzip
integrity, and applies retention. Logs go to `logs/backup.log`.

Two modes:

| Mode    | Schedule      | Retention                          |
| ------- | ------------- | ---------------------------------- |
| daily   | 03:00 daily   | `BACKUP_RETENTION_DAYS` (default 14) |
| weekly  | 04:00 Sunday  | `BACKUP_RETENTION_WEEKS` (default 8) |

## Environment

| Variable                | Default                                              |
| ----------------------- | ---------------------------------------------------- |
| `BACKUP_DIR`            | `<repo>/backups/postgres`                            |
| `BACKUP_RETENTION_DAYS` | `14`                                                 |
| `BACKUP_RETENTION_WEEKS`| `8`                                                  |
| `LOG_DIR`               | `<repo>/logs`                                        |
| `DB_HOST` / `DB_PORT`   | `localhost` / `5432`                                 |
| `DB_USER` / `DB_NAME`   | `vexa` / `vexa`                                      |
| `PGPASSWORD` / `DB_PASSWORD` | (no default — required)                        |

`.env` is loaded automatically if present.

## Install crontab (production host)

```bash
# system-wide (preferred on multi-user hosts)
sudo install -m 644 scripts/backup.crontab /etc/cron.d/vexa-backup
sudo systemctl restart cron

# verify
systemctl status cron
ls -l /etc/cron.d/vexa-backup
```

User-level alternative:

```bash
crontab scripts/backup.crontab
crontab -l
```

## Manual run

```bash
./scripts/backup.sh daily
BACKUP_RETENTION_WEEKS=12 ./scripts/backup.sh weekly
```

Output: `backups/postgres/vexa_<DB>_YYYYMMDD_HHMMSS.sql.gz`.

## Restore

```bash
# to live DB
gunzip -c backups/postgres/vexa_vexa_20260628_030000.sql.gz \
  | psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

# to a throwaway DB for verification
createdb -h "$DB_HOST" -U "$DB_USER" vexa_verify
gunzip -c backups/postgres/vexa_vexa_*.sql.gz \
  | psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/vexa_verify"
psql -d vexa_verify -c "SELECT count(*) FROM users;"
dropdb vexa_verify
```

## Failure runbook

1. Check log tail:
   ```bash
   tail -n 100 logs/backup.log
   ```
2. Common causes:
   - `pg_dump: error: connection to server` → DB host/port/credentials wrong.
   - `ERROR: pg_dump not found` → install postgresql-client (`apt-get install postgresql-client`).
   - `gzip integrity check failed` → disk full or OOM during dump; check `df -h` and `dmesg`.
3. Alert hook: a non-zero exit from `backup.sh` emits nothing by default.
   Add a wrapper that pings your alerting system (e.g. curl a healthchecks.io
   endpoint) when the script exits non-zero.
4. Restore the most recent successful dump to a verify DB (see above) to
   confirm the snapshot is usable.

## Integration test

`scripts/backup_test.sh` boots a throwaway postgres, dumps, restores, and
compares row counts.

```bash
./scripts/backup_test.sh
```

Requires Docker. Tears down automatically.