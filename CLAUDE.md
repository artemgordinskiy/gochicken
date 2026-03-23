# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A 2D side-scrolling chicken game built with Go and [Ebitengine](https://ebitengine.org/) (v2.5.5). The chicken can move left/right and jump, with a scrolling background. Single-file game (`main.go`), module name `chickenman`, Go 1.20.

## Build & Run

```bash
# Run the game (must be run from repo root — assets use relative paths)
go run .

# Build binary
go build -o gochicken .

# Run tests
go test ./...
```

**Note:** Ebitengine requires CGO and platform graphics libraries. On macOS this works out of the box. On Linux, you may need X11/Wayland dev packages.

## Architecture

Everything lives in `main.go` — a single `Game` struct implements Ebitengine's `ebiten.Game` interface (`Update`, `Draw`, `Layout`):

- **Update cycle**: `handleInput()` processes keyboard (arrow keys + space), `applyPhysics()` applies gravity/ground collision
- **Rendering**: `Draw()` renders scrolling background then the chicken sprite (flipped based on direction)
- **Assets**: loaded from `assets/` via `ebitenutil.NewImageFromFile` — paths are relative to CWD, so always run from repo root

## Assets

Located in `assets/`. Includes `.pxd` source files (Pixelmator Pro) alongside exported `.png` sprites. `jump.wav` and `obstacle.png` exist but are not yet wired into the game.
