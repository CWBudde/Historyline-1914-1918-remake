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
faction swap, terrain composition and an interactive check in the game window.

## Remaining Phase 2.3 work

Add portraits, screen layouts/fonts and animation metadata to purpose-built
runtime catalogs (unit sprites have one, above). Integrate accepted terrain with
the shared arbitrary-map/context compositor rather than treating default terrain
previews as unconditional replacements. Add reference compatibility transforms
separately, and embed only approved exports through `assets/assets.go`. The
`hlasset review` command, `art/manifest.json` integration, live pack switching in
the game, GPU caches beyond the unit textures, full-map acceptance, browser
playback and measured performance remain open.
