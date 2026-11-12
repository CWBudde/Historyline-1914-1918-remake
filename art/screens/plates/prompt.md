---
id: screens/plates
title: Order menu plates
kind: screen
status: candidate
runtime_ids:
masters: plate.png, plate-lit.png, plate-stripped.png, plate-attack.png, plate-attack-lit.png, icons.png
---

## Look
The web edition's order menu round a unit (the "A plates" design chosen on 2026-11-12): small worn steel hex
plates on the unit's neighbouring cells, flat-topped like the map's cells, a tarnished cast brass rim and two
slotted brass screws, field-grey paint chipped to bare steel and rust; a bone-white army stencil symbol per order
laid on the plate in the game. Lit (under the pointer or the focus) the rim is rubbed bright; an order not
possible now is the same plate stripped to rust with a screw missing; the attack order is painted red-oxide
primer. The materials of the start screen's armour plate (`screens/armour`): used, chipped, never glossy, no glow.
Every part is painted alone and blank; the game composes plate and symbol, and all text is live.

## Base prompt
Every part was generated with this style text, the part's text below, and the attachments named in its line
(rival A's parts sheet `build/review/ring-rivals/A/draft-3.png` and the start screen draft
`build/review/gen-2026-11-12/start/start-01.png`, neither committed; the repainted states also attach the idle
plate).

```text
Newly authored interface art for the order menu on the battle map of a turn-based World War I hex strategy game, 1914-1918.
Hand-painted digital illustration with believable, physical materials, not a photograph, not a glossy 3D render, not a cartoon, not pixel art, not flat vector UI. Look and feel: 1914-1918 military hardware. Heavy metal, army issue, worn by field use; serious and handsome, never cheap, never plastic, no modern sci-fi glow, no neon, no gradients that look like web buttons.
MOOD: worn out, sad and exhausted. No heroism, no glory, no clean parade-ground metal. Everything has been used too long in rain and mud: dented, scratched and re-painted steel, rust bleeding from screws, dull tarnished brass with only the touched spots still bright, chipped paint, dried mud, soot. Desaturated cold palette of field-grey, wet earth, rust and bone white. Soft overcast light from the upper left; quiet and heavy rather than dramatic. Do not copy any existing game's graphics. No watermark, no logo, no small print.
```

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| plate | | `plates-01` (exec-ec8a8bd2, 2026-11-12), 1536x1024 RGBA, plate 1175x960, mean RGB 80,73,58: the blank idle plate. |
| plate-lit | | `plates-03` (exec-4d179ed5), plate 1187x962, mean RGB 107,97,76: rim and screws rubbed bright, face a step lighter; `plate.png` attached as image 1. |
| plate-stripped | | `plates-04` (exec-01f290f8), plate 1179x958, mean RGB 82,61,46: bare rusty steel, the left screw missing; `plate.png` attached. |
| plate-attack | | `plates-05` (exec-4cfa0560), plate 1205x954, mean RGB 109,62,44: red-oxide face; `plate.png` attached. |
| plate-attack-lit | | `plates-06` (exec-8e508ebd), plate 1206x957, mean RGB 131,77,54: the red plate lit; `plate-attack.png` and `plate-lit.png` attached. |
| icons | | `plates-02` (exec-9d581b0f), 1774x887 RGBA, 4x2 cells: move, attack, contents, move out / produce, repair, cancel, empty; symbols 236-344 px. |

## Post
`hl assets build-screens` puts them into the screens catalog's plates family (`internal/hlasset/plates.go`): each
plate cleared of its alpha haze, trimmed and scaled to 160x134 logical so the states lie exactly on each other,
each symbol cut from its sheet cell, trimmed and centred in a 96x96 logical square, at 1x and 2x. The web
edition takes them as lite art only. Pressed is the plate darkened and sunk in CSS; the caption and the order
strip use the start screen's brass nameplates.

plate:

```text
PART: ONE BLANK HEXAGONAL STEEL ORDER PLATE, IDLE. Single interface piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Subject: one small steel plate in the shape of a FLAT-TOPPED HEXAGON seen straight on, centred on the canvas: flat horizontal top and bottom edges, a point at the left and a point at the right. Exactly 1152 px wide (point to point) and 960 px tall (top edge to bottom edge), width:height 1.2:1; the top and the bottom edge are each about 552 px long (48 % of the width), the four slanted edges run to the left and right points at mid height. Margins: 192 px left and right, 32 px above and below.
Build: a flat face of worn field-grey painted steel, mean about RGB 70,72,62, chipped down to dark bare steel (about RGB 44,44,38) along the edges and in scattered flakes, small rust blooms and short rust runs (about RGB 112,64,34); a narrow raised rim of dull tarnished cast brass about 34 px wide all around the outline (mean about RGB 112,88,50, darker tarnish in the corners about RGB 66,54,32, only faint dull highlights on the upper edges), slightly bevelled; two slotted brass screw heads about 90 px across, one near the left point and one near the right point, each about 150 px in from its point at mid height, the slots at different angles, a little rust bleeding from each screw.
BLANK FACE: no symbol, no arrow, no icon, no lettering, no numbers, no marks in the middle; the game lays its stencilled symbol on top later. Keep the middle of the face calm and even.
Alpha zero everywhere outside the plate: no ground, no grass, no map, no halo, no glow, no cast shadow, no checkerboard. Nothing else on the canvas.

Attachments: Image 1 is the concept sheet of the order menu's parts: take from its FIRST plate (top left, the plain arrow plate) the shape, the worn field-grey steel face, the tarnished brass rim and the two slotted screws; copy none of its icons. Image 2 is the game's start screen: take from it the same worn army materials, rust and palette. Paint ONLY the one part described above, alone on its canvas.
```

plate-lit:

```text
PART: THE SAME BLANK HEXAGONAL STEEL ORDER PLATE, LIT (the plate under the pointer or with the keyboard focus). Single interface piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Image 1 is the idle version of this exact plate. Repaint it with exactly the same flat-topped hexagonal outline, size (about 1175 x 960 px), position on the 1536x1024 canvas, brass rim width, screw positions and screw slot angles, and the same chips, flakes and rust; change ONLY the state of the metal: the plate a hand touches. The brass rim is rubbed bright and warm along its whole length: mean about RGB 170,134,74, worn warm highlights on the upper edges about RGB 214,178,108, tarnish only left in the corners and in scratches; the two screw heads rubbed bright too. The field-grey paint of the face one step lighter and cleaner: mean about RGB 92,94,82 instead of RGB 70,72,62. Still old, scratched and chipped; not shiny new gold, not a mirror, no white glare, no glow, no light rays.
BLANK FACE: no symbol, no icon, no lettering, no numbers.
Alpha zero everywhere outside the plate: no ground, no halo, no glow, no cast shadow, no checkerboard. Nothing else on the canvas.

Attachments: Image 1 is the idle plate to repaint (keep everything but the state of the metal). Image 2 is the game's start screen: its selected START nameplate shows the rubbed bright brass wanted for the rim.
```

plate-stripped:

```text
PART: THE SAME HEXAGONAL ORDER PLATE, STRIPPED AND RUSTED (an order that is not possible now). Single interface piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Image 1 is the idle version of this exact plate. Repaint it with exactly the same flat-topped hexagonal outline, size (about 1175 x 960 px), position on the 1536x1024 canvas, brass rim width, screw positions and screw slot angles; but this plate has been left out in the rain for a winter: ALL the field-grey paint is gone from the face, which is bare, pitted, scratched steel covered in brown rust (mean about RGB 92,64,44, darker rust pits about RGB 58,38,26, a few duller grey bare-steel patches about RGB 84,82,74); the brass rim is dark, nearly black-green tarnish (about RGB 62,54,36) with no bright spots; the LEFT screw is MISSING, leaving an empty dark round screw hole with a rust ring where it was; the right screw is rusty and dull. Flat, dead and lifeless compared with image 1, but still clearly the same plate.
BLANK FACE: no symbol, no icon, no lettering, no numbers.
Alpha zero everywhere outside the plate: no ground, no halo, no glow, no cast shadow, no checkerboard. Nothing else on the canvas.

Attachments: Image 1 is the idle plate to repaint (same outline, rim and screw places). Image 2 is the concept sheet of the parts: its fourth plate in the top row is the stripped, rusty disabled plate wanted (copy none of its arrow).
```

plate-attack:

```text
PART: THE SAME HEXAGONAL ORDER PLATE, PAINTED RUST-RED (the attack order). Single interface piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Image 1 is the idle version of this exact plate. Repaint it with exactly the same flat-topped hexagonal outline, size (about 1175 x 960 px), position on the 1536x1024 canvas, brass rim width, screw positions and screw slot angles, and the same brass rim and screws in the same dull tarnished state; change ONLY the paint of the face: instead of field-grey it is worn dull red-oxide primer paint, mean about RGB 104,46,34, lighter worn patches about RGB 128,66,48, chipped down to dark steel and brown rust in flakes along the edges and here and there, as the attack plate of image 2 (second row, first plate). Dull and dirty, not bright red, not glossy.
BLANK FACE: no symbol, no icon, no lettering, no numbers.
Alpha zero everywhere outside the plate: no ground, no halo, no glow, no cast shadow, no checkerboard. Nothing else on the canvas.

Attachments: Image 1 is the idle plate to repaint (keep everything but the colour of the face paint). Image 2 is the concept sheet of the parts: its red plate with crossed rifles shows the red-oxide paint wanted (copy none of its rifles).
```

plate-attack-lit:

```text
PART: THE SAME RUST-RED HEXAGONAL ORDER PLATE, LIT (the attack plate under the pointer or with the keyboard focus). Single interface piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Image 1 is the idle red plate. Repaint it with exactly the same flat-topped hexagonal outline, size (about 1205 x 954 px), position on the canvas, rim width, screw positions, screw slot angles, chips, flakes and rust; change ONLY the state of the metal, exactly as image 2 changes the grey plate: the brass rim and the two screw heads rubbed bright and warm (rim mean about RGB 170,134,74, worn warm highlights on the upper edges about RGB 214,178,108, tarnish only left in the corners and scratches), and the red-oxide paint of the face one step lighter and cleaner (mean about RGB 128,60,44 instead of RGB 104,46,34). Still old, scratched and chipped; not shiny, not glossy, no white glare, no glow.
BLANK FACE: no symbol, no icon, no lettering, no numbers.
Alpha zero everywhere outside the plate: no ground, no halo, no glow, no cast shadow, no checkerboard. Nothing else on the canvas.

Attachments: Image 1 is the idle red plate to repaint (keep everything but the state of the metal). Image 2 is the lit grey plate: copy its rubbed bright brass rim and screws, not its grey paint.
```

icons:

```text
PART: A SHEET OF SEVEN STENCIL SYMBOLS IN WORN BONE-WHITE PAINT. Sprite sheet, LANDSCAPE 1536x768 (2:1) canvas, true transparency: a grid of 4 columns by 2 rows of equal 384x384 cells, one symbol centred in each cell, the last cell (bottom right) EMPTY.
Each symbol is brushed through a military stencil in off-white paint, as the arrow on the first plate of image 1: bold, simple, solid silhouettes with a few small stencil bridges (gaps), then worn by years of use: chipped, scratched, partly flaked, slightly soft overspray at the edges. Paint colour bone white about RGB 214,206,186. Each symbol about 250 px across at its widest (about 65 % of its cell), centred in its cell, upright, the same visual weight for all seven.
Row 1, left to right:
1. MOVE: one bold straight arrow pointing straight UP, a broad arrow head on a thick shaft.
2. ATTACK: two crossed rifles with fixed bayonets, forming an X, muzzles at the top (as in image 1, second row, first plate).
3. CONTENTS: an open wooden supply crate with its lid tilted open, seen from the front and slightly above.
4. MOVE OUT: an open doorway (a door frame with the door swung open) with a bold arrow pointing RIGHT out of it.
Row 2, left to right:
5. PRODUCE: a heavy hammer, its head at the upper right, its handle running down to the lower left.
6. REPAIR: an open-ended spanner (wrench), its jaw at the upper right, the handle running down to the lower left.
7. CANCEL: a bold X of two crossed straight bars.
8. (empty cell: nothing at all)
Only the paint itself is opaque: alpha zero everywhere else, including inside the symbols' holes, the stencil gaps and the worn spots. No plate, no steel, no background, no frames around the symbols, no halo, no glow, no drop shadow, no checkerboard, no lettering, no numbers.

Attachments: Image 1 is the concept sheet of the order menu's parts: take from its plates only the style of the off-white stencil symbols (worn stencil paint) and the crossed rifles, hammer, spanner, crate, door and X motifs; copy none of the plates themselves. Image 2 is the game's start screen: its stencilled title shows the same worn bone-white stencil paint. Paint ONLY the seven symbols, alone on their transparent canvas.
```

## Avoid
- baked text, numbers or values
- glossy bevels, glow, light rays, clean new metal
- a symbol painted on the plate (the game lays it on)
- cast shadows, ground or halos around a part
