# plx ⚡

> **Lightweight, local-first workspace flight deck by [@xldplx](https://github.com/xldplx).**  
> Sub-millisecond perceived startup latency, zero background bloat, and instant situational awareness across all your local Git repositories.

```
┌── plx : Workspace Cockpit ──────────────────────────────── 14 Repos (3 Dirty) ──┐
│ Filter: [ api_                           ]                       Sort: [Recent] │
├─────────────────────────────────────────────────────────────────────────────────┤
│ > api-gateway       main     * 2 modified, 1 untracked    ↑1 ↓0   12m ago       │
│   auth-service      feat/v2    clean                      ↑0 ↓0    1h ago       │
│   frontend-core     main       clean                      ↑0 ↓2    3h ago       │
│   infra-terraform   stage    * 4 modified                 ↑0 ↓0    2d ago       │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Path: /Users/matt/src/api-gateway                                               │
│ Changed: [M] internal/router.go  [M] go.mod  [?] config.local.yaml              │
├─────────────────────────────────────────────────────────────────────────────────┤
│ [Enter] Jump   [o] Editor   [r] Rescan   [/] Filter   [q/Esc] Quit              │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Why `plx`?

Modern developers context-switch across dozens of local repositories every single day. Full-blown Git GUIs and heavy multiplexers introduce cognitive friction when you just need:
1. **Instant visibility:** Which repositories have uncommitted changes or unpushed commits?
2. **Zero-friction jumping:** Switch to any repository or open it in your editor with two keystrokes.
3. **Speed:** Starts in $<20\text{ ms}$ with zero background daemons or database servers.

---

## Features

- 🚀 **Sub-20ms Startup:** Powered by an atomic, XDG-compliant state cache for instant frame 0 rendering.
- ⚡ **Concurrent Git Scanner:** Inspects Git branches, porcelain dirty status, ahead/behind upstream counts, and commit ages across 50+ repositories concurrently.
- 🎯 **Fuzzy Search:** Built-in fuzzy filtering across repository names and directory paths.
- ⌨️ **Vim-Inspired Keybindings:** `j`/`k` navigation, `/` to search, `Enter` to jump, `o` to launch your editor (`$EDITOR` or VS Code).
- 🪶 **Zero Bloat:** Compiles to a single, standalone static binary with zero external dependencies.
- 🔄 **Headless Mode:** Full CLI support with `--json` and tabular output for scripting and shell pipelines.

---

## Installation

### From Source (Go 1.23+)

```bash
git clone https://github.com/xldplx/plx.git
cd plx
go build -o plx ./cmd/plx
```

Move `plx` into your `$PATH` (e.g. `/usr/local/bin` on Unix or `C:\Program Files\` / user bin on Windows).

---

## Shell Integration (The 1-Letter `x` Jump)

Add the `plx` wrapper to your shell config so typing `x` launches the interactive cockpit and `cd`'s directly into the chosen repository upon exit:

### PowerShell (`$PROFILE`)
```powershell
plx init powershell | Out-String | Invoke-Expression
```

### Zsh / Bash (`~/.zshrc` or `~/.bashrc`)
```bash
eval "$(plx init zsh)"
```

### Usage:
```bash
x             # Launch interactive TUI cockpit and jump on Enter
x api         # Instantly jump to repository fuzzy-matching "api"
```

---

## CLI Command Reference

```bash
plx                      # Launch interactive TUI flight deck
plx list                 # Print tabular status table of all repos
plx list --dirty         # Only list repos with uncommitted changes
plx list --json          # Output full repository metadata as JSON
plx jump <query>         # Emit matched absolute repository path to stdout
plx scan [path]          # Rescan workspace roots and refresh local cache
plx config               # Print active configuration and resolved XDG paths
plx init [shell]         # Print shell wrapper integration
```

---

## Configuration (`config.toml`)

`plx` complies with the **XDG Base Directory Specification**:
* **Linux/macOS:** `~/.config/plx/config.toml`
* **Windows:** `%APPDATA%\plx\config.toml`

```toml
# Directories to search for Git repositories
workspace_roots = [
  "~/Projects",
  "~/src",
  "~/Desktop"
]

# Max directory depth to traverse for .git markers
max_depth = 4

# Directories skipped during scan
ignore_dirs = [
  "node_modules",
  "vendor",
  "target",
  ".cargo",
  "dist",
  ".next",
  ".venv"
]

# Editor to open when pressing 'o'
default_editor = "code"
```

---

## License

MIT © [Matt (@xldplx)](https://github.com/xldplx)
