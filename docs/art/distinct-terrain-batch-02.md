# Distinct terrain batch 02

2026-10-15. [Gallery](distinct-terrain-review.html) ·
[Source manifest](../../art/terrain-distinct-v1.json) ·
[Runtime coverage](runtime-tile-coverage.md).

Six more newly authored 1374×1145 sources extend the pack to **twelve semantic
candidates**, with 24 tracked Lanczos3 exports at 192×160 / 384×320.
All inputs are newly generated family art. No original image, old reconstruction
master or original raster mask participates. Exact prompts, input/source/output
hashes and revision notes are retained in the six new recipes.

## Independent visual review

Subagent `distinct_batch02_art_review` inspected all six full sources and their
192×160 exports, compared meadow/river material references, then inspected both
corrected sources and their final reductions. The controller inspected all sources
and the final hill/mountain exports too.

| Subject | Selected source | Review result |
| --- | --- | --- |
| Open water | water-open.png | Candidate; consistent slate/teal currents, no isolated distracting highlight |
| Bare earth | earth-bare.png | Candidate; clearly brown ground, restrained fine detail |
| Snow ground | snow-ground.png | Candidate; readable snow, pale-unit contrast still needs an overlay check |
| Rocky ground | ground-rocky.png | Candidate; stone patches distinct from earth/meadow; unit contrast remains untested |
| Grassy hills | hills-grassy-02.png | Candidate after correction; broad connected ridge and saddles now survive reduction |
| Mountain | mountain-single-02.png | Provisional; improved aerial ridge network, but some oblique escarpment cues and extensive grass remain |

The first mountain had strongly oblique stacked peaks/cliffs; it is excluded.
A fresh generation using the new meadow style produces a more overhead ridge plan.
The first hills read as meadow mottling at small size; a focused edit creates a
broader connected ridge/saddle pattern. Rejected attempts remain in ignored
`build/gfx-research/distinct-v1/`; their hashes and prompts are recorded in
selected recipes. No old source pixels were repainted by a local script.

Do not silently promote the mountain to final acceptance. Compare it with game
units and adjoining terrain to decide whether it needs more rock coverage and a
stricter nadir camera. Both relief subjects need variation to reduce repeated landmarks.
No item has passed an assembled hex-seam or runtime acceptance review.

## Reproduction and verification

```sh
GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -pack art/terrain-distinct-v1.json \
  -out build/art-review/terrain-distinct-v1-batch02-final
```

Choose a fresh output name if it exists. The unchanged pack exporter pins exact
center crop → 384×320 → 192×160 with Lanczos3. All twelve current sources are
already exact6:5, so crops retain every pixel. Exports preserve rectangular
overscan; clipping and texture/port continuity remain future compositor work.

Two final builds produce 38 byte-identical artifacts: twelve source copies,
24 exports, a report and a gallery. Controller checks cover all manifest/recipe/
source/export hashes, dimensions, full opacity, unchanged crop bounds, tracked
export bytes and gallery links. The seven candidate associations in the new
134-row runtime ledger are checked against actual loader IDs/paths, campaign
usage and declared new sources. Repository `just ci` passes.

The first baseline found an unformatted ignored Go verifier left from batch01.
It was formatted, the full baseline passed, and scratch verification now lives
under an underscore-prefixed ignored directory to avoid inclusion by Go's `./...`.

## Coverage and next batch

The new ledger contains all 60 loaded backgrounds and 74 loaded infrastructure
IDs; seven backgrounds have provisional candidate-source associations and
127 IDs still need images. **Zero are accepted runtime replacements.**
Bare earth is a material source, not a finished wet-mud replacement. Alternate
meadow/forest candidates do not create extra runtime IDs.

This does not close the separate 350 native summer/winter resource checklist,
mapping, future-content gaps, sprites or UI. Continue missing winter forest,
wet mud, exposed water stones, winter rough/stone/shelled ground and houses.
Then build shared port/bank geometry and actual connected previews for roads,
streams and shores before completing those directional families.
