# Full-screen artwork candidates

This Phase 2.2 continuation delivers sixteen separate screen-frame candidates in
the existing slate/brass interface vocabulary, one painted backdrop and live
layout fixtures. Open the [review gallery](../../art/exports/screens-v1/index.html).
These are offline artwork candidates; no gameplay screen is enabled by this batch.

## Coverage

| Views | Inventory binding | Candidate composition |
| --- | --- | --- |
| Title | `screen:title` | Painted countryside, left menu, disabled Continue example. |
| Side / mission selection | `screen:side-mission-selection` | Separate side cards with the existing country flags; mission list and details with selected/available/completed/locked and invalid-code examples. |
| Briefing / campaign | `screen:briefing-campaign` | Separate narrative/objective and progression frames. Current levels have no briefing text or unlock metadata; fixtures invent no historical narrative. |
| Settings | `screen:options` | Grouped display, sound, animation and rules rows. Setting values are examples, not implemented options. |
| Pause | `screen:pause` | Central modal frame and resume/navigation controls. |
| Save / load | `screen:save-load` | Separate views, slots and metadata, current/empty/incompatible/unreadable examples. No save operation is implemented. |
| Handover | `screen:hot-seat-handover` | Fully opaque neutral screen; next-player label and Ready control; no previous scene or unit data. |
| Victory / defeat | `screen:victory-defeat` | Separate outcome fixtures with live reason, map/round/profile and continuation/retry controls. |
| Credits / help | `screen:credits-help` | Separate attribution and reference-control text regions. Release attribution review remains open. |
| Editor / new map | `screen:map-editor`, `screen:map-editor-new-map` | Palette/workspace frame and separate field/validation frame. No map compositor, editor or creation wizard is implemented. |

The bundle contains 24 asset IDs: sixteen new frames, one new backdrop and seven
reused controls/country symbols. It retains 71 text-free PNG exports, 23 editable
SVGs and sixteen 1280×800 font-rendered review compositions. The five button
states are normal, hover, keyboard focus, pressed and disabled. Country graphics
remain independent fixture layers, not baked into the side-selection frame.
`screens.json` records every region, example label, control and decoration.
`vectors.json` retains the common editable path model.

All frame and control artwork is text-free. Labels, player identity, mission and
round, outcomes, slot metadata, settings and errors remain live. The labelled
`fixtures/` PNGs are review artifacts, excluded from runtime-art counts. Example
data in the Go catalog is explicitly presentation data, not a new gameplay table.

## Style and reference limits

The new painting uses continuous illustrated countryside, a trench and distant
village with quiet menu space. It was generated with the built-in image tool.
The exact prompt is retained in
[`art/recipes/screens/v1/western-front.json`](../../art/recipes/screens/v1/western-front.json);
the unchanged 1672×941 master is in `art/source/screens/v1/western-front.png`.
The pinned manifest records the source hash. Generation itself need not repeat;
all subsequent exports work offline.

The manual's main-menu and side-selection grayscale images were inspected as
layout evidence: a left menu and two country choices remain familiar. Their
320×200 logical aspect informs the 640×400 layout studies. They are lossy,
platform-unverified references, not reliable colour or exact-pixel targets.
No original/manual pixels were submitted to generation or enter the exporter.
Other proposed views lack an established reference in this batch. Approximately
90% resemblance after Lanczos reduction remains an unapproved visual target;
this delivery does not claim a measured score or original-screen fidelity.

## Density and reproduction

```sh
go run ./tools/interface-art -manifest art/screens-v1.json -out build/screen-review/fresh
go test ./tools/interface-art -count=1
go vet ./tools/interface-art
golangci-lint run ./tools/interface-art/...
```

Frames are 640×400 logical pixels, rasterized natively at 2560×1600 and reduced
4× → 2× → 1× with the existing signed, premultiplied byte-RGB Lanczos3 filter.
The painting has an explicit 1504×940 crop at (84,0), preserving exactly 8:5.
It reduces to 1280×800 and then 640×400 without stretching or upscaling.
**A true 4× painted backdrop is unavailable** at this source resolution. The
optional 4× screen pack remains incomplete; vector-frame coverage does not close it.

The schema-1 `screens-v1` profile reuses the interface exporter's contained,
pinned source reads, font snapshots, fresh-output staging and report hashing.
Unknown fields, trailing JSON, missing/duplicate fonts, changed source bytes,
reference/escaping paths, invalid crops and upscaling fail. Existing outputs
are never replaced. `report.json` pins all other 118 bundle files.
Git attributes preserve bundle bytes and the manifest/font-licence LF endings
across Windows and Unix checkouts, so line-ending conversion cannot invalidate pins.

Native 2× review compositions render fonts directly at that density. Their vector
layers are rendered directly at 2×; export PNGs instead use the documented 4×
Lanczos chain. These compositions establish font placement, not byte-identical
assembly from exported frame PNGs. The HTML gallery uses retained PNG layers and
live local-font text. Fixed-size artboards scroll on small review windows; this
does not establish responsive game layouts or supported UI-scale coverage.

## Evidence and remaining work

The generated master and all sixteen saved compositions were inspected together.
Tests verify retained pins/dimensions, all separate views, every local HTML/CSS
asset link, full bundle byte-for-byte reproduction and no replacement of an
existing output. An independent red/blue backdrop perturbation proves the
handover composition is opaque and does not sample previous imagery; its HTML
also contains no backdrop image. Strict input and crop rejection are exercised.

The in-app browser was unavailable (`browsers.list()` returned no browsers).
Hover/focus rendering, keyboard traversal, mouse actions, scrolling and browser
screenshots were not exercised. Tests and native PNG inspection do not close
those checks. Full `just ci` is not claimed for this artwork continuation.

Keep open: visual/reference approval, dedicated input/toggle/slider/dropdown art,
loading/error/progress and confirmation-modal studies, long-text/overflow states,
all faction/season variants, campaign/intermission illustrations, source/license
review, the optional true 4× painting, responsive/high-DPI layouts, visibility
plumbing and actual screen/input flows. Screen behavior belongs to its owning
Phase 2–6 milestone. Effects/movement/event playback and assembled-map contrast
remain separate open Phase 2.2 acceptance work.
