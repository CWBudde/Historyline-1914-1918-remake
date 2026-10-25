# AGENTS.md

Guidance for AI coding agents (Claude Code, Codex, …) in this repository. `CLAUDE.md` just imports it.

## What this is

A Go/Ebitengine remake of Historyline 1914-1918 with newly authored art. `PLAN.md` is the roadmap and the
source of truth for what is done (tick boxes only with evidence you ran). `legacy/` holds the upstream
Python remake, used only as a test reference.

## Language policy

Write code and tooling in **Go**. Python remains only in `legacy/`; do not add new Python.

## Commands

```sh
just ci          # everything CI runs: lint, test, build, wasm, art-check, reference-check
just check       # ci without the tools/ tests (about 25 min, never cached): for changes outside tools/
just test        # go test ./...
go test ./internal/game -run TestName   # single test
just lint        # go vet + golangci-lint v2 (.golangci.yml); just lint-fix applies its fixes
just fmt         # treefmt (treefmt.toml): gofumpt, gci, typstyle, prettier, shfmt; just fix = lint-fix + fmt
just check-formatted  # fails if just fmt would change a tracked file (formats a copy)
just run         # play; just play uses the candidate art; just build -> bin/; just web 8080
just play-original  # just play, with F8 switching to the original tiles and units from the local original/ (-tags origtiles)
just art-index   # regenerate art/index.json after changing prompts or art; just art-check validates
just artbook     # build/artbook/Historyline_Art.pdf (needs typst)
```

## Architecture

- `cmd/historyline`: the Ebitengine front end; `shell.go` holds the menus around a battle, `app.go` runs
  and draws them.
- `internal/game`: rules engine with **no rendering**; UI and AI drive it through commands
  (`commands.go`). A single rule set (the former `Standard`); the Classic (Python) set was removed.
- `internal/data`: embedded JSON tables (terrain, weapons, units). Unit stats and per-slot attacks follow
  the original unit table (`testdata/unitdat.json`, numbers only). The weapon catalogue is bound to the
  weapon art, so change attack values in a unit's slot, never in `weapons.json`.
- `internal/dosrules`: pure Go resolver of the original rules profile (combat, movement, containers,
  phase change, the command cycle `Game` with save/resume, building capture, sapper trenches, victory
  and score, energy, repair and production), pinned by fixtures under `testdata/`; not wired into `game`
  yet.
- `internal/dosbind`: binds a shipped level to `dosrules` records (cells, terrain records from the
  numbers-only `records.json`, units, container records, building energy); checked against the local
  original maps.
- `internal/dosplay`: a `dosrules` game session on a level for the front end and the computer side: each
  side's view of its own map (no line of sight), destinations, routes, targets, the commands, and events
  in hexes (moves with routes, battles, captures, reveals, phase change, end).
- `internal/dosai`: the computer side on the order cycle: a greedy `Planner` that moves along terrain route
  costs towards the enemy HQ and units, takes buildings on its way and queues the attacks the exact battle
  outcome favours; `PlayOut`.
- `internal/dosscreen`, `internal/dospanel`, `internal/gesture`: the original's split battle screen under
  `-rules dos` (layout, steel frame, panel views, cursor gestures); see `docs/dos-screen.md`.
- `internal/hexgrid`: hex layout (even lines `Width` hexes, odd lines `Width-1`, half-shifted).
- `internal/level`: the legacy level format read/write, what a scenario setup becomes; as files only the
  PULSE reference map (`art/reference/pulse/level.json`) and test fixtures. `internal/tpwm`: Turbo Packer
  unpacker.
- `internal/finmap`: the original map files → versioned level JSON. `internal/origmaps` reads the full
  game's maps from the local `original/` (none of them is committed; tests that need them skip without),
  `cmd/finextract` writes them as JSON for a look. `dosbind.FromOriginal` binds such a map exactly,
  `internal/origview` draws it (a look per terrain record, `looks.json`); `-level original/NN` plays it
  exactly.
- `internal/scenario`: the level format of `levels/monthly/` (the campaign, one level per month,
  `1914-08.json`) and `levels/skirmish/` (the original's two-player maps): one map shared by every
  mode and a setup per mode (`cp`, `ec`, `2p`) with units, building owners, credits, builds and the code
  word, and an optional backdrop, a painting of the whole ground next to the file
  (`1914-08.jpg`) that the game draws in place of the terrain tiles; `File.Level` turns a setup into
  a `level.Level`. The original campaign maps take the first month of each two-month period, their CP
  version's terrain; `just levels-campaign` (`cmd/hlcampaign`) writes those months and the skirmish set
  from the local original maps and overwrites hand edits there (backdrops stay), the other months (newly
  authored levels such as Marne) stay. The menu campaign is a chain per side through the months;
  `-level monthly/1914-08:ec` plays one setup.
- `art/<kind>/<family>/`: selected masters plus `prompt.md` (front matter, look, prompts, variants);
  `art/style.md`, bundle manifests in `art/pipeline/`, the PULSE reference in `art/reference/pulse/`.
  Art guide: `art/README.md`. `assets/<kind>/<family>/`: runtime art only, every file listed in
  `assets/index.json` with its art key.
- `cmd/hlart`: writes and checks the generated `art/index.json` (prompts, masters, runtime files and IDs);
  `docs/artbook/` renders it with Typst as the art book.
- `internal/assets`, `internal/mapcompose`, `internal/unitsprite`, `cmd/hlasset`: runtime art packs built
  from `assets/index.json` (screen packs add the `art/screens` paintings) and map composition
  (`docs/runtime-assets.md`). `internal/maprender` draws a whole level with the runtime art (the art book's
  PULSE map, `hlgallery`'s PULSE study). `internal/waterdepth` shades open water by the original's depths
  (`water`, `waterDeep`, `waterSea`) as one smooth field, blends in the open sea's material and lays surf
  along sea coasts, in the game and in `maprender`.
- `tools/art-review`, `tools/interface-art`, `tools/unit-panel`: art export. `tools/reference-design`: the
  PULSE reference design (`just reference`). `just gfx-check` runs `art-check`, `reference-check` and the
  tool tests. `cmd/hlgallery` is the local comparison desk (`just gfx-game-review`). `tools/candidate-review`
  renders unused imagegen candidates in game context (units on the map, portraits in the panel) for a
  keep/delete decision page under `build/review/candidates/site`.

## Repo rules

- **Commit straight to `main` and `git push origin main`**; no branches or PRs, never force-push. Commits
  end with the `Co-Authored-By:` trailer given by the harness.
- `origin` is the fork CWBudde/Historyline-1914-1918-remake, `upstream` is PoulKalff; pass
  `-R CWBudde/Historyline-1914-1918-remake` to `gh`.
- `original/` (local original game files) and `build/` are git-ignored. Never commit original bytes,
  listings or strings; fixtures hold numbers only. The one exception are the maps' passwords, the menu
  codes in the scenario files. Do not document reverse-engineering methods in the repo.
- Generated output goes under `build/packs/<name>`, `build/web`, `build/manual`, `build/artbook`,
  `build/review/<name>/` or `build/scratch/`, never loose files at the `build/` top level; `just clean` wipes it.
- Keep docs short and current: describe what is, not how it came to be.
- `docs/manual/` is the user's reconstruction of the original manual: leave it alone.
- The graphics track runs in parallel by the user: do not edit `art/`, `assets/` or
  `tools/art-review/` unless asked. `art/index.json` is generated (`just art-index`); never edit it by hand.
- `PLAN.md` may have uncommitted user edits: stage only your own hunks.
- Never touch the `Large` drive (`/run/media/christian/Large`).
