# plx

lightweight, local-first workspace flight deck and context switcher by [@xldplx](https://github.com/xldplx).

```
┌── plx : workspace cockpit ──────────────────────────────── 14 repos (3 dirty) ──┐
│ filter: [ api_                           ]                       sort: [recent] │
├─────────────────────────────────────────────────────────────────────────────────┤
│ > api-gateway       main     * 2 modified, 1 untracked    ↑1 ↓0   12m ago       │
│   auth-service      feat/v2    clean                      ↑0 ↓0    1h ago       │
│   frontend-core     main       clean                      ↑0 ↓2    3h ago       │
│   infra-terraform   stage    * 4 modified                 ↑0 ↓0    2d ago       │
├─────────────────────────────────────────────────────────────────────────────────┤
│ path: /home/user/src/api-gateway                                                │
│ changed: [m] internal/router.go  [m] go.mod  [?] config.local.yaml              │
├─────────────────────────────────────────────────────────────────────────────────┤
│ [enter] jump   [o] editor   [r] rescan   [/] filter   [q/esc] quit              │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## overview

`plx` is a fast terminal cockpit designed to track, inspect, and jump between local git repositories without background daemons, heavy runtimes, or external databases.

it solves two daily developer frictions:
1. instant situational awareness: see which repositories across your machine have uncommitted changes, untracked files, or unpushed commits.
2. zero-friction navigation: jump to any project directory or launch your editor in two keystrokes.

---

## key features

- sub-20ms perceived startup latency using an xdg-compliant atomic state cache.
- bounded concurrent git scanner evaluating branches, porcelain dirty status, ahead/behind upstream counts, and commit age.
- interactive tui built on the elm architecture (`bubbletea` + `lipgloss`).
- real-time fuzzy matching across repository names and file paths.
- full headless cli support with tabular and json output for shell scripts and pipelines.
- single static binary with zero external dependencies.

---

## installation

### one-line install (go toolchain)

```bash
go install github.com/xldplx/plx/cmd/plx@latest
```

### pre-compiled binaries (zero dependencies)

download the latest archive for your platform from [github releases](https://github.com/xldplx/plx/releases):
- windows: `plx_Windows_x86_64.zip`
- macos (apple silicon): `plx_Darwin_arm64.tar.gz`
- macos (intel): `plx_Darwin_x86_64.tar.gz`
- linux: `plx_Linux_x86_64.tar.gz`

extract the `plx` binary into any directory in your `$PATH`.

### build from source

```bash
git clone https://github.com/xldplx/plx.git
cd plx
go build -ldflags="-s -w" -o plx ./cmd/plx
```

---

## shell integration

`plx` includes a shell integration hook that defines a one-letter `x` shortcut. typing `x` opens the visual cockpit and automatically `cd`s into the selected repository upon exit.

### automatic setup (recommended)

run the setup command to automatically configure your shell profile:

```bash
plx setup
```

### manual setup

#### powershell

add to your `$PROFILE`:

```powershell
plx init powershell | Out-String | Invoke-Expression
```

### zsh / bash

add to your `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(plx init zsh)"
```

### fish

add to `~/.config/fish/config.fish`:

```fish
plx init fish | source
```

### usage

```bash
plx           # open interactive tui cockpit and cd on enter
x             # 1-letter shortcut to open cockpit and cd on enter
x api         # resolve fuzzy query "api" and jump directly
```

---

## cli reference

| command | description |
| :--- | :--- |
| `plx` | launch interactive tui flight deck (falls back to `list` when piped) |
| `plx list` | print tabular overview of discovered repositories |
| `plx list --dirty` | filter output to only repositories with uncommitted changes |
| `plx list --json` | emit machine-readable json array of repository metadata |
| `plx jump <query>` | emit best fuzzy-matched absolute repository path to stdout |
| `plx scan [path]` | trigger immediate filesystem crawl and refresh state cache |
| `plx setup` | automatically install the 'x' shell hook into your profile |
| `plx config` | display active configuration and resolved paths |
| `plx init [shell]` | print shell wrapper script (`powershell`, `zsh`, `bash`, `fish`) |

---

## tui keybindings

| key | action |
| :--- | :--- |
| `↑` / `k` | move cursor up |
| `↓` / `j` | move cursor down |
| `/` | activate fuzzy filter input |
| `esc` | clear active filter or exit filter mode |
| `enter` | select highlighted repository and print path |
| `o` | open repository in configured editor (`$EDITOR` or `code`) |
| `r` | trigger concurrent rescan of all workspace roots |
| `q` / `ctrl+c` | quit |

---

## configuration

configuration is stored in `config.toml` following the xdg base directory specification:

- unix / macos: `~/.config/plx/config.toml`
- windows: `%APPDATA%\plx\config.toml`

### schema

```toml
# directories to search for git repositories
workspace_roots = [
  "~/projects",
  "~/src",
  "~/desktop"
]

# maximum directory depth to traverse
max_depth = 4

# directory names to prune from traversal
ignore_dirs = [
  "node_modules",
  "vendor",
  "target",
  ".cargo",
  "dist",
  ".next",
  ".venv",
  "build"
]

# editor executable launched with 'o'
default_editor = "code"
```

---

## cache & state management

`plx` maintains an atomic snapshot cache to ensure instantaneous startup even across hundreds of projects:

- unix / macos: `~/.local/state/plx/cache.json`
- windows: `%LOCALAPPDATA%\plx\cache.json`

### design guarantees
- **instant paint:** on boot, the interface renders frame 0 directly from the local cache.
- **async revalidation:** background workers refresh git states concurrently without locking the ui.
- **crash resilience:** cache flushes write to `.tmp.<pid>` before issuing an atomic rename, preventing file corruption on abrupt terminal termination.
- **graceful degradation:** operates entirely in memory if write permissions are unavailable.

---

## license

mit © [matt (@xldplx)](https://github.com/xldplx)
