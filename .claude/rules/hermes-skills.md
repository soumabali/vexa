# Hermes Skills Integration (Root)

> Root-level skill mapping for vexa orchestration.  
> SSOT lengkap ada di `00-meta/skills.md` dan Obsidian Vault `infra/vexa — Project Index.md`.

## Mandatory Hermes Skills

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

## Dispatch Command

```bash
cd /home/ubuntu/projects/vexa/02-application
ollama launch claude --model kimi-k2.7-code:cloud -- \
  -p "Read the plan at /home/ubuntu/projects/vexa/06-temp/plans/<PLAN>.md and execute." \
  --dangerously-skip-permissions \
  --allowedTools "Read,Write,Edit,Bash"
```

## App-Level Context

Claude Code inside `02-application/` must load:
- `02-application/CLAUDE.md`
- `02-application/.claude/settings.json`
- `02-application/.claude/rules/hermes-skills.md`

## Verification Gates

1. `go test ./...`
2. `go build ./...`
3. `npm run build`
4. Copy `.next/static` → `.next/standalone/.next/static`
5. E2E production 19/19 passing
