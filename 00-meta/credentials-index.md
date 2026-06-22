# vexa — Credentials Index

> Daftar semua kredensial project. **Nilai asli hanya di Obsidian Vault.**  
> File ini hanya index — tidak mengandung secret.

## Vault Locations

| Credential | Obsidian Path | Gunakan Untuk |
|------------|---------------|---------------|
| GitHub PAT (Ame Hermes) | `credentials/Github Credentials.md` | `gh` CLI, push GitHub |
| Project secrets | `credentials/vexa Credentials.md` | `.env`, DB, JWT, WireGuard |
| Server access | `credentials/vexa Credentials.md` | SSH/Tencent Cloud |

## Application Environment Variables

Lihat `02-application/.env.example` untuk daftar lengkap variabel.  
Nilai real di Obsidian Vault.

## Rules

1. **Never commit** `.env`, `.env.local`, `*.pem`, `*.key`, atau file kredensial apapun.
2. **Gunakan `[REDACTED]`** di chat, session notes, dan code comments.
3. **Rotasi** token/password minimal saat: (a) leak, (b) personnel change, (c) 90 hari untuk token.
