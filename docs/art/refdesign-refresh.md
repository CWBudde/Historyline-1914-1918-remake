# Terrain refresh from RefDesign

2026-10-20. [Current PULSE study](../../build/gfx-review/pulse.html) ·
[Committed art gallery](../../gallery/index.html) ·
[RefDesign comparison](../../build/refdesign/reference-comparison.html) ·
[Nine shore variants](shoreline-review.html).

The user-supplied `RefDesign.png` establishes the current target: detailed olive
grass with natural variation, rough muddy earthworks, modest rocky relief,
small timber huts, fine paved streets and irregular woodland canopies. Its
SHA-256 is `3924e55c20ca63cbe31a7e4b3d469977f20c9ddf7802912e92782360b4d1aa69`.
The reference is not included in the runtime artwork.

The refresh includes:

- Sixteen grass arrangements from a new painted meadow master. Summer grass
  now wraps without reflecting its strokes, removing the old symmetric motifs.
  Every variant retains a common boundary band so adjacent terrain still joins.
  A 768×480 crop sets the common feature scale; sixteen source phases vary the
  interiors without enlarging the grass strokes.
  Compatible painted flower/pebble and worn-meadow materials enter selected
  interiors at 65% strength over phased meadow, keeping the boundary band exact.
- A complete circular muddy earthwork with a raised broken rim, dark inner cut
  and grassy centre. The optional `mud_feature: "circular-earthwork"` profile
  preserves the painted crater composition instead of mirroring a mud texture.
- Fine grey asphalt, narrow dashed centre markings, continuous edge lines and
  dusty shoulders, following the existing reciprocal road curves.
- New bank earth, damp trench floor and weathered timber, with material-driven
  irregular trench silhouettes and broken planks inside the protected joins.
- Pale gravel river banks, dark wet lips and intermittent foam, sharing the
  shore-bank source while keeping a continuous water phase through bridges.
  A retained pixel-exact crop makes the river's larger stones readable within
  its wider 3.5-world-pixel bank; the nine lake profiles use the full master.
- Low fractured rocky outcrops and a single compact timber hut, with matching
  winter counterparts, replacing the previous mountain and hamlet masters.
- Warmer, deeper summer foliage with larger crowns inside the established
  atlas guards. The selected summer profile uses thirty 26.5-world-pixel crowns
  with a broadleaf-biased species roster. Actual resampled alpha support must
  fit the original radius bound, retaining the existing seven-cell proof and
  quiet band. Winter keeps its original crown size and positions.

The nine DOS-inspired shore profiles retain their bank shapes and water palette
while using the same new summer grass as the rest of the map. Water, unseen
terrain and winter ground retain their existing sampling method.

Selected sources and full generation records live under `art/source/` and
`art/recipes/` in the `refdesign-v1` directories. The built-in image generator
produced the new raster masters; Go performs source projection, terrain
composition, context selection and density exports. The grass base and sixteen
arrangements combine deterministic meadow phases with the two new painted
detail materials. Their generation records retain the exact edits and
whole-image projection used to keep their palette and feature scale compatible.

Both the game and PULSE map renderer select the same grass variations and
neighbour contexts. Existing campaign-specific ground, forest and hill profiles
are regenerated along with the fixed review fixtures. Unit surround previews
also use the new grass and earthwork ground.

Rebuild the optional grass arrangements with:

```sh
go test -tags grassgfx ./tools/art-review -run TestGenerateGrassVariations -count=1 -v
```

When refreshing a family that already retains additional campaign contexts,
augment its fresh bundle before retaining its PNGs and report:

```sh
go test -tags retainedgfx ./tools/art-review \
  -run '^TestGenerateRetainedContextBundle$' -count=1 -args \
  -context-family ground \
  -context-manifest art/ground-terrain-v1.json \
  -context-recipe art/recipes/ground-terrain-v1.json \
  -context-bundle build/<fresh-ground-bundle>
```

The same command accepts `forest` and `hills`. It checks the manifest snapshot
and retained context identities, then writes only inside the fresh bundle.

After retaining rebuilt art, `just gallery` refreshes the committed public
catalogue and prompt records. `just gfx-study` refreshes PULSE at 1×/2×/4×
and produces actual 720p, 1080p and 4K game captures. Visual resemblance should
be judged in those assembled views; no numerical fidelity percentage is claimed.

Validation for this refresh: all source-package tests and the complete art-review
test selection passed, including all fifteen unit families; lint reported no
issues. The retained-art audit checked 2,927 PNGs and 3,013 hash pins with no
problems. `just gallery-check` reproduced the public gallery exactly. The final
PULSE study uses nine fresh game captures from the selected terrain exports.
