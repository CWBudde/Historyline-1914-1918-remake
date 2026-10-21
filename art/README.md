# Art guide

All shipped art is newly authored. Keep the original's gameplay semantics, stable runtime IDs, map
placement, connection topology and familiar information layouts. Never ship original graphics or close
reconstructions of them.

## Layout

| Path | Holds |
| --- | --- |
| `art/<kind>/<family>/` | Selected masters of one art family plus its `prompt.md` |
| `art/style.md` | Shared prompt preamble and style rules per kind |
| `art/pipeline/*.json` | Export manifests (sources pinned by SHA-256, repository-relative) |
| `art/pipeline/contexts/*.json` | Ground, forest and hills context profiles (neighbours, phase, key) |
| `art/reference/pulse/` | PULSE reference design: families, exemplar cells, features and notes |
| `art/fonts/` | Interface typefaces and their licence |
| `art/index.json` | Generated catalogue of all families (`go run ./cmd/hlart index`) |
| `assets/<kind>/<family>/` | Runtime PNGs; `assets/index.json` lists each file with its `art` key |

The art key is the same path in both trees: `assets/index.json` entries with `"art": "terrain/road"`
come from `art/terrain/road/`. Kinds: `terrain`, `units`, `portraits`, `interface`, `screens`. Families
without masters (derived or vector art) still have a folder with `prompt.md`.

## prompt.md

Front matter `id` (the art key), `title`, `kind` (`terrain | unit | portrait | interface | screen`),
`status` (`candidate`: generated, not reviewed yet; `revised`: reworked against the PULSE reference;
`approved`: accepted), `runtime_ids` (terrain/infra IDs, unit types, or interface/screen IDs from
`assets/index.json`) and `masters` (file names in the folder; empty when derived). Then the sections, in this
order: `## Look`, `## Base prompt`, `## Variants` (table `| Variant | Runtime IDs | Prompt delta |`,
`—` for rows without their own ID), `## Edges & joins` (terrain only), `## Post`, `## Avoid`.

To generate: paste the `art/style.md` preamble and the section for the kind, then the base prompt and
one variant's delta. `go run ./cmd/hlart check` validates every prompt, master and runtime ID.

## Geometry

- Logical terrain tile 96×80 at `x = 142*column + 71*(line%2)`, `y = 40*line` (even lines `Width` hexes,
  odd lines `Width-1`, half-shifted; see `internal/hexgrid`).
- Hex polygon `(25,0),(71,0),(96,40),(71,80),(25,80),(0,40)` with complementary half-open ownership at
  pixel centres, so aligned 2×/4× blits give every pixel exactly one owner.
- Ports clockwise: 1 NE, 2 SE, 3 S, 4 SW, 5 NW, 6 N. Neighbour offsets
  `(71,-40),(71,40),(0,80),(-71,40),(-71,-40),(0,-80)`; shared port centres `(83.5,20),(83.5,60),(48,80),
  (12.5,60),(12.5,20),(48,0)`; opposite port `(p+2)%6+1`. Use the explicit port lists in the manifests.
- Unit sprites: 96×96 logical at terrain origin `T + (0,-9)`, anchor `(48,49)` on the terrain centre.
  Six separately painted views `N, NW, SW, S, SE, NE`, never rotations or mirrors; owner palettes `cp`,
  `ec`, `neutral` and the `spent` state are derived.
- Compounds (HQ, depot, factory) and massifs are painted as one scene and sliced into hex crops: HQ a
  seven-cell rosette (238×240 world px), depot/factory/massif a four-cell diamond (238×160). Seasons share
  exact placements.
- Shores keep land/water on the shared polygon vertices; each `lakeside*` ID lists its land vertices in
  `art/pipeline/shore-terrain-v1.json`.
- Ground, forest and hills pick a context by neighbours (hills also by lattice phase) from
  `art/pipeline/contexts/`.
- Terrain sources are 6:5; exports are fixed Lanczos3 at 2× (192×160) and 4× (384×320). Portraits are 105:94
  (210×188 logical). Centre-crop, never stretch.

## Regenerate

`tools/art-review` builds a fresh bundle from a manifest; the output directory must not exist:

```sh
go run ./tools/art-review -connected art/pipeline/connected-terrain-v1.json -out build/review/road
go run ./tools/art-review -unit art/pipeline/unit-infantry-v1.json -out build/review/infantry
```

Modes: `-connected` (road, stream, path, rail, trench), `-mixed` (bridges, crossings), `-shore`, `-compound`,
`-fog`, `-ground`, `-forest`, `-hills`, `-unit`, `-portrait`. Interface, weapons, effects and screens come
from `tools/interface-art` (see its README); the unit panel preview from `tools/unit-panel`. Tagged
generators in `tools/art-review` (`campaigngfx`, `grassgfx`, `riverbankgfx`, `retainedgfx`) rebuild derived
sources and missing contexts. `just gfx-check` runs the art checks.

Review real neighbours through the compositor (`internal/mapcompose`), including three-tile vertices:
matching style in isolation does not prove a seamless join. Export reproducibility does not mean a
generator rerun gives the same image, so keep the selected masters.

## Promote

Copy the reviewed export files into `assets/<kind>/<family>/`, update the master's pin in the manifest when
the master changed, then run `go run ./cmd/hlart index` and `go run ./cmd/hlart check`.

## Reference design

`art/reference/pulse/reference.json` registers the PULSE reference painting (`RefDesign2.png`, local only)
on the first level and names, per terrain family, exemplar cells and notes; `features` describe compounds,
craters, lake, streams and junctions. It is the target look for terrain. Render it with
`go run ./tools/reference-design -out build/review/reference` (tiles under `tiles/rNNcMM.png`;
`-dump-cells` lists each cell's terrain and infrastructure).
