# Interface, markers and effects candidate contract

This Phase 2.2 delivery adds offline candidates for ownership, information,
interaction markers, menu/panel frames, fifteen model-specific weapons, and
hit/death/capture/round presentation. It follows the existing slate/brass
unit-panel vocabulary and continuous painted weapon/effect treatment. Nothing in
this batch is a runtime-accepted replacement or a finished screen/input flow.

Open [reference comparisons](interface-review.html) or the
[complete gallery](../../art/exports/interface-v1/index.html). Every stable ID is
recorded in [runtime-interface-coverage.csv](runtime-interface-coverage.csv).
Exact generation prompts and input hashes are in `art/recipes/interface/v1/`;
untouched generated masters and the vector model are in `art/source/interface/v1/`.
The built-in image generator produced fifteen separate weapon masters and one
nine-frame explosion atlas. All export tooling is Go.

## Families and semantics

| Family | Delivered offline candidates |
| --- | --- |
| Ownership | Neutral/CP/EC 35×6 strips, explicitly authored angled factory variants, and a 35×18 `depotOwnership` atlas in legacy row order neutral/CP/EC. |
| Information | Germany/France country flags; experience 0–10; Capture and neutral unknown-skill symbols; sight/size/armour/weight, fuel/ammunition, unlimited/empty, target-class and repair/turn symbols; bar track, segments and unavailable marker. |
| Map interaction | 120×100 cursor, selected/origin/attack/reachable/attackable/invalid/path/editor markers and camera rectangle. Open centres preserve underlying art. |
| Controls | Move/attack/contain/exit, slot selection, weapon rows and generic buttons, each in normal/hover/focus/pressed/disabled states. |
| Panels | Text-free 654×110 terrain, 840×156 battle, 670×286 weapon and 527×670 container frames; 480×112 capture/round banners. The container reserves exactly 2×9 slots. |
| Weapons | All fifteen keys from `internal/data/weapons.json`, each with separate 252×50 logical art. The four-row fixture distinguishes finite, unlimited, zero and empty slots. |
| Effects | 72-square vector hit indicator and reduced ring; nine independently painted 96-square explosion frames; finite static capture/round banners. |

All assets have retained 1×/2×/4× PNGs. Vectors also have editable SVG masters and a
serialized common path model. Art contains no baked labels, ammunition, rank
numbers, capacity, loss counts, owner/player identity or round numbers.

Owner palette is independent of country. Germany/France flags identify type data,
not current controlling side. Ranks are experience indices, with live numeric
values and no invented military titles. Only skill 1 receives the code-backed
Capture legend; skills 5/7 and any future unknown ID use a neutral symbol plus a
live numeric ID. A repair symbol does not establish repair rules or eligibility.

Weapon IDs preserve `revolveur`, `lugerp08`, `wickers` and `morser`. The existing
`canon75` display-name error and other legacy weapon-label/model discrepancies
are left unchanged in game data. Art binds to the ID and its reference picture.
Names and stats in the gallery come from the embedded compatibility tables.
They are examples; runtime ammunition must read current state. A missing weapon,
nil/unlimited ammo and finite zero are separate states. Raw target flags do not
establish target immunity: Standard Ground weapons can attack vessels.

## Geometry, density and animation

The map marker anchor is (60,50), aligned to terrain centre (48,40) with offset
(-12,-10) from the 96×80 tile origin. This is separate from the existing 142×40
map pitch and unit anchor contract. Picking, odd-row placement, fractional zoom,
fog, map edges and minimap composition still require the runtime workbench.
Ownership placement on the newly painted buildings must be reviewed separately;
the legacy (31,14)/(61,9) blits cannot be assumed to fit those scenes.

The Go catalog renders vectors natively at 4×. Exports reduce 4× → 2× → 1× using
signed premultiplied byte-RGB Lanczos3, with zero RGB at zero alpha. Weapon crops
are explicitly recorded around alpha > 8 bounds plus eight source pixels. This
discards very faint generated alpha outside the crop; inside it, alpha is
preserved. Masters are retained unchanged. Proportional fitting and transparent
letterboxing preserve grips, stocks, muzzle/blade orientation and the reference
wheel crops, without stretching or upscaling. Small pistols occupy less width
than rifles. No original pixels enter builds or compositing.

The generated explosion master is 1254×1254: nine exact 418×418 cells, row-major,
reduced whole to 384/192/96 with stable logical anchor (48,48). Frames are not
independently trimmed or recentered. Death lasts 900 ms (nine 100 ms frames);
reduced playback is one static smoke frame for 200 ms. Hit is a static 1000 ms
indicator, reduced to a 250 ms ring. Loss count remains live text, so damage is
never used as an unchecked ten-frame atlas index. Capture/round last 900 ms,
reduced to static 250 ms. All tracks are finite, non-looping and skippable. The
gallery starts only on request; changing reduced mode or hiding it clears playback.

Movement/facing reuse the existing six authored views for all fifteen unit
families. `bindings.json` records anchor-preserving path translation and reduced
motion snapping as the integration contract. It supplies no movement renderer;
runtime interpolation, facing transitions and event sequencing remain open.

## Reproduction and verification

```sh
go run ./tools/interface-art -manifest art/interface-v1.json -out build/interface-review/fresh
go test ./tools/interface-art -count=1
go vet ./tools/interface-art
golangci-lint run ./tools/interface-art/...
```

The exporter pins and snapshots sources/fonts, rejects legacy/escaping paths and
bad hashes, validates the complete fifteen-ID binding set, and publishes only to
a fresh directory. The report pins every generated file except itself. Gallery
labels are live HTML using local pinned Go fonts; PNGs are text-free asset exports.

Saved logical-size art and every generated master were inspected. Focused tests
cover fitting without clipping/stretching, no upscaling, transparent marker
centres and contained factory geometry, finite/resolvable reduced tracks, source
containment/pins, strict inputs and signed Lanczos impulse/alpha behavior.
Final file counts, repeat evidence and completed checks are recorded in PLAN.md.

The in-app browser returned no available browser in this session. Browser
playback, keyboard interaction and screenshot composition were **not exercised**;
static path/image verification does not close those checks. Full `just ci` is not
claimed, and the previously recorded Windows exporter issues remain separate.

## Open acceptance and later screens

Approximately 90% resemblance after logical-size Lanczos reduction is the user's
visual target, not a measured similarity score. This batch retains the familiar
weapon orientation and broad silhouettes; the approved panel vocabulary governs
new vector frames and controls. Final visual/reference approval remains open.

Native ownership/marker mappings, building-specific placement, assembled map
contrast/fog/zoom tests, runtime manifests/loading, selection/input and visibility
plumbing, motion/event playback, responsive layouts and accessible keyboard/mouse
flows remain in 2.3–2.6 and G3/G4/G5. HTML fixtures establish no gameplay command.

Full-screen title, side/mission selection, briefing/campaign, options, pause,
save/load, privacy handover, victory/defeat, credits/help and editor artwork still
belong to G4/Phase 6. This delivery does not tick those screens complete.
