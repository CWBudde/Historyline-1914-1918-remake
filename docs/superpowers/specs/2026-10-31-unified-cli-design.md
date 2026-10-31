# Unified tooling CLI

The developer tools have one entry point, `hl` (`cmd/hl`). Preserve every
existing capability rather than discard specialist tools. The game remains
`cmd/historyline`. No art, assets, scenarios or generated catalogues change as
part of this migration.

## Commands

| Command | Implementation migrated from |
| --- | --- |
| `hl art index/check/thumbs/pulse/map` | `cmd/hlart` |
| `hl assets build/build-units/build-terrain/build-screens/packs/lite/terrain-gaps/validate` | `cmd/hlasset` |
| `hl export materials/kits` | `cmd/hlmaterial`, `cmd/hlkit` |
| `hl export terrain-stamps/unit-topview/unit-squad/unit-panel/interface` | Corresponding `tools/` packages |
| `hl export generate NAME` | Explicit tagged test generators in `tools/` |
| `hl review art/reference/candidates/gallery` | `tools/art-review`, `tools/reference-design`, `tools/candidate-review`, `cmd/hlgallery` |
| `hl levels extract/campaign` | `cmd/finextract`, `cmd/hlcampaign` |
| `hl levels backdrop jpeg/check` | `cmd/hlbackdrop` |
| `hl llmplay new/prompt/step/run/replay/render/review/gif` | `cmd/hlllplay` |
| `hl preview effects` | `cmd/hleffectpreview` |
| `hl serve` | `cmd/serve` |

## Structure and behaviour

`internal/hlcli` dispatches commands to importable Go packages. Existing
`cmd/hl*` implementations move to `internal/` at the same directory depth;
the offline `tools/` packages become libraries. Preserve existing flags,
defaults, output formats, fresh-output protections and LLM rejection exit
code 2. Help succeeds at every command depth; errors return nonzero without
exiting inside a library. No additional dependencies are needed.

The native preview stays in a separate executable so Ebitengine cannot
initialise a display during ordinary CLI commands or help. Prefer its sibling
executable; in a source checkout, launching it through Go is supported.
Tagged generators remain repository-only operations using Go. They are
listed with their purpose and run only when explicitly named. Verification
must not invoke generators that replace art or assets.

Candidate review calls the export libraries directly instead of building or
launching the previous tools. Update just recipes, CI, documentation and
subprocess tests to use `hl`. Keep existing tests alongside their code,
including fixtures and embedded review pages. Keep all caches off the project
drive. Commit the migration directly to `main`, preserving other work.

## Verification

Exercise help and errors for every leaf command with no display, real
catalogue/pack validation and an LLM session with rejected and accepted orders.
Run existing tests, lint, desktop and WASM builds, art and reference checks.
Review path-sensitive fixtures, build tags and subprocess callers after moves.
