# Reference-led grass and compound redraw

2026-10-16. This revision follows the user's review of the assembled graphics.
The prior HQ candidate read as a group of houses instead of a fortress, the
summer grass was too flat, and the depot/factory approaches were missing.
These are newly painted, reference-conditioned candidates. User visual approval
and the requested greater-than-90% small-size resemblance remain open.

The selected masters are under `art/source/terrain/reference-v2/` and
`art/source/compound/reference-v2/`. Matching generation recipes retain the exact
built-in imagegen prompts, reference roles and hashes. Original images remain
reference-only in `legacy/` or ignored `build/`; no original reference assembly
is included in the authored export packs.

| Current asset | Selected source | Result |
|---|---|---|
| Summer grass | `meadow-03.png` | Muted olive/lime cushions, soft relief and fine grass blades |
| HQ summer/winter | `hq-summer.png`, `hq-winter.png` | Four pale stone towers and enclosed ramparts, red north gate, north road, water/bridge, two rear bushes |
| Depot summer/winter | `depot-summer.png`, `depot-winter.png` | Fenced medical camp: L-shaped shelter/red cross, three white tents, southern annex, north road |
| Factory winter | `factory-winter.png` | Dense industrial roofs, tanks/basin and a street exiting to the right/east |

HQ master edits are retained in numbered recipes; only the selected unnumbered
summer/winter masters feed the active manifest. Grass pass 03 supersedes pass 02
for integration. The comparison desk selects the active exports, not those studies.

## Building geometry and framing

All seven HQ pieces in each season, all four depot pieces in each season, and all
four enabled winter factory pieces come from one coherent painting per scene.
The existing 238×240 HQ and 238×160 depot/factory geometry and stable gameplay IDs
are unchanged. The factory exit is east/right; HQ and camp approaches are north.

The active scenes explicitly select `"framing": "reference-layout"` and
`"inset": 1`. The original default compound framing still rejects opaque artwork
within eight world pixels of the outer footprint and fades its translucent fringe.
That rule forced reference roads/fences inward and therefore cannot reproduce the
requested edge-reaching approaches. The new opt-in preserves the full unmasked
affine artwork, records boundary diagnostics without calling them a safety pass,
and filters before the existing final union clipping and complementary ownership.
There is no internal tile-edge feathering. Other compound families retain the
default framing; the new mode is restricted to HQ, depot and factory.

The boundary report intentionally records opaque samples inside the former guard
band and, occasionally, just outside the final footprint. This does not claim that
the old guard passed. Whole scenes, the actual decoded tile PNG assemblies and both
surrounding line parities are exported at 2×/4×. Exact tile reassembly tests and
explicit north/east road-edge tests establish the crop/ownership behavior.

## Grass and resemblance

The new source is original-size-reference-conditioned painted grass rather than
uniform fine noise. The existing continuous sampler still mirror-wraps the source
over (142,80) and averages the two lattice cosets separated by (71,40). This helps
shared boundaries but changes the original clump arrangement. The sampler was
kept consistent across terrain, connected infrastructure and compound grounds.

Advisory measurements on the original-size opaque grass pixels illustrate both
the improvement and the remaining limit:

| Measurement | Original | Prior exported grass | Selected pass 03 export |
|---|---:|---:|---:|
| Luminance mean | 118.60 | 108.14 | 115.36 |
| Luminance standard deviation | 7.35 | 4.65 | 7.30 |
| Uniform 7×7 local SSIM against original | — | 0.582 | 0.472 |

The new export restores local contrast and brightness, but the SSIM result also
shows that the arrangement of clumps is not an exact match. Its direct-source
Lanczos comparison scores about 0.597; the shared sampler changes that structure.
These diagnostics are not percentages of visual likeness. A naive RGB-MAE score
already rated the rejected flat grass above 96%, so it cannot establish the user's
target. No greater-than-90% visual approval is claimed for this delivery.

## Review and reproduce

Regenerate the current comparison desk and editable collection page:

```sh
go run ./cmd/hlgfxreview
```

Open `build/gfx-review/index.html` for the grass at original size, or
`build/gfx-review/compose.html` for HQ, depot and factory templates in context.
Collections use stable IDs, so existing saved scenes pick up the selected exports.
Use side-by-side/wipe, both seasons, and 1× versus 2×/4× source inspection.

The existing exporters consume the active manifests, for example:

```sh
go run ./tools/art-review -connected art/connected-terrain-v1.json -out build/new-connected-review
go run ./tools/art-review -compound art/compound-hq-v1.json -out build/new-hq-review
go run ./tools/art-review -compound art/compound-depot-v1.json -out build/new-depot-review
go run ./tools/art-review -compound art/compound-factory-v1.json -out build/new-factory-review
```

Each output directory must be fresh. Grass-dependent producing manifests and
recipes are rebuilt with the selected source; shared output bytes must agree.
Old custom context fixtures that were not recreated remain explicitly historical,
with their original authored input snapshots/hashes retained rather than repinned
to new images. Full-map transitions, arbitrary neighbor context and runtime visual
acceptance remain open.

## Executed verification

- Rebuilt all 30 affected terrain/compound/unit-fixture bundles. Before changing
  inputs, the same staging process reproduced every baseline PNG byte-for-byte.
- Verified and promoted 915 files: 30 manifests, 31 recipes, 564 candidate export
  PNGs, 272 review PNGs and 18 historical authored input snapshots. All destinations
  are under `art/`; no original reference image bytes were promoted.
- Fresh `tools/phase2-audit` run: 2,927 PNG checks and 3,011 hash pins, zero problems.
- Focused compound framing/guard/manifest/actual-PNG/resource tests pass, as do
  `cmd/hlgfxreview`, `cmd/hleffectpreview` and coverage-auditor tests. Browser checks
  cover the single-image desk and editable collections at desktop/narrow widths.
- Vet and lint pass for the changed graphics tooling. Independent final-size
  review finds the requested fortress/camp/east-exit structures readable without
  obvious clipping; both seasonal HQ assemblies rejoin exactly from actual PNGs.

The refreshed review pages still contain 390 selected entries and 3,434 unique
image snapshots. This evidence establishes rebuilding, coverage and placement,
not the unapproved visual resemblance or gameplay/runtime acceptance gates.
