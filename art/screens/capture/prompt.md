---
id: screens/capture
title: Capture scenes
kind: screen
status: candidate
runtime_ids:
masters: depot.png, factory.png, hand.png, hq.png, hq-open.png
---

## Look
The scenes of a captured building, after the original's: in the capturer's panel a dark map room
lights up over a wooden scale model of the building on the staff map, a hand reaches in and plants a
pin with the capturer's flag on the model's roof, then withdraws. A battle won by taking the HQ
closes on a low view up at the enemy headquarters: its doors swing open on defenders who surrender,
and the winner's flag goes up the bare pole on the gable. Warm sepia for the map room, a muted dusk
palette for the headquarters. The pin, the flags and all motion are drawn in code.

## Base prompt
```text
Newly authored artwork for a turn-based World War I strategy game set in rural Europe, 1914-1918.
Continuous hand-painted digital illustration with fine, believable material detail. Not pixel art,
not a pixel enlargement, not a flat vector, not a cartoon, not a photograph or 3D render. No text,
letters, numbers, labels, logos, watermarks, flags, national insignia, frames or borders. Do not copy
or trace any existing game's graphics.
```

Each variant's prompt below opens with this text (the material list varies), then gives the scene.
The edits attach the master they change and keep everything outside the changed area.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| depot | | `cd-01` (exec-b29ee046), 1122x1402 opaque: the prompt below. |
| factory | | `cf-01` (exec-ec2ff3e8), 1122x1402 opaque: an edit of `depot.png`, the prompt below. |
| hand | | `ch-01` (exec-d45fc2fd), 1254x1254 RGBA: the prompt below. |
| hq | | `hq-01` (exec-48a6cf80), 1586x992 opaque: the prompt below. |
| hq-open | | `hq-02` (exec-6ff83a72), 1584x993 opaque: an edit of `hq.png`, the prompt below. |

depot:

```text
Newly authored artwork for a turn-based World War I strategy game set in rural Europe, 1914-1918.
Continuous hand-painted digital illustration with fine, believable material detail: wood, paper,
plaster, brass, wool. Not pixel art, not a pixel enlargement, not a flat vector, not a cartoon, not a
photograph or 3D render. No text, letters, numbers, labels, logos, watermarks, flags, national
insignia, grid lines, hex outlines, frames or borders. Do not copy or trace any existing game's
graphics.

Use case: a short in-game vignette shown when a side captures a supply depot. Generate ONE opaque
PORTRAIT painting, aspect 4:5, ideally 1024x1280 pixels.

Scene: night in a general staff's map room. Close view down onto a large wooden map table covered
with a painted campaign map (fields, roads, a river drawn as soft washes, no writing). In the middle
of the table, filling about the middle half of the canvas width, stands a detailed hand-made wooden
scale model of a military supply depot: one long low brick storehouse with a pitched roof and a
loading ramp, a few stacked crates and barrels beside it, a small timber water tower. The model's
roof ridge is clear and unobstructed, its centre about 40 percent down from the top of the canvas.
A single shaded oil lamp outside the picture at the upper left lights the model and the table
warmly; the light falls off into deep shadow toward all four edges and the corners, the upper fifth
of the canvas almost black. Nothing else on the table: no figures, no pins, no hands, no people, no
papers with writing.

Colour: a warm monochrome sepia painting, like an old toned photograph come alive: blacks of deep
brown (RGB about 30,20,10), mid tones of ochre-brown (about 140,105,60), highlights of cream (about
235,215,170). No other hues.

Composition must leave the area above the roof ridge empty and dark: a small flag will be planted
on the ridge later.
```

factory:

```text
Image 1 is a finished painting. Make ONE edited copy of it, the same canvas size (1122x1402), opaque.

Keep exactly as they are: the oil lamp at the upper left, the dark room, the map table, the painted
map with its fields, roads, river, bridge and village, the light and the warm monochrome sepia
colours, every outline and position.

Change ONLY the wooden scale model standing on the table (the brick storehouse, the water tower, the
wagon, crates and barrels): replace it with a hand-made wooden scale model of a small war factory on
the same footprint and at the same size: a long brick workshop hall with a sawtooth roof of
north-light glazing, two tall round brick chimneys at its left end, a small boiler house, a short rail
siding with one flat wagon and a few stacked steel plates. The hall's main roof ridge runs where the
storehouse's ridge was, clear and unobstructed, its middle at about 57 percent of the width and 30
percent of the height. Same lighting from the lamp, same shadows toward the edges. No people, no
hands, no pins, no flags, no writing, no smoke.
```

hand:

```text
Newly authored artwork for a turn-based World War I strategy game set in rural Europe, 1914-1918.
Continuous hand-painted digital illustration with fine, believable material detail. Not pixel art,
not a flat vector, not a cartoon, not a photograph or 3D render. No text, letters, logos, watermarks,
flags or insignia. Do not copy or trace any existing game's graphics.

Use case: a transparent cut-out layered over a map-table vignette, animated by the game.
Generate ONE image, SQUARE 1024x1024, RGBA with true transparency: alpha zero outside the object, no
ground plate, no halo, no checkerboard, no cast shadow.

Subject: the right hand and forearm of a First World War officer, reaching in from the lower right
corner toward the upper left at about 40 degrees. The forearm sleeve is a plain field-grey wool tunic
cuff with two small brass buttons; the hand is bare. The thumb and index finger pinch together at the
tip as if holding a very thin pin (the pin itself is NOT painted: the fingertips touch an empty
point); the other fingers curl loosely. The pinching fingertips sit at about 30 percent of the canvas
width from the left and 30 percent of the height from the top; the sleeve leaves the canvas through
the bottom right corner, cut off by the canvas edge.

Lighting: a warm oil lamp from the upper left, deep soft shadows on the lower right side of the hand.
Colour: a warm monochrome sepia painting: deep brown shadows (RGB about 30,20,10), ochre-brown mid
tones (about 140,105,60), cream highlights (about 235,215,170). No other hues.
```

hq:

```text
Newly authored artwork for a turn-based World War I strategy game set in rural Europe, 1914-1918.
Continuous hand-painted digital illustration with fine, believable material detail: stone, slate,
timber, iron, glass. Not pixel art, not a pixel enlargement, not a flat vector, not a cartoon, not a
photograph or 3D render. No text, letters, numbers, labels, logos, watermarks, flags, national
insignia, frames or borders. Do not copy or trace any existing game's graphics.

Use case: the full-screen scene shown when a side captures the enemy headquarters and wins the
battle. Generate ONE opaque LANDSCAPE painting, aspect 16:10, ideally 1536x960 pixels.

Scene: a dramatic low-angle view looking up at the front of a grand stone chateau used as an army
headquarters, at dusk under a cloudy sky. The facade rises into the picture with converging
verticals; a steep slate roof with a central gable. On top of the central gable stands a single tall
bare wooden FLAGPOLE with a small gilt finial, completely EMPTY: no flag, no rope with cloth. The pole
stands against a light patch of sky, its foot on the gable at about 50 percent of the width and 22
percent of the height, its top at about 50 percent of the width and 6 percent of the height.
At the bottom centre a wide stone staircase leads up to a pair of tall, heavy, CLOSED wooden double
doors with iron studs, in an arched doorway; the doorway spans about 40 to 60 percent of the width
and about 55 to 95 percent of the height. Tall windows to either side are dark. Bare trees and
sandbags at the far left and right edges. No people.

Colour: muted dusk palette: slate-blue clouds with a pale gap behind the flagpole, grey-ochre stone,
dark brown timber, faint warm lamplight in one upper window. Calm values.
```

hq-open:

```text
Image 1 is a finished painting. Make ONE edited copy of it, the same canvas size (1586x992), opaque.

Change ONLY the inside of the arched doorway at the top of the staircase (the area of the two tall
wooden doors, about x 655 to 925 and y 470 to 825). Everything outside that arch must stay exactly
as it is: every stone, window, lamp, step, tree, cloud and the empty flagpole on the gable, every
outline and position unchanged. Do not add a flag anywhere.

The edit: the two door leaves have swung fully open inward, out of sight into the building, so the
arched doorway is an open passage. Inside, a warm lamp-lit entrance hall recedes: a stone floor, a
staircase and a chandelier glow in the depth. In the doorway, on the threshold, stand three or four
defeated soldiers in plain, drab greatcoats without any insignia, badges, emblems or recognisable
nationality, caps or bare heads (no helmets with spikes), facing the viewer with both hands raised
in surrender; one holds up a plain white cloth. They are small, filling about the lower two thirds of
the doorway's height, lit from behind by the warm hall light so they read partly as silhouettes.
Same painterly style, palette and dusk light as image 1. No text, no flags, no insignia.
```

## Post
Masters are used unchanged. `hlasset build-screens` scales them into the screen catalog's `capture`
family: `capture/depot`, `capture/factory` and `capture/hand` 512 px wide at 1x and 1024 at 2x,
`capture/hq` and `capture/hq-open` to the full screen, 640x400 at 1x and 1280x800 at 2x.
`internal/dospanel` places the code-drawn parts by points measured on the masters, as fractions of
their size: the pin goes into the depot's roof ridge at 0.624, 0.303 and the factory's at 0.526,
0.300; the hand pinches at 0.184, 0.218; the HQ's flagpole stands at 0.500 from 0.199 (foot) to
0.050 (head); the door leaves fill the arch from 0.413 to 0.586 wide, from 0.474 (crown) and 0.544
(springing) down to 0.832. A new master must keep these points or change them in code. The map room
is drawn at the scene's light, the hand toned to sepia; the leaves of `hq.png` swing open over
`hq-open.png`.

## Avoid
- flags, pins, hands or people in the map room's paintings
- a flag on the HQ's pole, insignia or a recognisable nation on the defenders
- writing on the staff map
