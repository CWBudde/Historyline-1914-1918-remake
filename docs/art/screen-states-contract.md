# Controls and common screen-state candidates

The `screens-v2` continuation completes offline candidates for the dedicated
controls and common screen states left open by the first full-screen batch.
Open the [new gallery](../../art/exports/screens-v2/index.html); the sixteen
[original screen studies](../../art/exports/screens-v1/index.html) remain intact.
These are artwork/layout fixtures, with no save, loading, campaign or editor operation.

The new 400-file bundle retains 61 widget IDs, twenty screen-state frames, four
progress/scroll components, the five existing button states and two country
symbols. There are **92 editable SVGs, 276 vector PNG exports and two reused
painting PNGs**, plus twenty labelled 2× review compositions. Every vector has
1×/2×/4× exports; labels and values remain live. Saved review compositions are
excluded from runtime-art counts. The unchanged painting and pinned fonts are
snapshotted from the first batch; no new image generation was needed.

## Dedicated controls

| Family | Offline candidates |
| --- | --- |
| Input | Normal and invalid variants, each with normal/hover/focus/pressed/disabled states. The invalid field reserves an error-symbol region and underline. |
| Toggle | Off/on/mixed values, each in all five interaction states. Check, empty box and bar distinguish values without relying on colour. |
| Slider | Separate track, fill and thumb, each in five states; explicit anchor/travel and fill-clipping geometry. |
| Dropdown | Open/closed fields and normal/selected options, each in five states, plus a separate four-row list surface. |

`widget-geometry.json` records content insets and composition rules. A 220×32
slider moves its 24×32 thumb centre from (12,16) to (208,16). At value `v`,
clamp to [0,1], preserve the full fill image and clip at x=`12+196*v`; draw no
fill at zero. Draw the thumb above the fill. The 50% fixture clips at x=110
and anchors the thumb there. Shrinking the fill image would move its inset and
is explicitly tested against. `screens.json` retains local clip rectangles.

The dropdown list is 240×130, with four 238×32 options at (1,1), (1,33),
(1,65), (1,97). Labels use the reserved content regions, and disabled fixture
labels are muted. Native HTML selects suppress their own arrow so it cannot
overlap the painted chevron. The browser's actual popup is still platform UI;
the painted list is a separate candidate for an eventual in-game dropdown.

## Twenty screen-state studies

- Loading at 0%, 50% and 100%; mission read error and retry.
- Save error, unreadable save, empty saves and save progress.
- Exit, restart, overwrite and delete confirmations, with Cancel visibly focused.
- Settings composed with actual field/toggle/slider/dropdown art.
- Invalid mission code, an open dropdown, and invalid new-map dimensions.
- Long briefing, help and credits; a long Unicode save name.

The full strings remain in `screens.json`. Native review text wraps by measured
Unicode runes, handles long tokens after short words, clips to its declared
region and draws complete visible lines. The save-name field clips horizontally.
The HTML version retains full text in labelled, keyboard-focusable scroll regions.
These are test texts and example values; they add no campaign narrative or rules.

The gallery also includes real HTML text fields, a checkbox, range and select
for visual review. They alter only local control values. All 61 widget images
are shown at their stated logical size with SVG and density links. This does not
establish an exercised browser or gameplay input flow.

## Reproduction and evidence

```sh
go run ./tools/interface-art -manifest art/screens-v2.json -out build/screen-review/fresh-states
go test ./tools/interface-art -count=1
go run ./tools/phase2-audit
```

The same strict, contained input loading, source pins and fresh-output staging
apply to both screen profiles. The old `interface-v1` and `screens-v1` bundles
remain byte-identical. Native 4× vectors reduce 4× → 2× → 1× through the existing
signed premultiplied byte-RGB Lanczos3 filter. The reused painting remains 1×/2×,
without upscaling. Git attributes preserve all byte-pinned bundle outputs.

Tests verify every saved pin/dimension and density, zero RGB at zero alpha,
all gallery HTML/CSS links, the complete 400-file repeat, clipped fill/foreground
order, Unicode wrapping and strict text-region containment. The v1 regression
continues comparing all 119 original screen-bundle files and 469 interface files.
All twenty saved studies and the 61 logical widget images were visually inspected.
Independent code review caught and corrected fill alignment, long-token wrapping,
logical-size gallery display and duplicate browser-chevron issues.

The in-app browser again returned no available browsers. Native HTML control
interaction, browser screenshots, keyboard navigation, responsive layouts,
tooltips and accessibility acceptance remain open. The gallery's fixed 640×400
studies do not implement responsive game layouts. No runtime replacement is accepted.

## Remaining ownership

Phase 2.3–2.8 owns runtime packs, actual selection/visibility, layouts and gameplay
flows. G4 and the owning content phases supply future production/repair/campaign
and editor-specific states when their behavior exists. Native/campaign illustrations
need verified content IDs in Phase 5. Provenance, original correspondence and
approximately 90% visual resemblance remain review gates. A true 4× painted
backdrop remains optional-pack work; the required 2× candidate is already retained.
