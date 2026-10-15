# Runtime tile coverage

[runtime-tile-coverage.csv](runtime-tile-coverage.csv) records the enabled remake
loader contract: **60 background IDs and 74 infrastructure IDs**, 134 total.
The authoritative dictionaries are `bgTiles` and `infraIcons` in
[hlrData.py](../../hlrData.py:83), not filenames inferred from the native library.

The independent coverage audit inspected both dictionaries, the inventory,
aliases, rule data and campaign files. Controller extraction and verification
confirm every ledger ID/path and the campaign-use flags.

## Status

Fourteen backgrounds have a `candidate-source` association:
`forest`, `forest_w`, `water`, `waterStones`, `mud`, `stone`,
`hills`, `mountain`, `mountain_w`, `house`, `house_w`, `rough_w`, `stoney_w`, `shelled_w`.
These point to generated art and 2×/4× rectangle exports, **not approved runtime
bindings**. Family geometry, shared phase and scale/quality review remain open;
forest and mountain in particular retain recorded quality caveats.

Summer/winter ground, eight depot pieces, fourteen HQ pieces, four factory pieces
and all 74 enabled infrastructure IDs are
`candidate-built`, with `patch-reviewed` seams and explicit component/traversal or
shared-scene geometry. These 102 rows point to their respective build manifests and
192×160 / 384×320 polygon exports. Actual saved-PNG patches pass both parities and
densities; independent [road/stream review](connected-family-batch-01.md) and
[path review](distinct-terrain-batch-04.md), [rail review](distinct-terrain-batch-05.md)
and [trench review](distinct-terrain-batch-06.md), plus
[bridge/crossing review](distinct-terrain-batch-07.md), provisionally pass. The
[depot review](distinct-terrain-batch-08.md) verifies complete summer/winter scenes,
actual cropped PNG assembly and surrounding ground at both densities/parities.
The [HQ review](distinct-terrain-batch-09.md) adds seven-cell seasonal scenes, with
the manor's lower-N/entrance-in-C placement recorded for runtime usability review.
The [factory review](distinct-terrain-batch-10.md) adds all four winter pieces,
with the eastern loading wing associated with gameplay cell E. All 26 enabled
building IDs now have provisional candidates from five whole-scene masters.
Full maps, transitions, seasonal integration, fractional zoom and runtime/release
acceptance remain open. The other 18 rows say `needs-image`; all are backgrounds.
Total: **116 associated IDs, zero accepted runtime replacements**.

The `build_manifest` column names the producing manifest; `ports` records the
declared directions for built overlays. Source candidates are not interchangeable
with built runtime candidates, and `patch-reviewed` is not a release acceptance.

Twenty-seven semantic/material source images do not mean twenty-seven replaced runtime IDs: dry meadow
and dense/sparse forests are variations; bare earth is a material variant;
the opaque road/river paintings remain family studies. Separate new material inputs
feed the compositors, which author transparent road/stream/path/rail/trench/bridge/crossing overlays.
Trench `ports=none` identifies the closed ring; blind spurs/dugouts are declared
separately in its manifest. The provisional 5a interpretation remains explicit.
Mixed rows use typed port strings: `path:2|5`, or `road:1|4;rail:3|6`, rather
than merging two different routes into one junction. Finite road-bridge structures
use `traversal:1|4;overlay:none`: they have no painted boundary contacts and require
a separate upper route. Runtime IDs retain `rBridge` case; their export IDs are lowercase.
Depot `cmpN/W/E/S` and winter counterparts are cropped from one new master per
season, with fixed origins and a quiet shared-ground boundary. `grass_w` is now a
polygon-built winter ground candidate instead of the earlier rectangle association.
HQ `hqN/NW/NE/C/SW/SE/S` and winter counterparts use their own shared-scene
manifest, with fixed seven-cell origins and the same complementary polygon ownership.
Factory `factoryN_w`, `factoryW_w`, `factoryE_w`, `factoryS_w` share a single winter
master and fixed four-cell origins, with E retained as the entry cell.

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
