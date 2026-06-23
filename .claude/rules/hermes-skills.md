# Hermes Skills Integration (App Level)

> App-level skill mapping for `02-application/`.  
> Root mapping is at `../../../.claude/rules/hermes-skills.md`.

---

## Mandatory Hermes Skills (Ame)

| Skill | When to use | Mandatory? |
|-------|-------------|------------|
| `claude-code` | Dispatch coding tasks to Claude Code CLI | ✅ Yes |
| `production-e2e-regression` | Run E2E suite against production | ✅ Yes |
| `open-source-project-cleanup` | Rebrand, remove legacy infra, rename identifiers | When relevant |
| `requesting-code-review` | Independent pre-commit review | For large changes |
| `plan` | Write detailed plan for medium+ tasks | ✅ Yes |
| `graphify` | Analyze codebase structure during verification | When relevant |
| `project-directory-structure` | Project structure | When setup/scaffold |
| `hermes-agent` | Hermes config/troubleshooting | When relevant |

---

## App-Level Context

Claude Code inside `02-application/` must load:

- `02-application/CLAUDE.md`
- `02-application/.claude/settings.json`
- `02-application/.claude/rules/hermes-skills.md`

---

## Required Skills / MCP for Claude Code (WAJIB)

Claude Code **wajib** memuat skill/MCP ini **pada setiap sesi coding** di `02-application/`.
**Stack ini bukan optional** — mereka adalah standard project vexa.

| Skill/MCP | Location | Purpose | When to use | Runtime Status |
|-----------|----------|---------|-------------|----------------|
| `superpowers` | `~/.claude/skills/superpowers/` | Coding superpowers | Every coding session | `superpowers@skills-dir` loaded |
| `caveman` | `~/.claude/plugins/cache/caveman/caveman/` | Caveman hooks and workflows | Every coding session | `caveman@caveman` enabled |
| `graphify` | `~/.claude/skills/graphify/` | Codebase graph understanding | Exploration, refactoring, architecture audit | `graphify@skills-dir` loaded |
| `playwright` | MCP `@executeautomation/playwright-mcp-server` | Browser automation | E2E tests or UI verification | configured in `~/.claude/settings.json` |

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

### Required Prompt Phrase

Every dispatch prompt from Ame must include:

> "Use superpowers, caveman, and graphify. If E2E changes are needed, also use the playwright MCP. After execution, respond ONLY with valid JSON according to `/home/ubuntu/projects/vexa/00-meta/claude-response-schema.json`."

---

## Dispatch Command

```bash
/home/ubuntu/projects/vexa/scripts/dispatch-claude.sh \
  /home/ubuntu/projects/vexa/06-temp/plans/<PLAN>.md \
  <SLUG>
```

This wrapper runs Claude Code with restricted permissions:
> `--allowedTools "Read,Write,Edit,Bash(go test),Bash(go build),Bash(npm run build),Bash(make),Bash(cp),Bash(rm -f),Bash(mkdir),Bash(ls),Bash(grep),Bash(cd)"`

For any Bash command outside this allowlist, use `/home/ubuntu/projects/vexa/scripts/safe-exec.sh` or ask Ame to run it manually.

Do not use `--dangerously-skip-permissions` or `--allow-dangerously-skip-permissions` without explicit Ame approval.
