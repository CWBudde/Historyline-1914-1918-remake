# Runtime PNG pack foundation

Phase 2.3 now has a renderer-independent loader in `internal/assets` and local
`hlasset build` / `validate` commands. This first adapter packages the existing
interface candidates; it does not approve artwork or change the game window.

From the repository root, with an existing `build/` directory:

```powershell
go run ./cmd/hlasset build -density 2 -out build/interface-runtime-2x
go run ./cmd/hlasset validate -root build/interface-runtime-2x -allow-candidates
go run ./cmd/hlasset build -density 4 -out build/interface-runtime-4x
```

`-source` defaults to `art/exports/interface-v1`. The adapter accepts only its
schema-1 report and emits a **candidate** manifest plus 117 exact-density PNGs
(118 files). Density may be 1, 2 or 4; every reported runtime ID must provide the
requested density. It copies the verified byte snapshots without resampling or
recompression. It excludes SVG sources, reviews, HTML and authoring metadata.
Each output directory must be new; its parent must exist. Validation and complete
PNG decoding happen before output creation. Failed writes clean up only files
created by the build, and an existing output directory is never replaced.

Separate density packs keep decoding bounded. Keeping all three interface
densities in one pack would exceed the loader's conservative 256 MiB decoded
image budget. That estimate allows eight bytes per pixel for 16-bit RGBA; it is
not a measured total process/GPU memory or performance claim.

## Manifest and loading contract

```json
{
  "schema_version": 1,
  "id": "example-2x",
  "status": "candidate",
  "entries": [
    {
      "id": "map/selected",
      "path": "images/0000.png",
      "sha256": "<64 lowercase hex digits>",
      "width": 240,
      "height": 200,
      "density": 2,
      "logical_anchor": [60, 50]
    }
  ]
}
```

Paths are relative to the supplied filesystem root. `Open` reads this strict,
versioned JSON format; `Load` accepts the same manifest as a Go value. Disk callers
use `os.Root.FS()` to contain filesystem traversal and symlinks. Embedded packs
can use `embed.FS`. Portable slash paths must be local and cannot contain `legacy`
or `original` components. Entry paths are unique, including case-insensitive
collisions; stable `(ID, density)` pairs are unique too.

`Load` checks all catalog geometry before reading images, snapshots each file,
verifies its SHA-256 and PNG header, and checks exact pixel dimensions. Repeated
densities of an ID must agree on logical dimensions and anchor. Invalid hashes,
non-finite/out-of-bounds anchors, unknown densities, invalid paths and missing
files return errors. Limits are 4,096 entries, a 2 MiB manifest, 32 MiB per PNG,
128 MiB of encoded snapshots and the 256 MiB decoded-image estimate above.

`Image(ID, density)` decodes a snapshot once using a concurrent cache, including
cached decode errors. `Validate()` fully decodes all entries in manifest order,
checking the pixel streams and PNG CRCs as well as the headers. `WritePNG` writes
the exact snapshot for reproducible packaging. Callers must treat returned images
as read-only. Manifest access returns a copy, and separate pack instances never
share an image cache. Switching packs therefore preserves IDs without retaining
the previous pack's images. GPU texture caches remain the renderer's responsibility.

There is no implicit density fallback, sprite rotation, recolouring or reference
lookup. An unavailable exact variant returns `*assets.MissingError` identifying
the pack, visual ID and requested density. This leaves zoom and UI scale separate
from image density.

Candidate loading requires `Options{AllowCandidates: true}` or the CLI flag
`-allow-candidates`. The status field is a declared review status, not a signature
or proof of provenance. Changing it does not establish approval. No approved pack
is embedded yet, and original/reference bytes are not copied by this adapter.

## Consumer and verification (2026-10-15)

`cmd/hleffectpreview` now uses the shared loader for its selected interface PNGs,
with explicit candidate opt-in and stable ID/anchor metadata. Animation/binding
metadata retains its existing pin checks. Unit sprites and the grass fixture
retain their previous loader; the command does not yet consume arbitrary packs.

Executed checks:

- `go test ./internal/assets ./cmd/hlasset ./cmd/hleffectpreview -count=1`:
  exact variants, missing-ID errors, immutable metadata/byte snapshots, separate
  caches, concurrent first decode, malformed metadata, density geometry, resource
  limits, corrupt hashes/PNG pixel streams and candidate opt-in.
- Builds from all three retained interface densities preserve all 117 IDs and
  exact source PNG bytes. Two 2× builds repeat all 118 files byte-for-byte.
  Existing-output refusal and invalid-input output absence are exercised.
- `go vet` and `golangci-lint run` for those three packages pass. The race detector
  was unavailable in this environment (`CGO_ENABLED=0`, no C compiler on PATH);
  concurrent cache tests ran under the ordinary test runner.
- The updated effects workbench compiles for `GOOS=js GOARCH=wasm`; browser execution
  and browser filesystem loading are not exercised by that build.
- A real CLI 2× build and full `validate` pass under
  `build/runtime-interface-2x` (ignored).
- The rebuilt native effects workbench passes normal (440 ticks, 25 captures),
  reduced (59 ticks, 5 captures) and skip (111 ticks, 12 captures) checks. Ignored
  evidence lives in `build/runtime-pack-effects-{normal,reduced,skip}/`. Movement
  and registration were visually inspected. Screenshot bytes are not a cross-run
  golden-image contract; input PNG bytes and playback state have their own checks.

## Unit catalog (2026-10-16)

`hlasset build-units` packages the unit sprite exports of all 15 families into a
local candidate catalog that the game can draw:

```sh
go run ./cmd/hlasset build-units -out build/runtime-units
go run ./cmd/hlasset validate -root build/runtime-units/4x/cavalry -allow-candidates
go run ./cmd/historyline -units build/runtime-units -allow-candidates
```

It reads every `art/recipes/unit-*-v1.json` bundle report (`-recipes`) and the
matching `art/exports/unit-*-v1/` directories (`-exports`), and writes one pack per
family and density: `<out>/<d>x/<family>/{manifest.json,images/NNNN.png}`. Each pack
holds 36 sprites (palettes `cp`/`ec`/`neutral` × six facings × `normal`/`spent`),
copied byte for byte; the retained exports give 2× and 4× (30 packs, 1,080 PNGs).
Separate packs keep decoding bounded: all 4× units in one pack would need about
637 MB by the loader's eight-bytes-per-pixel estimate, over its 256 MiB budget,
while one 4× family is about 42 MB.

Stable IDs are `unit/<family>/<palette>/<facing>-<state>`, for example
`unit/light-artillery/ec/nw-spent`. Family slugs are the recipe IDs without `unit-`
and `-v1`; facings are `n`, `nw`, `sw`, `s`, `se`, `ne` (export `facing_index` 0–5).
Every sprite is 96×96 logical pixels with its `logical_anchor` (48, 49) on the hex
centre. `internal/unitsprite` holds the unit-type → family table, the facing,
palette and state names and the ID format, shared by the builder and the game. The
build refuses recipes whose `class_ids` disagree with that table, unknown
palettes/states/facings, duplicate or missing variants, a density missing for any
variant, sizes that are not 96 logical pixels, pin mismatches and a catalog missing
a family. Everything is validated before the new output directory is created.

The game opens the catalog only with `-units`, and candidate packs only with
`-allow-candidates`; the top-right development label then reads `DEV: CANDIDATE-ART`. At load, it checks that each family has every sprite at both
densities. The palette follows the unit's current owner, not its country, so
captured units change colour. Units that have moved or attacked use the `spent`
(greyscale) sprite; remembered enemy units use it dimmed. Facing is presentation
state in the UI: units start facing north, turn along their last move step and
towards the opponent in combat, and the walking unit faces its current step. The
drawn density is the asset density of the effective zoom, at least 2×. Below 2×
the 2× sprite is drawn smaller with linear filtering. Textures are uploaded on
first use and cached per ID and density. Without `-units` the map keeps the
placeholder dots. The browser build cannot read a local catalog.

Not established: approval of any unit art, the native Amiga A–F facing mapping and
the exact physical pivot (both open in the art reports), the reference five-shade
faction swap and an interactive check in the game window.

## Terrain catalog and compositor (2026-10-16)

`hlasset build-terrain` packages the terrain, infrastructure and context exports
into a local candidate catalog, and the game composes each level from it:

```sh
go run ./cmd/hlasset build-terrain -out build/runtime-terrain
go run ./cmd/hlasset validate -root build/runtime-terrain/4x/context-hills -allow-candidates
go run ./cmd/hlasset terrain-gaps -terrain build/runtime-terrain
go run ./cmd/historyline -terrain build/runtime-terrain -units build/runtime-units -allow-candidates
```

The builder reads `docs/art/runtime-tile-coverage.csv` (`-coverage`: one row per
runtime ID with its 2×/4× export) and the recipes (`-recipes`). Every export must
carry a `{path, sha256, dimensions}` pin in the row's recipe, wherever it is nested.
Context tiles come from the `profiles` of the ground, forest and hills recipes. It
writes five groups per density, `<out>/<d>x/<group>/`:
- `backgrounds` (60 tiles, including the fog tile `unseen`);
- `infrastructure` (74);
- `context-ground` (23), `context-forest` (20) and `context-hills` (121).

That makes 10 packs and 596 PNGs. All 4× terrain in one pack would be about 280 MB
by the loader's estimate, over its 256 MiB budget. Tiles are 96×80 logical
pixels, anchored (48, 40) on the hex centre. IDs are `tile/<runtime ID>` and
`context/<family>/<context key>`. The build refuses:
- a row without a pin or without a density;
- a size or name that disagrees with the pin, and pin mismatches;
- a context profile whose descriptor or key does not follow the contract below.

`internal/mapcompose` turns a level into what each hex draws: a background, then
the stored infrastructure overlays in order. Mud, stone and the winter rough
grounds (`ground`), forest and hills look different depending on their neighbours.
For those it builds the offline exporters' descriptor (season, centre, the six
neighbours NE, SE, S, SW, NW, N, and for hills the tile origin modulo 568×320) and
uses the context tile whose key is the SHA-256 of that descriptor.
`TestDescriptorMatchesRecipeProfiles` reproduces all 164 recipe keys.

Where no context tile exists, the hex shows its default tile and the compositor
records a gap. A gap is either:
- **missing**: the contract defines the context, but it was not exported;
- **undefined**: a neighbour from another family or season, or the map edge, for
  which the contract has no rule.

`hlasset terrain-gaps` lists every gap with its coordinates and descriptor for the
art track. Summary for the shipped levels:

| Level | Context tiles selected | Missing | Undefined |
| --- | --- | --- | --- |
| Level 1 | 5 | 9 | 142 |
| Level 2 | 35 | 48 | 150 |
| Level 3 | 19 | 13 | 143 |

HQ, depot, factory and mountain pieces must form complete groups (rosette or
diamond, per `docs/art/compound-building-geometry.md`). A stray piece is reported
and drawn as it is, never completed. A mountain quarter outside a complete group
is shown as the whole `mountain`/`mountain_w`, for display only (level 3, line 21,
col 6).

The game opens the catalog only with `-terrain`, and candidates only with
`-allow-candidates`. At load it composes the level and checks that every tile it
draws, plus `unseen`, exists at both densities. The development label adds
`CONTEXT-GAP n` (and `COMPOUND-ISSUE n` when there are stray pieces).

Drawing works as follows:
- **Passes:** all backgrounds first, then the overlays, then owner bars, then
  units, so a later hex never covers anything drawn beyond an earlier one's
  outline.
- **Fog:** unseen hexes show the fog tile, and seen (remembered) hexes are drawn
  grey and darkened.
- **Density:** the same rule as for units, at least 2×.
- **Seams:** the tiles own complementary pixels, so a tiled field is exactly
  opaque. Scaled, filtered quads do not meet exactly, because Ebitengine snaps quad
  corners to a sub-pixel grid. That left thin seams at fractional zoom, which
  additive blending could not fix either. Each background texture therefore gets
  a 3-texel border into which its opaque edge is grown, so neighbours overlap.
  Overlays keep a transparent border.

Without `-terrain` the map keeps the coloured hexes. The browser build cannot
read a local catalog.

Not established:
- approval of any terrain art;
- exports or rules for the missing and undefined contexts;
- the HQ entrance cell, the `trench5a` port and the inferred lakeside corners
  (open in `docs/art/`);
- ownership strips;
- an interactive check in the game window. Screenshots at zoom 0.68, 1.0, 1.77
  and 3.8 show no seams.

## Remaining Phase 2.3 work

Add portraits, screen layouts/fonts and animation metadata to purpose-built
runtime catalogs (unit sprites and terrain have one, above). Close the terrain
context gaps listed by `hlasset terrain-gaps`. Add reference compatibility transforms
separately, and embed only approved exports through `assets/assets.go`. The
`hlasset review` command, `art/manifest.json` integration, live pack switching in
the game, GPU caches beyond the unit textures, full-map acceptance, browser
playback and measured performance remain open.
