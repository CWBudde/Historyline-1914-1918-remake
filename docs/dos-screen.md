# The split battle screen (`-rules dos`)

Under `-rules dos` the battle screen follows the original's layout. Two map panels sit side by side, one per
side, inside a riveted steel frame whose inner edge follows the outline of the visible hex cells, with a
status bar under each panel. Info, unit, contents, strategic-map and battle views open inside a panel within
the same frame. `Classic` and `Standard` keep the single-map HUD. The look is newly authored (painted steel,
the remake's terrain and units); only layout, proportions, information and behaviour follow the original.

## Reference

Local frames of the original at 320×200 lie under `original/reference/screenshots/` (git-ignored, never
committed): `dosbox/` (menus, every in-game view below, the phase change, both sides' turns on PULSE) and
`strategywiki/` (screens, action icons and their diagrams). `dosbox/game_pulse_start.png` is the PULSE
opening and matches StrategyWiki's Screen 3. Measurements are in those native pixels; the remake scales them
with its art (2×, optional 4×).

- **Screen** (312×200 used): left panel x 0–151, divider x 152–159, right panel x 160–311; map
  viewports y 0–179, status bars y 184–197.
- **Panels:** left = player A (the Central Powers on the shipped levels), right = player B. Each panel
  shows its side's own map (its units and what that side knows), its own cursor and its own status bar.
  Against the computer, the computer's panel stays visible and shows its view and its moves as they happen.
- **Viewport:** about 9 cell columns × 7½ cell rows (24×24-px cells, columns 16 px apart). The frame's
  inner edge is the outline of the outermost cells: teeth every 32 px along the top and bottom (12 px
  deep) and every 24 px along the sides (8 px deep). The view scrolls in whole cells, so the teeth always
  meet cell edges.
- **Frame:** olive-grey steel plates (neutral greys with an olive cast, about `#24241C`–`#B2B2A6`), a light
  bevel on edges facing the top-left light, a dark lip along the inner outline, round rivets along every
  tooth, the divider as a vertical post with a highlight line, a bolt at the foot of the divider between
  the status bars, scuffed highlights near the lower corners.
- **Status bar** (152×14 per panel): a recessed well with a near-black olive fill (`#181C00`) and a thin
  brown bevel (`#453828` / `#24180C`), one line of centred capitals in a pixel font. Text is pale green
  (`#BED392`) on the left and cream (`#FBE3C7`) on the right; numbers are amber.
  - Cursor on an own unit: group size at the left, unit name centred, the owner's flag at the right.
  - Otherwise a state message: waiting for the other side's mode; the computer carries out its orders;
    the unit has moved; confirm the mode change with F1 / leave with the fire button; the new mode; the
    computer analyses the position, with a thin progress bar along the bottom of the well.
- **Cursor:** a hex ring of the same steel around the cell. While fire is held, the ring becomes a solid
  hex plate that shows one action icon (cancel first). Pushing a direction while holding picks an action;
  releasing runs it. After an order the plain ring returns; the moved unit is drawn greyed (no hatched
  plate).
- **Action icons:** hex steel plates with an embossed glyph: cancel (X), move (four arrows), attack (fist),
  mode change (double arrow), open (box), strategic map (D on a scroll), info (?). Around the centre:
  move and attack above, mode change and open to the left, map to the right, info below. In a container:
  move (out) above, produce (hammer) to the left, cancel to the right, repair (wrench) below.
- **Move range:** every cell out of reach (terrain and units) is darkened to about 0.62 brightness;
  reachable cells keep full colour.
- **Unit info** (info on a unit): the map gives way to a steel inner plate with a sepia photo (top left,
  about 100×84), the unit's sprite on a small hex plate beside it, a brown name plate with brass rivets
  and a stats plate: weapon strength and range against air, ground and water; armour, speed, weight,
  maximum group size.
- **Game info** (info on the HQ): the map darkens and the text stands on it: round, level, mode, best
  and current score; a table with columns for side one, side two and the whole map (units, depots,
  factory units, energy total and income); move count and move limit.
- **Strategic map:** a dark steel plate with a recessed frame holding the whole map: terrain colour per
  cell, HQs as labelled plates in the side colours, depots as D plates, units as dots in the side colours,
  the panel's current view as a light rectangle.
- **Contents** (open on a container): a sepia photo of the unit under the cursor, a plate with the
  container's name and energy (stored, total), the unit's name plate, a large plate for details, and two
  rows of hex slots holding the units inside; the cursor moves over the slots.
- **Battle:** a sepia animation fills one panel inside the frame (air: close-ups of a pilot and a
  biplane) while the other panel shows the map. F9 turns the animations off.
- **Phase change:** both panels fade to greyscale with the confirmation prompts, go black with only the
  cursor plates left, show the new mode of each side on the grey maps, then colour returns while the
  computer's panel shows its analysis progress.

## Remake

The package docs hold the details; this is the map of where each part lives.

- **Layout** (`internal/dosscreen` `Layout`): two equal panels, the divider and the two status bars fill the
  window at any aspect ratio. The panel height sets the cell scale so that `TargetRows` (8) cells per
  column, 8½ rows with the half-cell shift, fit; extra width shows more columns (no letterboxing). The
  smallest window keeps `MinCols` × `MinRows` (7×6) cells per panel. The fixed pieces (divider, wells, bolt)
  are reference pixels at 2 logical pixels each, times the UI scale.
- **Zoom** (`dosscreen.Zooms`): ½, ¾, 1, 1¼, 1½ and 2× the fitted scale with the mouse wheel over the human
  panel, shared by both panels, keeping the hover cell in view; at least 4×4 cells stay visible.
- **Frame** (`dosscreen.Frame`): drawn at runtime, not one painting: steel masked by the union of the
  panel's visible cells, bevel and lip from the mask edge, rivets at the tooth vertices, the divider post,
  bolt and status wells as fixed pieces. Cached per window size, zoom and on-map cell pattern; cells beyond
  the map edge stay steel, so the frame closes in at the edge.
- **Cameras** (`cmd/historyline/split.go`): one per panel, snapping to whole column pairs so the teeth keep
  their phase; each follows its side's cursor and that side's revealed moves. The pointer on the human
  panel's outer half cells steps the camera (edge scroll, `cursor.go`).
- **Input** (`internal/gesture`, `cmd/historyline/cursor.go`): the human side's panel takes the mouse (a
  click selects and confirms, a drag of one cell pitch picks a gesture, right-click opens a ring of every
  available action) and key set A (arrows + space).
- **Status wells** (`cmd/historyline/well.go`, font `internal/pixfont`): the unit line (group size,
  experience badge, name, flag) or the side's state message in the panel's colour; the computer plans on a
  `Session.Clone` off the frame (`plan.go`) and its progress bar runs in its well.
- **Unit markers** (`unitmarks.go`): the experience badge above a unit, the group size at the lower right
  of its hex, from the `rank` art of the screen catalogue.
- **Overlays** (`internal/dospanel`, `cmd/historyline/overlay.go`, `contents.go`): unit info with the
  authored portrait through a sepia tint; game info on the darkened map (BEST from `internal/scores`);
  contents with a slot cursor for move out, produce and repair. RESOURCES is the building's credits, TOTAL
  the side's account.
- **Battle** (`dospanel.DrawBattle`, `phase.go`): both groups built up figure by figure in the other panel,
  each loss a muzzle flash, a blast and a wreck while the count steps down; the map follows loss by loss
  and a destroyed unit leaves at its explosion's peak (`effect/explosion-00..08`, `banner.go`).
- **Capture** (`taking.go`, `dospanel.DrawCapture`, `DrawFinale`): a captured depot or factory shows its
  site while the other panel plays a hand planting the capturer's flag on a model of the building; taking
  the HQ ends on its doors opening on the surrendering defenders and the winner's flag going up.
- **Phase change** (`phase.go`): the F1 prompt in grey, black with the cursor plates, the battles, the new
  modes in grey, colour, then the computer's analysis.

F9 turns battle and capture scenes off, reduced motion shows their end, and a key or click skips one.
