# Web edition

The TypeScript edition in `webapp/` is the site GitHub Pages serves. It plays the original rules profile, the
extension units and the campaign like the desktop game, on any resolution and device, and installs as an app.
Commands are in [AGENTS.md](../AGENTS.md); assets, offline play and the import of the player's own original game
are in [runtime-assets.md](runtime-assets.md); the split screen and the order menu in [dos-screen.md](dos-screen.md).

## Architecture

- A pnpm workspace in strict TypeScript, built by Vite: `webapp/engine` and `webapp/app`.
- `webapp/engine` is a full port of `dosrules`, `dosbind`, `dosplay` and `dosai` (with `hexgrid`, `level`,
  `scenario`, `campaign` and the original map readers). It is pure: ESLint forbids rendering, DOM and Node
  imports outside tests, so it runs in Node, the browser and a worker. Saves share the Go JSON format.
- `webapp/app`: PixiJS v8 draws the map and the battle scene; menus, panels and the HUD are Preact with
  signals, which never drive the frame loop. The computer side plans in a Web Worker (`computer.worker.ts`)
  on session snapshots and returns replayable commands.
- Go stays the source of the data: exports and goldens are `hl webapp` commands (`internal/hlwebapp`), and the
  CPU-heavy map work (`mapcompose`, `waterdepth`, `seasonblend`) is baked into the export. Only the original
  game's maps, which have no baked layers, are composed in the browser.

## Lockstep with Go

The engine must give exactly what Go gives:

- every record of the 14 fixture files under `internal/dosrules/testdata/`, and Go saves that load and re-save
  byte-identical;
- `hl webapp golden` plays `dosai` against `dosai` on every shipped setup (156, up to 300 resolutions each) and
  writes the commands, events and save after each resolution to `build/webapp/golden/`.
  `playGolden.test.ts` replays the commands and also plans both sides in TS; events and saves must match;
- `hl webapp original` exports Go's bindings and traces of the local original maps for the same checks
  (skipped without `original/`).

`just webapp-test` and `just webapp-check` regenerate the goldens first, so a Go rule change without its TS twin
fails CI.

## Layouts and play

- The layout setting is auto, single or original (`storage/store.ts`). Auto shows the original split screen on a
  wide window (`splitByDefault`) and the single map otherwise: the map fills the viewport, with the panel at the
  side or, upright, an orders board.
- Two humans on one screen and hot-seat with protected handovers play the `2p` setups.
- The desktop game's keys, the right-click order menu as a long press on touch, tap to select, move and attack;
  a UI scale setting.
- Saves live in IndexedDB (slots, export and import as a file); settings and best scores beside them.

## Tests and budgets

- Vitest for the engine and the app; Playwright (`webapp/app/e2e`) against the production build on Chromium,
  Firefox and WebKit, plus Android and iPhone emulation, including a playthrough of a level to its end.
- `just webapp-size` fails when the site exceeds `DefaultSiteLimitMB` (`internal/hlwebapp/size.go`); the
  `budget` spec keeps the first playable battle within 8 MB of transfer on Chromium.
