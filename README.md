# Historyline 1914-1918 REMAKE

A remake of the classic Amiga strategy game, written in Go with [Ebitengine](https://ebitengine.org) and
newly painted art. The roadmap is [PLAN.md](PLAN.md).

## Repository layout

| Path | Contents |
| --- | --- |
| `cmd/historyline` | The game |
| `internal/` | Rules engine, data tables, level I/O, hex grid, asset packs, map compositor, AI |
| `levels/` | Level files |
| `art/` | Art masters per runtime family, recipes (prompts) and bundle manifests; see [docs/art](docs/art/README.md) |
| `assets/` | Runtime art the game packs are built from, listed in `assets/index.json` |
| `tools/` | Offline art tooling |
| `docs/` | Art guide, runtime asset packs, reconstructed manual |
| `legacy/` | The upstream Python remake, kept as a test reference only |

## Running

You need Go (see `go.mod`) and [just](https://github.com/casey/just). On Linux, Ebitengine also needs the
OpenGL and X11 headers, e.g. `sudo apt-get install libgl1-mesa-dev xorg-dev libasound2-dev`. `just lint`
needs [golangci-lint](https://golangci-lint.run) v2:
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1`.

```sh
just run        # play with placeholder graphics
just play       # play with the candidate art (builds the packs into build/packs/ on first run)
just test       # go test ./...
just lint       # go vet + golangci-lint
just build      # bin/historyline
just web 8080   # web build into build/web, served on :8080
just ci         # everything CI runs
just clean      # delete all generated output under build/
```

## Art

All shipped art is newly authored; see the [art guide](docs/art/README.md). The art in `legacy/python/gfx/`
has unclear rights (two files are Battle Isle 3 screenshots) and is never redistributed.

```sh
just gfx-check         # art tool tests
just gfx-game-review   # local comparison desk under build/review/gfx/
just gfx-study         # PULSE map study with real game captures at 720p/1080p/4K
```

## Manual

[`docs/manual/`](docs/manual/README.md) is a Typst reconstruction of the original English manual and weapons
booklet. `just manual` builds `build/manual/Historyline_1914-1918_Manual.pdf`.
