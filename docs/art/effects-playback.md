# Native candidate motion and effect playback

`cmd/hleffectpreview` now exercises the delivered graphics in a bounded native
Ebitengine renderer. Its reusable `internal/presentation` scheduler consumes
copied visual steps using caller time, independently of the rules engine.
The [90-view montage](effects-playback-contact.png) shows all fifteen families
at six movement midpoints. [Retained evidence](effects-playback-evidence.json)
records every run, exact input fingerprints and the montage hash.

## Exercised sequence

Six independently authored CP normal-state facings move in order through
n/nw/sw/s/se/ne, followed by hit, all nine death/explosion frames, capture,
round and completion. Unit anchor (48,49) stays fixed relative to path positions;
terrain centre is (48,40). The preview never rotates or reflects the sprites.
Hit count, capture message and round labels remain live fixture text.

The three modes finish without looping:

| Native runs | Ticks per run | Screenshots per run | Result |
| --- | ---: | ---: | --- |
| Normal, all 15 families | 440 | 25 | All six headings, their early samples/midpoints, nine explosion frames and every event stage captured. |
| Reduced, infantry | 59 | 5 | Movement snaps; reduced hit ring and final smoke alternative; capture/round and completion captured. |
| Skip, infantry | 111 | 12 | Active steps skip after 150 ms, preserve destination/next event, finish once. |

The retained summary covers seventeen runs and **392 actual GPU screenshots**;
375 are from the fifteen-family normal matrix. Full-size captures and individual
evidence files remain under ignored `build/effects-families-{slug}-20261015/`,
`build/effects-reduced-verified-20261015/` and
`build/effects-skip-verified-20261015/`. Early samples are the first captured
render within a heading, not guaranteed exact time-zero frames. Coverage assertions
fail if a slow display omits a required event/frame or movement midpoint.

## Reproduction

```sh
go test ./internal/presentation ./cmd/hleffectpreview
go build -o bin/hleffectpreview.exe ./cmd/hleffectpreview
bin/hleffectpreview.exe -family infantry -self-check -mode normal -out build/effects-fresh
```

Use `-family` for any of the fifteen slugs listed in the command README.
Output must be fresh under an existing parent. Omit `-self-check` for the
interactive review: Space skips, R toggles reduced motion/restarts, Enter restarts,
P pauses, Escape exits. Automated runs do not certify those physical keyboard flows.

The loader checks the interface report, animation/binding hashes, frame dimensions
and anchors, six-view family manifest binding, and the pinned grass image.
Unit output bundles have no output report consumed by this viewer: it checks their
dimensions and records exact byte fingerprints. The separate coverage auditor
verifies those outputs against retained recipe pins. Reads are repository-contained;
no legacy or original bytes are loaded, embedded or copied into evidence.

Scheduler tests cover exact frame/event boundaries, update-frequency independence,
copied inputs, all six headings, negative/oversized time rejection, non-skippable
effects, destination-preserving skip and reduced alternatives. Loader tests exercise
all fifteen six-view packs, changed frame bytes, invalid dimensions, escaping paths
and family selection. Focused tests, vet, build and lint pass.

## Acceptance limits

This small grass fixture and scripted sequence establish native candidate playback.
Moving a bunker, train or boat across this fixture is a registration diagnostic;
it asserts no movement legality or terrain suitability. CP normal-state sprites
are the rendered roster matrix; other owners/states, fractional zoom/densities,
fog, buildings, full maps and contextual contrast remain open. The white anchor
cross aids review but does not establish original physical-contact correspondence.

No gameplay command or authoritative state is present. The main game still needs
Phase 2.3–2.8 runtime packs, composition, command/event binding and input flows.
The in-app browser remains unavailable, so browser/WASM playback and screenshots
were not exercised. Subjective likeness, original mappings, source provenance and
final visual/runtime acceptance remain open.
