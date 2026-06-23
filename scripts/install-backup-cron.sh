#!/usr/bin/env bash
# install-backup-cron.sh — Install daily backup cron for vexa
# Run once on production host.

set -euo pipefail

CRON_LINE="0 2 * * * cd /home/ubuntu/projects/vexa/02-application && /usr/bin/docker compose -f docker-compose.prod.yml run --rm backup >> /var/log/vexa-backup.log 2>&1"

if crontab -l 2>/dev/null | grep -q "vexa-backup"; then
  echo "Backup cron already installed."
else
  (crontab -l 2>/dev/null || true; echo "$CRON_LINE") | crontab -
  echo "Backup cron installed: $CRON_LINE"
fi

echo "Current crontab:"
crontab -l | grep "vexa-backup" || true
