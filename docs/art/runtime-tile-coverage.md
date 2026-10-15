# Runtime tile coverage

[runtime-tile-coverage.csv](runtime-tile-coverage.csv) records the enabled remake
loader contract: **60 background IDs and 74 infrastructure IDs**, 134 total.
The authoritative dictionaries are `bgTiles` and `infraIcons` in
[hlrData.py](../../hlrData.py:83), not filenames inferred from the native library.

The independent coverage audit inspected both dictionaries, the inventory,
aliases, rule data and campaign files. Controller extraction and verification
confirm every ledger ID/path and the campaign-use flags.

## Status

Twelve backgrounds have a `candidate-source` association:
`grass_w`, `forest`, `forest_w`, `water`, `waterStones`, `mud`, `stone`,
`hills`, `mountain`, `rough_w`, `stoney_w`, `shelled_w`.
These point to generated art and 2×/4× rectangle exports, **not approved runtime
bindings**. Family geometry, shared phase and scale/quality review remain open;
forest and mountain in particular retain recorded quality caveats.

Grass plus all 16 road and 11 stream IDs are `candidate-built`, with `patch-reviewed`
seams and explicit ports. These 28 rows point to the new connected manifest and
192×160 / 384×320 polygon exports. Actual saved-PNG patches pass both parities and
densities; [independent review](connected-family-batch-01.md) provisionally passes.
Full maps, transitions, seasonal integration, fractional zoom and runtime/release
acceptance remain open. The other 94 rows say `needs-image`.
Total: **40 associated IDs, zero accepted runtime replacements**.

The `build_manifest` column names the producing manifest; `ports` records the
declared directions for built overlays. Source candidates are not interchangeable
with built runtime candidates, and `patch-reviewed` is not a release acceptance.

Nineteen semantic/material source images do not mean nineteen replaced runtime IDs: dry meadow
and dense/sparse forests are variations; bare earth is a material variant;
the opaque road/river paintings remain family studies. Separate new material inputs
feed the shared compositor, which authors the transparent road/stream overlays.

The table preserves all required loadable IDs, including editor-only entries.
The three supplied missions place 58 background and 70 infrastructure IDs.
`lakeside12`, `path125`, `path135`, `trench125`, `rBridge14` remain loadable
despite no campaign placement. `unseen` is needed for fog. The `<none>`/empty
infrastructure sentinels need no image. Twenty-three filename spelling aliases
remain separate stable IDs; they are included in 134.

## Boundaries

- The existing [350 native-resource inventory](terrain-production.csv) remains a
  separate summer/winter coverage and mapping obligation. Completing 134 bindings
  alone cannot close that checklist or claim full original-content coverage.
- `depotOwnership.png` is an additional board ownership stripe image, outside the
  stable terrain/infra ID count. It still needs a new treatment.
- Units, portraits, UI frames, markers, effects and full-screen illustrations have
  their own existing inventory and remain part of overall graphics completion.
- Summer factories, barbed wire, terrain `trenches`, and summer rough/shelled/stoney
  are recorded future-content gaps in [inventory notes](inventory-notes.md).
  They are not extra enabled loader IDs and must not be silently substituted.
- Filename digits alone are insufficient port evidence. [Geometry research](runtime-connection-geometry.md)
  confirms the six directions and records exceptions. Unbuilt connected replacements
  and shores retain `ports=unverified` until their own declarations/artwork are tested.

Keep the CSV up to date when assigning generated art. Require explicit quality,
geometry and seam evidence before moving any row to accepted.
