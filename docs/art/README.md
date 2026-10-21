# Art guide

All shipped art is newly authored. Keep the original's gameplay semantics, stable runtime IDs, map
placement and connection topology, and its familiar information layouts. Never ship original graphics or
close reconstructions of them.

## Style

- Continuous painted illustration with fine material detail. No visible pixel squares, staircase edges,
  nearest-neighbour enlargement or grid tracing.
- Terrain is seen directly from above, under soft diffuse light from the upper left. Ground is muted
  sage/moss in calm values so units read clearly. Trees show aerial crowns, and world scale stays the
  same across ground, sparse forest and dense forest.
- Period roads are unmarked gravel or macadam. Rivers and roads continue past their declared ports,
  including banks and verges.
- Interface: slate/brass vector vocabulary for frames, controls, symbols and map markers; painted
  weapons and explosions. Art never contains baked text or values; labels are live.
- Units: six separately authored views per family (`N, NW, SW, S, SE, NE`, sprite indices 0–5), never
  rotations or mirrors. The owner palette (`cp`, `ec`, `neutral`) is applied at runtime, so capture changes
  colour without changing the unit's class or country.

## Geometry

- Logical terrain tile: 96×80, placed at `x = 142*column + 71*(line%2)`, `y = 40*line` (even lines are
  `Width` hexes, odd lines `Width-1`, half-shifted; see `internal/hexgrid`).
- Hex polygon `(25,0),(71,0),(96,40),(71,80),(25,80),(0,40)` with complementary half-open ownership at
  pixel centres, so aligned 2×/4× source-over blits give every pixel exactly one owner.
- Ports are numbered clockwise: 1 NE, 2 SE, 3 S, 4 SW, 5 NW, 6 N. Neighbour offsets
  `(71,-40),(71,40),(0,80),(-71,40),(-71,-40),(0,-80)`; shared port centres `(83.5,20),(83.5,60),(48,80),
  (12.5,60),(12.5,20),(48,0)`; opposite port = `(p+2)%6+1`. Filename digits are not authoritative:
  use the explicit port lists in the manifests.
- Unit sprites: 96×96 logical canvas at terrain origin `T + (0,-9)`, so anchor `(48,49)` sits on the
  terrain centre `T + (48,40)`.
- Compounds (HQ, depot, factory) and mountains are painted as one whole scene and sliced into hex crops.
  HQ is a seven-cell rosette (238×240 world px), depot/factory a four-cell diamond (238×160). Roofs, yards
  and shadows stay inside the union outline. Seasons share exact placements.
- Shores: land/water state lives on the shared polygon vertices; each `lakeside*` ID lists its land
  vertices explicitly in `art/shore-terrain-v1.json`.
- Hills select a context by neighbours *and* absolute lattice phase; use the matching context rather than
  repeating the isolated default.
- Terrain sources are 1374×1145 (exact 6:5). Exports are fixed Lanczos3 at 2× (192×160) and 4× (384×320);
  1× is derived. Centre-crop to 6:5, never stretch.

## Pipeline

| Step | Where |
| --- | --- |
| Prompt, inputs and output pins per asset | `art/recipes/**.json` |
| Untouched generated masters | `art/source/` |
| Bundle manifests (one per family) | `art/*-v1.json` |
| Exports (2×/4×, interface also 1×) | `art/exports/`, built by `tools/art-review`, `tools/interface-art`, `tools/unit-panel` |
| Runtime ledgers: every ID → export, status | `runtime-{tile,unit,interface,screen}-coverage.csv` (here) |
| Runtime packs | `go run ./cmd/hlasset build-terrain\|build-units\|build -density N -out build/packs/…` (see `internal/assets`) |
| Browsable gallery with prompts | `gallery/` (`just gallery`) |
| Private comparison desk, composer, PULSE study | `just gfx-game-review`, `just gfx-study` → `build/review/gfx/` |

`asset-inventory.csv` records the legacy assets and `terrain-production.csv` the original terrain resources
still to be covered (both are read by `tools/art-inventory` and `tools/phase2-audit`).

Review real neighbours through the compositor (`internal/mapcompose`), including three-tile vertices:
matching style in isolation does not prove a seamless join. Keep masters, prompts and hashes; export
reproducibility does not mean that a generator rerun gives the same image.

## Status

All 134 enabled tile IDs (60 backgrounds, 74 infrastructure), all 15 unit families with portraits, 117
interface assets and the screen frames exist as **candidates**. Remaining work is tracked in
[PLAN.md](../../PLAN.md): runtime integration, full-map seam review and final visual acceptance.
