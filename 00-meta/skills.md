# vexa — Skill Mapping (WAJIB)

> SSOT untuk skill/MCP yang **wajib** digunakan dalam project vexa.  
> Hermes skills untuk Ame; Claude Code skills/MCP untuk coding executor.  
> **Skill ini bukan optional.** Kalau tidak tersedia, laporkan ke Ame sebelum coding.

---

## Hermes Skills (Ame — Orchestrator)

| Skill | When to use | Mandatory? |
|-------|-------------|------------|
| `claude-code` | Dispatch coding tasks to Claude Code CLI | ✅ Ya |
| `production-e2e-regression` | Run E2E suite against production | ✅ Ya |
| `open-source-project-cleanup` | Rebrand, remove legacy infra, rename identifiers | Saat relevan |
| `requesting-code-review` | Independent pre-commit review | Saat perubahan besar |
| `plan` | Write detailed plan for medium+ tasks | ✅ Ya |
| `graphify` | Analisis struktur codebase saat verifikasi | Saat relevan |
| `project-directory-structure` | Struktur project | Saat setup/scaffold |
| `hermes-agent` | Hermes config/troubleshooting | Saat setup/troubleshoot Hermes |

---

## Claude Code Skills/MCP (Claude Executor)

**WAJIB aktif pada setiap sesi coding di `02-application/`:**

| Skill/MCP | Lokasi | Fungsi | Kapan digunakan |
|-----------|--------|--------|-----------------|
| `superpowers` | `~/.claude/skills/superpowers/` | Coding superpowers | Setiap sesi coding |
| `caveman` | `~/.claude/skills/caveman/` | Caveman hooks dan workflows | Setiap sesi coding |
| `graphify` | `~/.claude/skills/graphify/` | Codebase graph understanding | Saat eksplorasi, refactoring, audit arsitektur |
| `playwright` | MCP `@executeautomation/playwright-mcp-server` | Browser automation | Saat membuat/mengubah test E2E atau verifikasi UI |

### Cara Verifikasi Instalasi

```bash
ls -d ~/.claude/skills/{superpowers,caveman,graphify}
grep -A3 '"playwright"' ~/.claude/settings.json
```

### Cara Menggunakan dalam Prompt

Sertakan kalimat berikut di setiap prompt dispatch:

> "Gunakan skill superpowers, caveman, dan graphify. Jika ada perubahan E2E, gunakan playwright MCP."

---

## Verification Gates

Sebelum commit/push:

1. `cd 02-application/apps/api && go test ./...`
2. `cd 02-application/apps/api && go build ./...`
3. `cd 02-application/apps/web && npm run build`
4. Copy `02-application/apps/web/.next/static` → `02-application/apps/web/.next/standalone/.next/static`
5. E2E production: `SKIP_WEBSERVER=true BASE_URL=https://vexa.nexigo.my.id API_BASE_URL=https://api-vexa.nexigo.my.id npx playwright test --workers=1`

---

## Dispatch Template

```bash
cd /home/ubuntu/projects/vexa/02-application
ollama launch claude --model kimi-k2.7-code:cloud -- \
  -p "Read the plan at /home/ubuntu/projects/vexa/06-temp/plans/<PLAN>.md and execute. Gunakan skill superpowers, caveman, dan graphify. Jika ada perubahan E2E, gunakan playwright MCP." \
  --dangerously-skip-permissions \
  --allowedTools "Read,Write,Edit,Bash"
```
