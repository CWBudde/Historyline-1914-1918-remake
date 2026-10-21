# PLAN — Historyline 1914-1918 remake

**Goal:** a complete, playable remake in Go/Ebitengine with newly authored art (2×, optional 4×), a fully
recreated UI and an optional original-rules profile.

**Art direction:** continuous painted illustration (see [docs/art/README.md](docs/art/README.md)). Keep the
original's gameplay IDs, map placement, connection topology and information layouts; never ship original art
or close reconstructions. Unit sprites and portraits follow their reference closely in composition at the
original size, in the painted style.

**Rule sets:** `Classic` (the inherited Python behaviour), `Standard` (Classic with fixes), and the original
profile in `internal/dosrules` (not wired into the game yet). Keep them separate; never mix their mechanics.

Tick a box only with evidence you ran. Record decisions here, not their history.

## Where things stand

- **Rules engine** (`internal/game`): hex map, terrain and seasons, movement (Dijkstra, movement types),
  three-state fog with remembered units/owners, combat with counter fire and ammo, containers (HQ, depots,
  factories), capture, HQ victory, end turn. Unit stats and attacks follow the original unit table
  (`internal/data/testdata/unitdat.json`).
- **Front end** (`cmd/historyline`): loads the shipped levels, camera with zoom and independent UI scale and
  asset density, HiDPI and fullscreen, minimap, text info box for terrain and units, placeholder battle panel,
  hit/death/capture/round banners, reduced motion, keyboard/mouse controls, computer opponent.
- **Runtime art**: unit and terrain catalogs with context-aware composition (`internal/mapcompose`), drawn with
  `just play`. Interface packs build, but the game does not use them yet.
- **AI** (`internal/ai`): greedy prototype through game commands and its own vision; plays all shipped levels.
- **Original rules** (`internal/dosrules`): combat (casualties, support, distance, experience, order handler),
  movement queries and routes, loading/unloading/exits, visibility and the phase change, with fixtures for the
  full game and the demo.
- **Art**: candidates for every enabled ID (134 tiles, 15 unit families with portraits, interface, screens);
  browse them in `gallery/`. Nothing is approved yet.

## Open work

### 1. Playable game (art and UI)

- [ ] Ownership strips on units and buildings; check the composed map by hand in the window, including
      fractional zoom and 4K/HiDPI.
- [ ] Context rules for mixed-family neighbours (forest next to hills, mud, mountain, buildings) and the
      map edge; contexts for imported maps and mixed seasons.
- [ ] Decide the open art ambiguities: HQ entrance cell (N or C), the `trench5a` port, four inferred
      lakeside corners.
- [ ] Runtime catalogs for portraits, screens, fonts and animation metadata; live pack switching; embed
      approved exports through `assets/assets.go`.
- [ ] Terrain panel: movement/battle/sight bars, selected tile and infrastructure, blank/hidden/impassable
      states, using the panel art.
- [ ] Unit panel: portrait, skills, four weapon rows (empty/unlimited/zero distinct), hidden data suppressed
      before binding; all units, long labels, every UI scale.
- [ ] Battle panel with real bars and sprites; explosion sprites; capture/round banners from the interface art.
- [ ] Hot-seat with handover screen and vision isolation.
- [ ] Hand-test move, attack, enter/leave container, capture and end turn; play all three missions under
      Classic and Standard, both sides, summer/winter; smoke-test desktop and WASM.
- [ ] **Milestone:** all three levels play like the Python version (apart from fixed bugs).

### 2. Art completion and acceptance

- [ ] Approve terrain: grass and context transitions, full-map repetition, seasons, seams at three-tile
      vertices.
- [ ] Approve units: silhouettes and colour separation at the original size, mechanical/camera consistency,
      physical contact, all owners and states.
- [ ] Art for the 33 original unit types without art yet (construction unit, AA, anti-tank guns, five tanks,
      five ship types, balloon, 18 aircraft): six facings, three palettes, normal/spent, 2×/4×, portraits.
- [ ] Confirm whether summer factories, barbed wire, background `trenches` and summer rough/shelled/stony
      ground are needed, then author them.
- [ ] Campaign, intermission, production/repair and editor illustrations once those features exist.
- [ ] In-engine art preview (`cmd/hlartpreview`): seven-hex patches, random fixtures, shipped maps; switch
      family, season, facing and state; grid/port overlays; reload exports; screenshots.
- [ ] Broken-join fixtures for the compositor (wrong ports, shifted anchors, missing corners, double edges).
- [ ] `just art-build` / `art-check` / `art-preview`; `art-check` in CI.
- [ ] Windows: make the exporter tests pass (path/identity handling, directory rollback).
- [ ] Visual acceptance on all shipped maps at 720p/1080p/1440p/4K, display scales 100–200 %, UI scales
      100–150 %, zoom 0.75–2; check halos, bleed, repetition, cracks, picking, text.
- [ ] Performance: name baseline hardware; 60 FPS (p95 ≤ 16.7 ms) on the stress map; startup time, download
      size and texture memory. Ship 4× only if it passes the same review.

### 3. Original rules profile

- [ ] Data: name the 36 unit records without a Go type and decide which share a family; alternate group
      strength; terrain record names; which season a map uses.
- [ ] Order cycle: movement/action phases, both players' confirmations, cancellation, eligibility, move
      limits, conflicting destinations and swaps, targets destroyed before their order executes.
      Known so far: only the acting side's attack list resolves, in order, without an alive check; a
      destroyed unit leaves the map at once; a surviving target is fought again; a cancelled exit returns
      the unit to its slot.
- [ ] Engine state for profile, cycle, readiness, pending orders; validation separate from execution; ordered
      events for UI, AI and replays. Save/resume of unconfirmed and confirmed orders.
- [ ] Energy and buildings: credits per building, repair (15 energy, 50 for construction units; restores
      group size, keeps experience, resets construction charges to 2), factory build lists and prices,
      capture of buildings and their stored units, depot construction, sapper trenches, barbed wire.
- [ ] Victory and score: annihilation, ties, capture timing; +100 for hidden inventories, ×4/×3/×2 for
      4/8/16 move limits, cap 32,500.
- [ ] Wire the profile into the game and UI (phase display, pending orders, energy/repair/production
      controls); one integrated scenario covering both cycles, combat, transport, capture, repair and
      production; both factions and seasons; deterministic replay. Classic/Standard stay green.

### 4. Computer opponent

- [ ] Difficulty setting (optionally allowed to cheat); artillery stand-off; skip the rest of the AI turn.
- [ ] Original-profile AI on the order-cycle API: legal orders per cycle, evaluated with the resolver
      (return fire, support, hidden opposing orders), plus capture value, HQ distance, terrain defence.
- [ ] Strategic roles (attack, defend HQ/depots, capture, escort) and an energy budget.
- [ ] Headless AI-vs-AI runner recording rules/data versions and orders for replay.

### 5. Content

- [ ] Full map format and importer (`cmd/finextract`): header, dimensions, tiles, units, buildings,
      inventories, energy, availability, mission metadata → versioned level JSON (the current reader stays).
- [ ] Remaining campaign maps for both sides (only `*_cp` exist).
- [ ] Complete unit roster with domain-specific firepower (aircraft, armour, trains, ships).
- [ ] Briefings, campaign progression, level codes.

### 6. Polish and distribution

- [ ] Screens: title, side/mission selection, briefing/campaign, options, pause, save/load, handover,
      victory/defeat, credits, with hover/focus/pressed/disabled states and keyboard navigation.
- [ ] Sound and music (`ebiten/audio`).
- [ ] Map editor (`-edit` or `cmd/hlmapedit`) with tile/infra/unit/container palettes and a new-map wizard.
- [ ] WASM build on GitHub Pages; release binaries via GoReleaser.
- [ ] Remove `legacy/` once nothing depends on it.

### Beyond the original

- [ ] Artillery can destroy infrastructure.
- [ ] Dedicated ammo and fuel supply units.
- [ ] Engineer units that build infrastructure.

## Open questions

1. Which original quirks should `Standard` change, and which stay in the original profile?
2. Level format successor: exact schema (rules profile, briefing, side, season, energy, availability,
   move limits); keep mission round deadlines separate from per-cycle move limits.
3. Is a complete 4× pack worth it after profiling? The minimum is complete 2× art.
