# Phase 2.2 retained candidate coverage audit

Run from the repository root:

```sh
go run ./tools/phase2-audit
go test ./tools/phase2-audit
```

The command writes deterministic JSON and Markdown reports to ignored
`build/phase2-audit/`. `-root` selects the repository; `-out` selects a report
directory contained within it. Any missing file, stale pin, invalid PNG dimension,
duplicate binding, absent enabled ID or ledger/report association fails the command.

The audit compares enabled terrain/infrastructure IDs with literal image-load entries
inside the existing Python loader dictionaries without importing Python. Unit and
weapon coverage comes from embedded Go data. It verifies all unit/portrait recipe
exports and their class/family associations, source/manifest pins, interface/screen
bundle file pins and ledger dimensions. Existing tile manifests provide input pins;
the audit verifies their 2×/4× output dimensions. Tile output byte fidelity and
neighbor/context selection remain the responsibility of their exporter tests.

Counts distinguish logical bindings from authored families and physical PNGs. The
unit `verified_recipe_pngs` total includes candidate planes and review fixtures;
`candidate_pngs`, `neutral_tint_plane_pngs` and `review_fixture_pngs` are separate.
Screen candidate PNGs
count the three frame densities; the JSON breakdown also counts ancillary assets.
Unique PNG/pin totals include retained inputs and fixture PNGs in bundle reports.

When `art/exports/screens-v2/report.json` exists, it is audited as an additional
bundle. Its state-view layouts, widget IDs, vector/density files, source pins and
review fixtures are checked separately from the sixteen v1 screen requirements.
Review compositions never count as candidate artwork. Missing or malformed v2
files fail the audit rather than silently skipping a partially retained bundle.

Native terrain status is counted directly from the separate 350-row resource ledger.
Nine future-background decisions record actual loader/modifier presence; they never
enable content or assume names establish requirements. The known 306 native map
sprite and 49 native illustration obligations are documented separately and are
not inferred from current compatibility unit counts.

A successful report establishes retained offline coverage only. It does not prove
original correspondence, provenance, native directions/factions/anchors, physical
display aspect, visual resemblance, full-map seams or variation, browser behavior,
input flows, animation playback, or runtime acceptance. These remain explicit open
gates in the report and `PLAN.md`.
