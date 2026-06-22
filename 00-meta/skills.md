# vexa — Skill Mapping

> SSOT untuk skill/MCP yang wajib digunakan dalam project vexa.  
> Hermes skills untuk Ame; Claude Code skills/MCP untuk coding executor.

## Hermes Skills (Ame)

| Skill | When to use |
|-------|-------------|
| `claude-code` | Dispatch coding tasks to Claude Code CLI |
| `production-e2e-regression` | Run E2E suite against production |
| `open-source-project-cleanup` | Rebrand, remove legacy infra, rename identifiers |
| `requesting-code-review` | Independent pre-commit review |
| `plan` | Write detailed plan for medium+ tasks |
| `graphify` | Analisis struktur codebase |
| `project-directory-structure` | Struktur project |
| `hermes-agent` | Hermes config/troubleshooting |

## Claude Code Skills/MCP (Claude Executor)

| Skill/MCP | Lokasi | Fungsi |
|-----------|--------|--------|
| `superpowers` | `~/.claude/skills/superpowers/` | Coding superpowers |
| `caveman` | `~/.claude/skills/caveman/` | Caveman hooks dan workflows |
| `graphify` | `~/.claude/skills/graphify/` | Codebase graph understanding |
| `playwright` | MCP `@executeautomation/playwright-mcp-server` | Browser automation |

## Verification Gates

Sebelum commit/push:

1. `cd 02-application/apps/api && go test ./...`
2. `cd 02-application/apps/api && go build ./...`
3. `cd 02-application/apps/web && npm run build`
4. Copy `02-application/apps/web/.next/static` → `02-application/apps/web/.next/standalone/.next/static`
5. E2E production: `SKIP_WEBSERVER=true BASE_URL=https://vexa.nexigo.my.id API_BASE_URL=https://api-vexa.nexigo.my.id npx playwright test --workers=1`

## Dispatch Template

```bash
cd /home/ubuntu/projects/vexa/02-application
ollama launch claude --model kimi-k2.7-code:cloud -- \
  -p "Read the plan at /home/ubuntu/projects/vexa/06-temp/plans/<PLAN>.md and execute." \
  --dangerously-skip-permissions \
  --allowedTools "Read,Write,Edit,Bash"
```
