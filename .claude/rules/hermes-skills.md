# Hermes Skills Integration (Root)

> Root-level skill mapping for vexa orchestration.

## Mandatory Hermes Skills

| Skill | When to use |
|-------|-------------|
| `open-source-project-cleanup` | Rebrand, remove legacy infra, rename identifiers |
| `production-e2e-regression` | Run E2E suite against production |
| `claude-code` | Dispatch coding tasks to Claude Code CLI |
| `requesting-code-review` | Independent pre-commit review |
| `plan` | Write detailed plan for medium+ tasks |

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
