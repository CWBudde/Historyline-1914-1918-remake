---
id: screens/armour
title: Start screen armour plate
kind: screen
status: candidate
runtime_ids:
masters: plate.png, title.png, nameplate.png, nameplate-polished.png, embrasure.png, rivets.png, plaque-germany.png, plaque-france.png
---

## Look
The web edition's start screen and shell pages as the inside of a field-grey riveted steel gun shield
(the "armour plate" concept draft chosen on 2026-11-12): thick chipped grey-green plate with rust
bleeding from dome rivets, the title stencilled in worn bone-white paint, the menu entries on cast
brass nameplates fixed with slotted screws (dull when idle, rubbed bright when selected), two enamel
flag plaques and a heavy embrasure frame with its shutter pushed open, through which the menu's
landscape painting shows. Worn out and sad rather than heroic. Every part is painted alone; all text
but the title is live.

## Base prompt
Every part was generated with this style text, the part's text below, and the concept draft
(`build/review/gen-2026-11-12/start/start-01.png`, not committed) attached as image 1.

```text
Newly authored interface art for the start screen of a turn-based World War I hex strategy game, 1914-1918.
Hand-painted digital illustration with believable, physical materials, not a photograph, not a glossy 3D render, not a cartoon, not pixel art, not flat vector UI. Look and feel: 1914-1918 military hardware. Heavy metal, army issue, worn by field use; serious and handsome, never cheap, never plastic, no modern sci-fi glow, no neon, no gradients that look like web buttons.
MOOD: worn out, sad and exhausted. No heroism, no glory, no clean parade-ground metal. Everything has been used too long in rain and mud: dented, scratched and re-painted steel, rust bleeding from rivets, dull tarnished brass with only the touched spots still bright, chipped enamel, dried mud, soot. Desaturated cold palette of field-grey, wet earth, rust and bone white. Soft overcast light from the upper left; quiet and heavy rather than dramatic. Do not copy any existing game's graphics. No watermark, no logo, no small print.
```

The attachment line of every part:

```text
Attachment: Image 1 is the chosen concept draft of the whole start screen. Take from it the same materials, wear and palette (its grey-green riveted armour plate, its cast brass nameplates, its enamel flag plaques, its stencil paint). Paint ONLY the one part described above, alone on its canvas, not the whole screen; copy none of its lettering, layout or scenery unless this prompt asks for it.
```

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| plate | | `start-parts-plate1` (exec-cc64153b, 2026-11-12), 1672x941 opaque, mean RGB 77,74,59, sdL\* 5.0, even over a 4x2 grid (L\* 30.9-31.7): the bare plate, no rivets or seams, so any crop works. |
| title | | `start-parts-title1` (exec-c06072a0), 1536x1024 RGBA, lettering 1420x303, spelled HISTORYLINE / 1914 - 1918: worn stencil paint. |
| nameplate | | `start-parts-name1` (exec-40015ca6), 1536x1024 RGBA, plate 1446x236, mean RGB 102,79,50: the blank dull brass plate. |
| nameplate-polished | | `start-parts-name2` (exec-3aa4a066), 1536x1024 RGBA, plate 1446x233, mean RGB 158,117,65: the same plate rubbed bright; `nameplate.png` attached as image 2. |
| embrasure | | `start-parts-emb1` (exec-0453baab), 1672x941 RGBA, frame 1605x866, clear opening and margin, mean RGB 74,64,50. |
| rivets | | `start-parts-rivet1` (exec-108bc84a), 1774x887 RGBA, 4x2 heads of about 183 px with rust runs of 220-268 px. |
| plaque-germany | | `start-parts-de1` (exec-613f42e4), 1536x1024 RGBA, plaque 1410x898: black-white-red enamel. |
| plaque-france | | `start-parts-fr1` (exec-73de1a56), 1536x1024 RGBA, plaque 1367x874: blue-white-red enamel. |

plate:

```text
PART: THE BARE ARMOUR PLATE (screen background). Surface painting, LANDSCAPE 16:9, 1672x941 pixels, opaque, filling the whole canvas edge to edge.
Subject: the inside face of one thick rolled armour steel plate of a field-grey gun shield, seen straight on and flat: the same grey-green painted steel as the plate in image 1. Chipped and re-painted field-grey paint, mean colour about RGB 72,75,64; lighter worn and scuffed patches about RGB 104,104,90; dark scratches, pits and bare steel about RGB 38,38,32. A few shallow dents, long fine scratches, soot and oil smudges, faint dried mud spatter towards the lower edge, small rust blooms where paint has flaked and thin rust runs (rust about RGB 110,62,34).
The detail is spread evenly over the whole canvas with no focal point, so that ANY crop of it looks right: the full width, or a tall narrow crop from the middle for a phone screen. No single mark larger than about 1/8 of the canvas width.
Strictly NOT on it: no rivets, no bolts, no screws, no seams, no weld lines, no plate edges or corners, no window, no opening, no frame, no shutter, no hinges, no objects, no lettering, no numbers, no stencils, no flags, no vignette, no dark corners, and no light gradient across the canvas (flat even overcast light). Calm enough that brass nameplates and off-white stencil lettering read clearly on top of it.

Attachment: Image 1 is the chosen concept draft of the whole start screen. Take from it the same materials, wear and palette (its grey-green riveted armour plate, its cast brass nameplates, its enamel flag plaques, its stencil paint). Paint ONLY the one part described above, alone on its canvas, not the whole screen; copy none of its lettering, layout or scenery unless this prompt asks for it.
```

title:

```text
PART: THE TITLE LETTERING. Lettering piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Exactly this lettering, spelled exactly, and no other words, letters or marks:
- first line: HISTORYLINE (H-I-S-T-O-R-Y-L-I-N-E, eleven letters, one word)
- second line, centred under it, letters about 40 % as tall as the first line: 1914 - 1918
Both lines in worn off-white army stencil paint as on the title of image 1: bold serif stencil capitals with stencil bridges (small gaps) in the letters, brushed through a stencil onto steel and then worn by years of use: chipped, scratched, partly flaked, thinner in places, slightly soft overspray at the edges. Paint colour bone white about RGB 214,206,186, with worn spots where the paint is gone.
Layout: the first line spans about 1400 px (about 91 % of the canvas width), centred, capitals about 220 px tall; the two lines together centred vertically on the canvas.
Only the paint itself is opaque: alpha zero everywhere else, including inside the letter counters, the stencil gaps and the worn spots. No steel plate behind the letters, no ground, no halo, no glow, no drop shadow, no checkerboard.

Attachment: Image 1 is the chosen concept draft of the whole start screen. Take from it the same materials, wear and palette (its grey-green riveted armour plate, its cast brass nameplates, its enamel flag plaques, its stencil paint). Paint ONLY the one part described above, alone on its canvas, not the whole screen; copy none of its lettering, layout or scenery unless this prompt asks for it.
```

nameplate:

```text
PART: ONE BLANK BRASS NAMEPLATE, DULL (a menu entry, unselected). Single interface piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Subject: one blank cast brass nameplate seen straight on, centred on the canvas: a long horizontal strip 1440 px wide and 240 px tall (6:1; nearly the full canvas width, 48 px margin left and right), with small chamfered (cut-off) corners, a raised narrow bevelled rim about 14 px wide all around, a flat face, and one slotted brass screw head about 56 px across near each end, centred vertically, about 70 px in from each end; the screw slots at different angles.
BLANK: no lettering, no engraving, no numbers, no symbols, no marks in the middle; the game engraves its text later.
Brass: dull, tarnished, old cast brass like the UNSELECTED plates (CODE, EXTRAS) of image 1: mean colour about RGB 118,94,52; darker brown-green tarnish in the scratches and around the screws about RGB 70,58,34; fine scratches, a few small dents and nicks on the rim, dark grime in the rim corners, only faint dull highlights on the upper rim.
The face's wear is spread evenly along its length so that the middle can be stretched sideways: no single feature in the middle third.
Alpha zero everywhere outside the plate: no ground plate, no steel behind it, no halo, no glow, no cast shadow, no checkerboard. Nothing else on the canvas.

Attachment: Image 1 is the chosen concept draft of the whole start screen. Take from it the same materials, wear and palette (its grey-green riveted armour plate, its cast brass nameplates, its enamel flag plaques, its stencil paint). Paint ONLY the one part described above, alone on its canvas, not the whole screen; copy none of its lettering, layout or scenery unless this prompt asks for it.
```

nameplate-polished:

```text
PART: THE SAME BLANK BRASS NAMEPLATE, POLISHED (the selected menu entry). Single interface piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Image 2 is the dull version of this exact nameplate. Repaint it with exactly the same outline, size (1446 x 236 px), position on the canvas, chamfered corners, rim, screw heads, screw slot angles, dents and scratches; change ONLY the state of the brass: it is the plate a hand touches every day, like the selected START plate of image 1. Rubbed, warmer and brighter brass: mean colour about RGB 168,132,72; broad warm worn highlights along the upper rim and across the middle of the face where it was rubbed (about RGB 214,178,108), the tarnish remaining only in the scratches, in the rim corners and around the screws. Still old, scratched and dented; not shiny new gold, not a mirror, no white glare, no glow.
BLANK: no lettering, no engraving, no numbers, no symbols on the face.
Alpha zero everywhere outside the plate: no ground plate, no halo, no glow, no cast shadow, no checkerboard. Nothing else on the canvas.

Attachment: Image 1 is the chosen concept draft of the whole start screen. Take from it the same materials, wear and palette (its grey-green riveted armour plate, its cast brass nameplates, its enamel flag plaques, its stencil paint). Paint ONLY the one part described above, alone on its canvas, not the whole screen; copy none of its lettering, layout or scenery unless this prompt asks for it.
```

embrasure:

```text
PART: THE EMBRASURE FRAME (a window frame; the game shows its own landscape painting through the opening). Single interface piece, LANDSCAPE 16:9, 1672x941 pixels canvas, true transparency.
Subject: the heavy steel frame of a wide horizontal embrasure (vision slit) cut through armour plate, seen straight on, like the embrasure in image 1 but alone: a thick riveted steel frame around a wide rectangular opening with rounded corners. The frame band is about 64 px wide on the left, right and bottom, with dome-head rivets spaced evenly around it (about 30 px across); the top band is deeper, about 210 px, and carries the heavy sliding shutter plate pushed up and open on a horizontal rail, with a latch and a worn steel pull handle near its right end, as in image 1.
Geometry: the frame's outer edge runs about 24 px inside the canvas edges. The opening's edge runs about 88 px inside the left, right and bottom canvas edges and about 240 px below the top canvas edge; its corners are rounded with a radius of about 36 px. The inner lip of the opening shows a dark bevel, the depth of the thick steel.
Inside the opening: alpha zero, fully transparent, nothing visible through it (no landscape, no sky, no glass, no colour). Outside the frame's outer edge: alpha zero too (no plate around the frame, no ground).
Material: the same chipped field-grey green paint over steel as in image 1 (mean about RGB 78,80,68), rust bleeding from the rivets and along the shutter rail, dark grime in the corners, scratches and dents. No lettering, no numbers. No halo, no glow, no cast shadow outside the frame, no checkerboard.

Attachment: Image 1 is the chosen concept draft of the whole start screen. Take from it the same materials, wear and palette (its grey-green riveted armour plate, its cast brass nameplates, its enamel flag plaques, its stencil paint). Paint ONLY the one part described above, alone on its canvas, not the whole screen; copy none of its lettering, layout or scenery unless this prompt asks for it.
```

rivets:

```text
PART: A SHEET OF EIGHT ARMOUR RIVETS. Sprite sheet, LANDSCAPE 1536x768 (2:1) canvas, true transparency: a grid of 4 columns by 2 rows of equal 384x384 cells, one rivet in each cell, its head centred horizontally and about 140 px below the top of the cell.
Each rivet: a large round dome-head armour rivet seen straight from the front, its head about 120 px across (about 1/13 of the canvas width, under a third of the cell width), perfectly round outline, a low dome. Painted the same field-grey green as the plate in image 1 (about RGB 80,82,70), the paint worn off the crown where dark bare steel shows (about RGB 50,50,44), a thin dark crevice ring where the head meets the plate. Below each head a rust stain bleeding straight down from it, about 34 px wide and 130 px long, orange-brown at the top (about RGB 120,66,34) and fading out to nothing, as on the rivets of image 1. Soft highlight on the upper-left shoulder of the dome, shade on the lower right. Every rivet differs a little in wear and in the length of its rust stain.
Alpha zero everywhere outside the rivet head and its rust stain: no plate, no washer, no ground, no halo, no glow, no cast shadow, no checkerboard; the rust stain is semi-transparent where it fades.

Attachment: Image 1 is the chosen concept draft of the whole start screen. Take from it the same materials, wear and palette (its grey-green riveted armour plate, its cast brass nameplates, its enamel flag plaques, its stencil paint). Paint ONLY the one part described above, alone on its canvas, not the whole screen; copy none of its lettering, layout or scenery unless this prompt asks for it.
```

plaque-germany:

```text
PART: ONE ENAMEL FLAG PLAQUE, GERMAN EMPIRE. Single interface piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Subject: one small enamel plaque seen straight on, centred on the canvas: a rectangular steel plate 1240 px wide and 820 px tall with slightly rounded corners and a thin raised steel rim about 24 px wide, a small dome rivet about 46 px across in each corner of the rim. Its face is fired enamel showing the flag of the German Empire of 1914: three equal horizontal bands, black on top, white in the middle, red at the bottom; the colours aged and dull (black about RGB 32,32,30, bone white about RGB 214,206,186, faded red about RGB 150,42,34).
Wear as on the plaques of image 1: enamel chipped at the edges and corners showing dark steel and small rust spots, a few hairline crazing cracks, grime and dried mud in the rim, a scratch or two across the face. Nothing else on it: no eagle, no crest, no cross, no lettering.
Alpha zero outside the plaque: no ground, no plate behind it, no halo, no glow, no cast shadow, no checkerboard.

Attachment: Image 1 is the chosen concept draft of the whole start screen. Take from it the same materials, wear and palette (its grey-green riveted armour plate, its cast brass nameplates, its enamel flag plaques, its stencil paint). Paint ONLY the one part described above, alone on its canvas, not the whole screen; copy none of its lettering, layout or scenery unless this prompt asks for it.
```

plaque-france:

```text
PART: ONE ENAMEL FLAG PLAQUE, FRANCE. Single interface piece, LANDSCAPE 1536x1024 (3:2) canvas, true transparency.
Subject: one small enamel plaque seen straight on, centred on the canvas: a rectangular steel plate 1240 px wide and 820 px tall with slightly rounded corners and a thin raised steel rim about 24 px wide, a small dome rivet about 46 px across in each corner of the rim. Its face is fired enamel showing the flag of France: three equal VERTICAL bands, blue at the left (hoist), white in the middle, red at the right; the colours aged and dull (faded blue about RGB 40,58,112, bone white about RGB 214,206,186, faded red about RGB 150,42,34).
Wear as on the plaques of image 1: enamel chipped at the edges and corners showing dark steel and small rust spots, a few hairline crazing cracks, grime and dried mud in the rim, a scratch or two across the face. Nothing else on it: no emblem, no crest, no lettering.
Alpha zero outside the plaque: no ground, no plate behind it, no halo, no glow, no cast shadow, no checkerboard.

Attachment: Image 1 is the chosen concept draft of the whole start screen. Take from it the same materials, wear and palette (its grey-green riveted armour plate, its cast brass nameplates, its enamel flag plaques, its stencil paint). Paint ONLY the one part described above, alone on its canvas, not the whole screen; copy none of its lettering, layout or scenery unless this prompt asks for it.
```

## Post
Masters are used unchanged. `hl assets build-screens` puts them in the screen catalog's `armour`
family (`internal/hlasset/armour.go`): each part loses its alpha haze (alpha 16 and below), is trimmed
to what is left and to an even size, and is offered at a logical size of half the trimmed master
(2x is the master itself), smaller for the parts drawn small (nameplates 0.3, plaques 0.12, rivets 0.22
of the master). The rivet sheet becomes two strips that repeat: `armour/rivets-row` (the eight rivets
side by side at the sheet's cell pitch) and `armour/rivets-column` (stacked at their tallest rust run).
The web shell (`webapp/app/src/shell`) covers the window with `armour/plate`, lays the rivet strips along
its edges, draws the nameplates as 9-slice borders under live engraved text and shows
`backdrop/western-front` through the embrasure's opening.

## Avoid
- rivets, seams, plate edges, windows or lettering in the plate
- lettering, engraving or marks on the nameplates
- landscape, glass or colour in the embrasure's opening; a plate around its frame
- crests, eagles or lettering on the plaques
- any word but HISTORYLINE and 1914 - 1918 in the title
