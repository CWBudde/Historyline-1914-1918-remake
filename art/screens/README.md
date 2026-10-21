# Screen paintings

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
new-map input uses editor. Buttons, widgets, country marks, progress bars and scrollbars remain vectors.

Handover's painting is a closed canvas partition containing no map or player information. Its current
runtime frame remains opaque and never samples a previous gameplay image.

The illustrations are native 1586×992. For 2×, trim one pixel from each edge to 1584×990, then downsample
to 1280×800 to preserve the composition. The backdrop is 1672×941. Native 4× (2560×1600) sources and
visual acceptance remain open; upscaling does not close them.
