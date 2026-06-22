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

### Verification Commands

```bash
# 1. List plugins — must show superpowers@skills-dir and graphify@skills-dir as loaded,
#    and caveman@caveman as enabled.
ollama launch claude --model kimi-k2.7-code:cloud -- plugin list

# 2. Verify skill directories exist
ls -d ~/.claude/skills/{superpowers,caveman,graphify}

# 3. Verify playwright MCP configured
grep -A3 '"playwright"' ~/.claude/settings.json
```

If any skill/MCP is missing or not loaded, stop and report to Ame before coding.

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

### Dispatch Template

```bash
/home/ubuntu/projects/vexa/scripts/dispatch-claude.sh \
  /home/ubuntu/projects/vexa/06-temp/plans/<PLAN>.md \
  <SLUG>
```

Template prompt di dalam plan harus menyertakan:

> "Gunakan skill superpowers, caveman, dan graphify. Jika ada perubahan E2E, gunakan playwright MCP. Setelah selesai, respond ONLY dengan valid JSON sesuai `/home/ubuntu/projects/vexa/00-meta/claude-response-schema.json`."

Claude Code akan dijalankan dengan permission terbatas:
> `--allowedTools "Read,Write,Edit,Bash(go test),Bash(go build),Bash(npm run build),Bash(make),Bash(cp),Bash(rm -f),Bash(mkdir),Bash(ls),Bash(grep),Bash(cd)"`

Untuk command Bash di luar allowlist, gunakan `scripts/safe-exec.sh` atau jalankan manual oleh Ame.

### Expected First Response from Claude Code

Claude Code harus membalas dengan JSON yang valid terhadap schema. Konten minimal harus mengonfirmasi:

1. ✅ `AGENTS.md` dibaca
2. ✅ Root `CLAUDE.md` dibaca
3. ✅ `00-meta/skills.md` dibaca
4. ✅ `00-meta/git-structure.md` dibaca
5. ✅ `02-application/CLAUDE.md` dibaca
6. ✅ Plan dibaca
7. ✅ `superpowers`, `caveman`, `graphify` tersedia
8. ✅ `playwright` MCP tersedia untuk E2E/UI jika diperlukan
9. ✅ Hanya akan edit file di `02-application/`
10. ✅ Tidak akan push/commit dari `02-application/`

Jika tidak, Ame hentikan task dan perbaiki compliance dulu.
