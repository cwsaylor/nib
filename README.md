# SocNotes

Stream of consciousness notes. A fast, keyboard-driven terminal notes app. Single binary, no dependencies, SQLite-backed.

## Install

```bash
make build
./socnotes
```

Requires Go 1.26+.

## Usage

Notes are stored at `~/.local/share/socnotes/socnotes.db`. The first line of each note becomes its title.

### List view

| Key | Action |
|-----|--------|
| `j` / `k` | Move up / down |
| `n` | New note |
| `e` / `Enter` | Edit selected note |
| `s` / `/` | Search |
| `d` | Delete (moves to trash) |
| `y` | Copy note to clipboard |
| `t` | Open trash |
| `?` | Help |
| `q` | Quit |

### Editor

| Key | Action |
|-----|--------|
| `Ctrl+S` | Save |
| `Esc` | Save and return to list |
| `Ctrl+D` | Discard changes |

### Search

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate results |
| `Enter` | Open selected result |
| `Esc` | Close search |

### Trash

| Key | Action |
|-----|--------|
| `j` / `k` | Move up / down |
| `r` | Restore note |
| `x` | Permanently delete |
| `Esc` | Back to list |

## Features

- **Full-text search** with SQLite FTS5 and live results as you type
- **Soft delete** with 30-day trash (auto-purged on startup)
- **Clipboard support** on macOS, Linux, and WSL
- **Responsive layout** — side-by-side list and preview on wide terminals, stacked on narrow ones
- **Infinite scroll** with cursor-based pagination

## Tech

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO).
