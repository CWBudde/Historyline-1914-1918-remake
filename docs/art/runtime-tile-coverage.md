# Runtime tile coverage

[runtime-tile-coverage.csv](runtime-tile-coverage.csv) records the enabled remake
loader contract: **60 background IDs and 74 infrastructure IDs**, 134 total.
The authoritative dictionaries are `bgTiles` and `infraIcons` in
[hlrData.py](../../hlrData.py:83), not filenames inferred from the native library.

The independent coverage audit inspected both dictionaries, the inventory,
aliases, rule data and campaign files. Controller extraction and verification
confirm every ledger ID/path and the campaign-use flags.

## Status

Seven backgrounds have a `candidate-source` association:
`grass`, `grass_w`, `forest`, `water`, `stone`, `hills`, `mountain`.
These point to generated art and 2×/4× rectangle exports, **not approved runtime
bindings**. Original masks, hex clipping, shared phase and scale/quality review
remain open; forest and mountain in particular retain recorded quality caveats.
The other 127 rows say `needs-image`. Every seam review is pending.
No completed replacements are claimed.

Twelve semantic source images do not mean twelve replaced runtime IDs: dry meadow
and dense/sparse forests are variations; bare earth needs separate wet-mud work;
the current opaque road/river paintings are family studies rather than runtime
infrastructure overlays. Their ports and ground separation need shared geometry.

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
- Filename digits are insufficient port evidence. `ports=unverified` keeps every
  connected overlay and shore open until tested through actual geometry.

Keep the CSV up to date when assigning generated art. Require explicit quality,
geometry and seam evidence before moving any row to accepted.
