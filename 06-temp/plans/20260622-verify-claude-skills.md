# Plan: Verify Claude Code Skills/MCP for vexa

## Goal

Memastikan skill/MCP wajib untuk Claude Code sudah terinstall, aktif, dan dapat digunakan. Skill/MCP yang harus diverifikasi:

1. `superpowers` — `~/.claude/skills/superpowers/`
2. `caveman` — `~/.claude/skills/caveman/` + hooks
3. `graphify` — `~/.claude/skills/graphify/` + binary
4. `playwright` MCP — `@executeautomation/playwright-mcp-server`

## Verification Steps

### 1. superpowers

```bash
ls -la ~/.claude/skills/superpowers/
# Expected: SKILL.md dan sub-folder references/templates/scripts
```

### 2. caveman

```bash
ls -la ~/.claude/skills/caveman/
ls -la ~/.claude/hooks/
cat ~/.claude/settings.json | grep -A 5 caveman
# Expected: hooks caveman-activate.js, caveman-mode-tracker.js, enabledPlugins caveman@caveman
```

### 3. graphify

```bash
ls -la ~/.claude/skills/graphify/
which graphify
graphify --help
# Expected: skill directory dan binary tersedia
```

### 4. playwright MCP

```bash
cat ~/.claude/settings.json | grep -A 5 playwright
npx -y @executeautomation/playwright-mcp-server --help
# Expected: mcpServers.playwright teregister dan package bisa dijalankan
```

## Fallback / Repair

Kalau salah satu missing:

- **superpowers**: clone `https://github.com/obra/superpowers` ke `~/.claude/skills/superpowers/`
- **caveman**: clone `https://github.com/JuliusBrussee/caveman`, jalankan `node bin/install.js --only claude --with-hooks --non-interactive`
- **graphify**: `pip install "graphifyy[ollama,mcp]"` atau `uv tool install "graphifyy[ollama,mcp]"`, lalu `graphify install`
- **playwright MCP**: tambahkan ke `~/.claude/settings.json`:
  ```json
  {
    "mcpServers": {
      "playwright": {
        "command": "npx",
        "args": ["-y", "@executeautomation/playwright-mcp-server"]
      }
    }
  }
  ```

## Documentation

Setelah verifikasi, update:

- `infra/vexa Claude Code Setup.md`
- `infra/vexa Rules.md`
- `infra/vexa — Project Index.md`
- root `CLAUDE.md`

## Verification Gate

Semua skill/MCP harus lolos check command di atas sebelum plan dinyatakan selesai.
