#!/bin/sh
set -e

# Ensure the mounted logs directory is writable by the vexa user
if [ -d /app/logs ]; then
    chown -R vexa:vexa /app/logs
fi

# Ensure other writable dirs are owned by vexa user
chown -R vexa:vexa /etc/wireguard /var/lib/vexa

# Drop privileges and run the server
exec setpriv --reuid=vexa --regid=vexa --init-groups -- /app/server "$@"
