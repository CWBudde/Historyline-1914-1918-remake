# Historyline 1914-1918 REMAKE

A remake of the classic Amiga strategy game, written in Go with [Ebitengine](https://ebitengine.org) and
newly painted art. The roadmap is [PLAN.md](PLAN.md).

## Repository layout

| Path | Contents |
| --- | --- |
| `cmd/historyline` | The game |
| `internal/` | Rules engine, data tables, level I/O, hex grid, asset packs, map compositor, AI |
| `levels/` | Level files |
| `art/` | Art masters and `prompt.md` per runtime family, style guide, PULSE reference; see the [art guide](art/README.md) |
| `assets/` | Runtime art the game packs are built from, listed in `assets/index.json` |
| `cmd/hl` | Unified developer CLI: art, assets, exports, reviews, levels, LLM play and web serving |
| `tools/` | Importable offline art tooling and the PULSE reference design |
| `docs/` | Art book, runtime asset packs, reconstructed manual |
| `legacy/` | The upstream Python remake, kept as a test reference only |

## Running

You need Go (see `go.mod`) and [just](https://github.com/casey/just). On Linux, Ebitengine also needs the
OpenGL and X11 headers, e.g. `sudo apt-get install libgl1-mesa-dev xorg-dev libasound2-dev`. `just lint`
needs [golangci-lint](https://golangci-lint.run) v2:
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1`.

```sh
just run        # play with placeholder graphics
just play       # play with the candidate art and menu screens (builds the packs into build/packs/ when the art changed)
just test       # go test ./...
just lint       # go vet + golangci-lint
just build      # bin/historyline
just build-tools # bin/hl and its native preview helper
just hl --help  # all developer tools; see docs/cli.md
just web 8080   # web build with the art packs into build/web, served on :8080
just ci         # everything CI runs
just check      # ci without the slow art tool tests
just clean      # delete all generated output under build/
```

All developer tools use the [`hl` CLI](docs/cli.md), for example
`just hl art check`, `just hl review gallery` and `just hl llmplay new`.

The game opens with a main menu laid out like the original's: start (country selection, then the battle),
the map code (a month's code word for its side, PULSE or BATLE to start; a victory sets the next month),
game values (full or short battles, the computer or a passive opponent, a move limit), control
(fullscreen, scrolling) and extras (the score list, and LOAD GAME for the one saved battle).
The loading page shows a one-line briefing derived from the map (side, season, two-player maps).
In battle, Q asks whether to leave it — STAY, SAVE AND LEAVE (one slot, `-rules dos` only, continued
via EXTRAS → LOAD GAME) or LEAVE; after its end Enter shows the outcome. `-level` or `-player`
skip the menu and start that battle directly.

v1 is one-player: two human players are not in v1. Maps 48-71 are two-player maps in the original;
here you play one side against the computer or a passive opponent.

`-rules` picks the rule set (pass it after `just run` or `just play`): `dos` (default), the original order
cycle, or `standard`, the remake rules. In `dos` both sides give their orders at the same
time: in the movement phase click an own unit and a blue hex, in the action phase an own unit and a red hex
(a queued attack), click an own building for its contents, F1 confirms (again: reopens); once both sides
have confirmed the resolution plays back and the phases swap. `-player central|entente` picks the side,
`-opponent ai|pass` the other side's player.

## Art

All shipped art is newly authored; see the [art guide](art/README.md). The art in `legacy/python/gfx/`
has unclear rights (two files are Battle Isle 3 screenshots) and is never redistributed.

```sh
just art-index         # regenerate art/index.json from the prompts, masters and assets/index.json
just artbook           # art book: every family's prompt, masters and runtime samples, the rendered PULSE map (needs Typst)
just gfx-check         # art catalogue, PULSE reference data and art tool tests
just reference         # PULSE reference review and reference/render comparison (needs RefDesign2.png)
just gfx-game-review   # local comparison desk under build/review/gfx/
just gfx-study         # PULSE map study with real game captures at 720p/1080p/4K
```

## Manual

[`docs/manual/`](docs/manual/README.md) is a Typst reconstruction of the original English manual and weapons
booklet. `just manual` builds `build/manual/Historyline_1914-1918_Manual.pdf`.
