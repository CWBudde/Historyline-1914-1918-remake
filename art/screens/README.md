# Screen art

Each family holds its `prompt.md` and imagegen source PNGs. Full-screen interface sources are
`frame.png`; control filenames use the variant name with slashes and spaces replaced by hyphens.
For example, the button states are `button/normal.png`, `hover.png`, `focus.png`, `pressed.png` and
`disabled.png`; an invalid focused input is `widget/input-invalid-focus.png`.

There are 108 frame/control source variants: 36 frames, five buttons, two country swatches, two progress
parts, two scrollbar parts and 61 widget states. These are native imagegen candidates, separate from
the existing runtime renders in `assets/screens/`. Their dimensions are recorded in `art/index.json`.
Visual acceptance and export bindings remain open.

Button sources: [normal](button/normal.png), [hover](button/hover.png), [focus](button/focus.png),
[pressed](button/pressed.png), [disabled](button/disabled.png).

## Scene paintings

Opaque, text-free candidate scenes in the existing slate, olive, earth and brass style. Each painting's
family `prompt.md` records its prompt, style references and native dimensions. Generated interface
frames and live text remain separate. These scene assignments await runtime feature bindings.

| Screen | Painting |
| --- | --- |
| Title | `backdrop/western-front.png` |
| Side selection | `side-selection/illustration.png` |
| Mission selection | `mission-selection/illustration.png` |
| Briefing | `briefing/illustration.png` |
| Campaign | `campaign/illustration.png` |
| Options | `options/illustration.png` |
| Pause | `pause/illustration.png` |
| Save / load | `save/illustration.png` |
| Handover | `handover/illustration.png` |
| Victory | `victory/illustration.png` |
| Defeat | `defeat/illustration.png` |
| Credits | `credits/illustration.png` |
| Help | `help/illustration.png` |
| Editor / new map | `editor/illustration.png` |
| Intermission | `intermission/illustration.png` |
| Production / repair | `production-repair/illustration.png` |

Related states reuse their main scene: loading, mission errors and invalid mission codes use mission
selection; save/load errors, progress, names and file confirmations use save; exit/restart confirmations
use pause; settings/dropdowns use options; long text uses its briefing/help/credits scene; invalid
new-map input uses editor. Runtime controls retain their vector exports alongside the imagegen sources.

Handover's painting is a closed canvas partition containing no map or player information. Its current
runtime frame remains opaque and never samples a previous gameplay image.

The illustrations are native 1586×992. For 2×, trim one pixel from each edge to 1584×990, then downsample
to 1280×800 to preserve the composition. The backdrop is 1672×941. Native 4× (2560×1600) sources and
visual acceptance remain open; upscaling does not close them.

## Panel surfaces

`surface/steel.png` and `surface/ground.png` are seamless 1254×1254 tiles under the battle screen's
drawn panels: the steel of the unit, contents and battle panels and the ground of the battle scene's
hexes. `surface/rivets.png` is a 4×2 sheet of brass rivet heads for the plates' rivets.
`hlasset build-screens` area-averages them to 512 px wide (1x) and 1024 (2x) as `surface/steel`,
`surface/ground` and `surface/rivets`; `internal/dospanel` uses only the tiles' texture, graded to its
own palette, and the heads in their own hue at its brass's lightness.

## Capture scenes

`capture/depot.png` and `capture/factory.png` are the staff-map room of the capture scene with a model
of each building, `capture/hand.png` the hand that plants the flag, and `capture/hq.png` and
`capture/hq-open.png` the headquarters of the finale, closed and open. `hlasset build-screens` scales
the room and the hand to 512 px wide (1x) and 1024 (2x), the headquarters to 640x400 and 1280x800.
`internal/dospanel` draws the pin, the flags and all motion; the points it needs on the masters are
listed in `capture/prompt.md`.
