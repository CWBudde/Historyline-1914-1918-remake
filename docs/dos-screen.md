# The split battle screen

The battle screen follows the original's layout. Two map panels sit side by side, one per
side, inside a riveted steel frame whose inner edge follows the outline of the visible hex cells, with a
status bar under each panel. Info, unit, contents, strategic-map and battle views open inside a panel within
the same frame. The look is newly authored (painted steel,
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
  cell, HQs as labelled plates in the side colours, depots as round D plates, factories as square F plates,
  units as dots in the side colours (greyish once they have their order, as their sprites turn grey), the
  panel's current view as a light rectangle.
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
- **Zoom** (`dosscreen.Zooms`): ½ to 2× the fitted scale in steps of about a tenth (seven per doubling) with the mouse wheel over the human
  panel, shared by both panels, keeping the hover cell in view; at least 4×4 cells stay visible. In the
  slim frame zooming out stops where the map would no longer fill the openings (`Screen.FillZoom`).
- **Frame** (`dosscreen.Frame`): drawn at runtime, not one painting: steel masked by the union of the
  panel's visible cells, bevel and lip from the mask edge, rivets at the tooth vertices, the divider post,
  bolt and status wells as fixed pieces. Cached per window size, zoom and on-map cell pattern; cells beyond
  the playable area stay steel, so the frame closes in at its end. The playable area is the map without its
  border row and column (`dosbind.Board.Playable`): cameras, cursor, frame, map drawing and the strategic
  map never show the border. This plate frame belongs to the tiles, the remake's and the original's; the
  painted backdrop uses the slim frame instead (`Frame.Slim`, `slim.go`): plain rectangular openings edge to
  edge, a thin gunmetal border with a brass hairline facing the map, gunmetal status strips, a brass knob at the
  divider's foot; every playable hex whose centre lies in the opening takes the cursor and the pointer. The
  opening always ends on the map; only a map smaller than the opening shows gunmetal past its ends
  (`Frame.DrawBeyond`). The web edition keeps this plate frame, its pixel-font wells and overlays for the
  original design only; its painted and tile designs sit in the menus' riveted plate (`ui/armour.css`,
  `split/SplitUI.tsx`): two rectangular openings in a margin of plate, a riveted bar between them, under each
  a steel strip with a signal lamp, the side's plaque, the status stencilled and brass plates, and what opens
  (situation, small map, unit, contents, menu) as a steel board over its window.
- **Cameras** (`cmd/historyline/split.go`): one per panel, snapping to whole column pairs so the teeth keep
  their phase; each follows its side's cursor and that side's revealed moves. The pointer on the human
  panel's outer half cells steps the camera (edge scroll, `cursor.go`). In the slim frame a camera at a
  map end shifts by whole device pixels so that its opening ends on the playable area
  (`dosscreen.Screen.SlimFit`).
- **Input** (`internal/gesture`, `cmd/historyline/cursor.go`): the human side's panel takes the mouse (a
  click selects and confirms, a drag of one cell pitch picks a gesture, right-click opens a ring of every
  available action) and key set A (arrows + space). Every cell of a building stands for its entrance: a click
  on any of them opens its contents or moves the aimed unit in.
- **Two humans** (`cmd/historyline/local.go`): the game values' PLAY row (or `-play` with a `2p` setup such
  as `-level 1918-11:2p`) starts the month's two-player setup with no computer side; the phase changes once
  both humans have confirmed, and neither confirm gives the other's orders. TWO HUMANS plays both panels at
  once: the left (Central Powers) takes the mouse, key set A and the global keys (F1, Esc, F2), the right
  (Entente) key set B, D/C/X/V as arrows and the left Ctrl as fire, and confirms with the mode change twice
  (Ctrl and X). Each human keeps its own cursor, selection, prompt, open view and message. HOT-SEAT plays one
  side at a time on the mouse and key set A with the other panel black; after each confirm a curtain covers
  both maps until a key or click takes over, the resolution plays for the human who takes over, and the
  other one sees it behind its own curtain on its next turn. A save made behind the curtain resumes there
  (`saves.Envelope.Mode`).
- **Status wells** (`cmd/historyline/well.go`, font `internal/pixfont`): the unit line (group size,
  experience badge, name, flag) or the side's state message in the panel's colour; the computer plans on a
  `Session.Clone` off the frame (`plan.go`) and its progress bar runs in its well; then its orders are
  carried out one by one, each move walking on its panel with the camera following, before it confirms;
  once the human side has confirmed, they play three times as fast.
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

## Web order menu

The web edition (`webapp/app/src/ui/orderMenu.ts`, `OrderMenu.tsx`) gives a unit's own orders as painted hex
plates (`art/screens/plates`) on its neighbouring cells, in both layouts and on a phone; the Go game keeps its
ring. Each order has one fixed slot; an order that does not apply leaves its cell free, one that applies but is
not possible now stays as a stripped, rusty plate that names its reason on the caption. The unit stays visible
under its cursor ring; the centre cancels.

| Slot | Key | On the map | In a container's contents |
| --- | --- | --- | --- |
| up (N) | W, 8 | Move; Attack (red plate) in the attack mode | Move out |
| up-left (NW) | Q, 7 | Contents of an own container | Produce |
| down (S) | S, 2 | | Repair |
| up-right, down-right, down-left | E, D, A (9, 3, 1) | free | free |
| centre | Escape | Cancel | Cancel |

Up, left and down keep the desktop gestures (up move, left open or produce, down repair). End phase, the
strategic map and information are not plates: they are brass nameplates on the order strip with their keys,
F1 END PHASE, O MAP and I INFO, on the steel strip under the single map and at the outer end of each human
panel's strip on the split screen, End phase set a little apart and a Menu plate beside them for the settings,
surrender and leaving. No flick can reach End phase; the keyboard gestures (hold Space or left Control and push a
direction) still give the desktop's mode change, map and info.

Upright, the map runs from edge to edge: a 48 px strip on top (lamp, plaque, the status over the selected or
hovered unit or ground, the order count, Menu) and under the map only the row of the three plates. Info there
opens and closes a board rising from that row with the unit, its orders, the ground and the queued attacks, and a
Situation plate; a building's contents open as a board over the map.

- **Opening:** right-click a cell with orders, click the selected unit again, tap an own unit (it is selected
  and its plates appear), long-press, or Enter on the selected unit (the first plate takes the focus).
- **Running:** click or tap a plate; or press on the unit and flick towards a slot (24 CSS px, the plates
  appear only after 150 ms so a fast flick does not flash them); or the slot's key while the menu is open.
  Hit areas are 60° wedges from half a neighbour's distance to 1.75 of it, not the plates; outside them a tap
  closes the menu and, with a unit selected, acts on that cell (a tap on a destination moves there).
- **Sizes:** plates are 0.7 of a cell with a mouse (72-160 CSS px wide), 0.92 on touch (at least 58 px wide,
  48 px across); where they outgrow the cell the rosette grows past the neighbours. At a window edge the camera
  pans first; what is left moves the whole rosette inwards, tied to the unit by a brass bracket, never
  mirrored.
- **States:** idle, lit (rim rubbed bright; the caption hangs from the plate's outer edge, on touch above the
  rosette), pressed (sunk), not possible (stripped). The plates are `menuitem` buttons for the keyboard and
  screen readers: arrows walk the focus, Enter runs, a chalk outline shows the focus.
