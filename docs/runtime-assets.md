# Runtime asset packs

`internal/assets` loads versioned PNG packs; `cmd/hlasset` builds and checks them from the runtime art in
`assets/`. `assets/index.json` lists every runtime file (ID, pack, density, path below `assets/`, SHA-256,
size, anchor, art key such as `terrain/road`); packed entries are in pack order. Packs live in ignored
`build/packs/` until art is approved; only approved art will be embedded (`assets/assets.go`).

```sh
go run ./cmd/hlasset build -density 2 -out build/packs/interface-2x        # interface (1, 2 or 4)
go run ./cmd/hlasset build-units   -out build/packs/units                  # 15 unit families
go run ./cmd/hlasset build-terrain -out build/packs/terrain                # tiles, overlays, contexts
go run ./cmd/hlasset validate -root <pack> -allow-candidates
go run ./cmd/hlasset terrain-gaps -terrain build/packs/terrain
just play                                                                  # game with both catalogs
```

## Pack format

`manifest.json` (`schema_version` 1, `id`, `status`, `entries[]`); each entry has `id`, `path`, `sha256`,
`width`, `height`, `density` and `logical_anchor`. `Load` checks paths (local, unique, no `legacy`/`original`
components), hashes, PNG headers and exact dimensions before use; all densities of an ID share logical size
and anchor. Limits: 4,096 entries, 2 MiB manifest, 32 MiB per PNG, 128 MiB encoded, 256 MiB decoded
(8 bytes/pixel estimate), so each density and family is its own pack.

- No implicit fallback: a missing `(ID, density)` is an `*assets.MissingError`. No rotation or recolouring.
- `candidate` packs load only with `Options{AllowCandidates: true}` / `-allow-candidates`; the game then
  shows `DEV: CANDIDATE-ART`.
- Builders validate everything before creating the output directory, never overwrite an existing one, and
  copy export bytes unchanged.

## Units

IDs `unit/<family>/<palette>/<facing>-<state>` (palettes `cp`/`ec`/`neutral`, facings `n nw sw s se ne`,
states `normal`/`spent`): 36 sprites per family, 96×96 logical, anchor (48,49). `internal/unitsprite` holds
the unit type → family table and the ID format. The palette follows the current owner; moved/attacked units
use `spent`; remembered enemies are dimmed. Facing is UI state (start north, turn along moves and towards
the opponent).

## Terrain

Groups per density:
`backgrounds` (60, incl. fog tile `unseen`), `infrastructure` (74), `context-ground`, `context-forest`,
`context-hills`. IDs `tile/<runtime id>` and `context/<family>/<key>`; 96×80 logical, anchor (48,40).

`internal/mapcompose` turns a level into per-hex draws: background, then stored overlays in order. Ground
materials, forest and hills pick a context tile whose key is the SHA-256 of a descriptor (season, centre,
neighbours NE SE S SW NW N, plus tile origin mod 568×320 for hills). Neighbours from another family count as
the seasonal base (`grass`/`grass_w`); outside the map the centre continues. Without a context tile the
default tile is drawn and a gap is recorded (`missing` or `undefined`). The three shipped levels have no gaps.
Compounds must form complete groups; stray pieces are reported, never completed.

Drawing order: all backgrounds, overlays, owner bars, units. Unseen hexes show `unseen`, remembered hexes
are greyed. Draw density is the effective zoom's asset density, at least 2×. Background textures get a
3-texel grown border so filtered quads at fractional zoom overlap without seams.

## Open

Packs for portraits, screens, animation metadata (all listed in `assets/index.json`) and fonts;
ownership strips; contexts for imported maps and mixed seasons; live pack switching; browser loading;
art approval and embedding.
