# Retained Phase 2.2 coverage audit

`go run ./tools/phase2-audit` writes deterministic JSON and Markdown to
`build/phase2-audit/` and fails on missing files, stale pins, wrong dimensions,
duplicate bindings or incomplete enabled-ID coverage. It reads the current Python
loader dictionaries statically and compares unit/weapon IDs with embedded Go data.
It never imports Python, enables a future type or promotes an acceptance status.

The audit run for this continuation verified **2,927 unique PNG headers/dimensions
and 3,011 SHA-256 pins with zero problems**:

| Scope | Coverage established |
| --- | --- |
| Enabled tiles | 134 bindings: 60 backgrounds and 74 infrastructure IDs, 268 default 2×/4× PNG references. General context selection remains separate. |
| Enabled units | 27 IDs across 15 visual families: 1,080 tactical images and 45 portraits. |
| Unit support/review | Another 360 neutral/tint planes and 420 review PNGs; the 1,905 recipe-export total includes these and is not the runtime-image count. |
| Interface | 117 stable IDs and 351 three-density candidate PNGs. |
| First screen batch | Sixteen separate views, 48 frame PNGs and sixteen labelled review compositions. |
| Screen-state continuation | Twenty state views, 61 widget IDs, four progress/scroll components and reused controls/countries: 276 vector PNGs; twenty labelled reviews are separate. |

Unique PNG/pin totals also include retained inputs and supporting bundle artifacts.
The two-density painted backdrop is an ancillary image, distinct from frame coverage.
Every runtime-acceptance count remains **zero**.

## Native work is a separate obligation

The native terrain ledger contains **350** records: 346 pending, two candidate
and two style-approved, with four recipe assignments. All 175 winter records
remain pending. Enabled terrain associations cannot close those records.
Native sprite direction/faction/anchor mapping still covers 306 map-sprite records
and 49 unit illustrations. Their resource variants are not counts of gameplay types.
Additional original aircraft, armour, specialists, naval variants and campaign
illustrations await verified content IDs and correspondence in Phase 5/G0.9/G0.10.

The nine future backgrounds `factoryN`, `factoryS`, `factoryE`, `factoryW`,
`barbedWire`, `trenches`, `rough`, `shelled` and `stoney` are absent from the literal
background loader. None is infrastructure-loadable under those exact IDs.
Only `barbedWire` and `trenches` have modifier entries. Existing winter variants
are separate enabled IDs. Names and modifiers do not establish a requirement
to author or enable these backgrounds. Their product/content decision stays open.

## Verification limits

Tile manifests pin sources and this audit checks default output dimensions; their
exporter tests own byte fidelity and neighboring context behavior. Unit recipes and
interface/screen reports provide output pins. PNG-header checks do not decode
every terrain pixel or judge seams, lighting, anatomy, readability or resemblance.
Focused tests exercise corrupt pins/dimensions, escaping paths, duplicate literal
IDs and a corrupt optional screens-v2 bundle. These checks establish saved-file
coverage, not provenance, original fidelity, browser behavior or runtime acceptance.

Run `go test ./tools/phase2-audit` when changing the auditor. Its README describes
the optional new-screen-bundle checks and exact count meanings.
