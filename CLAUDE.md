# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

jsawn is a CLI tool that displays JSON in an interactive, tree-based TUI viewer with collapsible nodes, syntax highlighting, and vim-style navigation. Built with Go.

## Build & Run Commands

```bash
# Build
go build -v

# Run (piped JSON or file argument)
echo '{"key": "value"}' | ./jsawn
cat file.json | ./jsawn
./jsawn data.json

# Release (triggered by pushing a git tag)
git tag v1.0.0 && git push origin v1.0.0
```

There are no tests or linter configured.

## Architecture

**Entry flow:** `main.go` → `cmd/root.go` → `tree/` → `viewPage/`

- **`cmd/root.go`** — Cobra CLI setup. Accepts JSON via piped stdin or a file path argument. Parses JSON into a tree, then launches the BubbleTea TUI.
- **`tree/node.go`** — JSON tree data model. `Parse()` unmarshals JSON into a `Node` tree (object, array, string, number, bool, null). Supports collapse/expand, depth-based folding, and computing visible entries for rendering.
- **`tree/render.go`** — Renders tree entries to styled strings using lipgloss. Keys=cyan, strings=green, numbers=yellow, booleans=magenta, null=red. Handles cursor highlighting, collapsed summaries (e.g. `{... 3 keys}`), and indentation.
- **`viewPage/viewPage.go`** — BubbleTea model with custom scrolling viewport. Vim-style keys: j/k or arrows to move, h/l to collapse/expand, space to toggle, e/E to expand/collapse all, 1-9 for depth, gg/G for top/bottom, q/esc to quit. Mouse wheel scrolling supported.

## Distribution

GoReleaser builds for macOS and Windows (amd64/arm64). Homebrew tap at `aryankumar07/homebrew-tap`.
