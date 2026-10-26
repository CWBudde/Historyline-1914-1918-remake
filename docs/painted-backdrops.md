# Painted backdrops

A scenario file may name a backdrop (`"backdrop": "1914-08.jpg"` next to the file): one painting of the
whole level's ground that the game draws in place of the terrain tiles (`-backdrop off` draws the tiles).
It must match the level's world aspect; any resolution works. Shipped: PULSE (`1914-08.jpg`, 1536×2706),
CIVIL (`1914-10.jpg`, 2336×2706) and Marne (`1914-09.jpg`, 3137×1578, six sections in a 3×2 grid),
all at 1536/1090 ≈ 1.41 image px per world px. MOUSE/SPORT (`1914-12.jpg`, 2852×3302) uses about
1.72 image px per world px, with settled winter snow and cold overcast daylight. VENOM/BIMBO
(`1915-02.jpg`, 2852×4403) uses the same density, with patchy late-winter snow and clear frosty light.

Imagegen outputs are about 1.57 megapixels (1536×1024 or 1254×1254), far less than a whole map. A
backdrop is therefore painted in overlapping sections, each registered to a render of the level, and
stitched.

## 1. Layout guide

Render the level's ground with the current tile art:

```sh
go run ./cmd/hlart map -level levels/monthly/1914-10.json:cp -out build/review/painted-civil/guide   # map-ground.png, 4x
```

The render is the layout every section must follow; the tiles' flat look, hex-shaped patches and black
border notches are not to be copied.

## 2. Sections

Cut the world into sections of 1090×727 world px (3:2, one 1536×1024 canvas each, so one hex is about
64 canvas px), neighbours overlapping by at least 130 world px (rows by 131, CIVIL's columns by 522):

| Map                  | World px  | Sections (world origin)                      |
| -------------------- | --------- | -------------------------------------------- |
| PULSE, 8 hexes wide  | 1090×1920 | 1 column × 3 rows: y 0, 596, 1193            |
| CIVIL, 12 hexes wide | 1658×1920 | 2 columns × 3 rows: x 0, 568; y 0, 596, 1193 |

MOUSE/SPORT uses nine 728×727 world-pixel sections: three columns at x 0, 465, 930 and three rows at
y 0, 596, 1193. The smaller sections provide its higher image density.

VENOM/BIMBO uses twelve of those sections, the same columns and four rows at y 0, 611, 1222, 1833
(116 world px of row overlap). The northwest shore includes a small cove within the transport's water
hex; the northeast seam retains the harbour basin's open channel into the lake.

Crop each section from the 4x guide and scale it to the section canvas: 1536×1024 for PULSE/CIVIL,
1252×1250 for MOUSE/VENOM.

## 3. Painting a section

One `/gfx-imagegen` run per section (`gen.sh`, prompt from `assemble.sh terrain`):

- Image 1: the section's guide crop, "the exact layout, followed position for position": every road,
  river, lake, forest edge, crater, building and fort stays where it is, at the same size and outline
  (within about 10 px), so the painting registers with the hex grid.
- Image 2 (and 3): style references only, never layout: the target mock-up, for later maps a section of
  an installed backdrop so the maps match, for a season a material sample such as winter grass.
- The prompt says the canvas is one section of a larger map, its size in canvas px and world px, the
  hex size on the canvas (about 64 px), the season and the light.

Because every section is pinned to the same guide, neighbours agree in their overlaps. Run a second
round (fresh, or from the guide instead of an earlier painting) and pick the better try per section;
picks from different rounds stitch just as cleanly.

## 4. Stitching

Scale each pick to its section's size in the output and stack the sections:

1. Colour: each section gets an RGB gain so that the mean of its overlap matches its neighbour's,
   chained along the stack and normalised to a mean gain of 1.
2. Seam: in each overlap, a dynamic-programming path (one row step per column) of least colour
   difference between the two sections, on a 7×7-blurred cost so it follows broad agreement rather than
   single pixels. Cutting where the paintings agree avoids doubled features.
3. Blend: a smoothstep cross-fade of ±10 px around the seam.

A two-column map is stitched per row along x first (transpose, stack, transpose back), then the rows
along y. Check the joins at full resolution and with a hex-grid overlay for registration.

## 5. Finishing

- Grade towards the target look: a mild desaturation and gain (PULSE: 0.2 and 0.9; 0.4 turned it grey
  and lost the warmth). Compare L\* and chroma with the target.
- Objects on exact hexes (bunker emplacements): imagegen misplaces asked-for objects inside a section
  by 80 to 500 px. Paint the section without them, generate the objects as a separate top-down sheet
  and paste them on their hexes with a round, feathered blend and a gain matched to the surroundings;
  never rotate or transpose them.
- Save as JPEG next to the scenario file and set `backdrop`; check it in the game with units on top at
  normal and close zoom.

A painting follows the map it was painted from: when a month's map changes, the changed hexes (stones,
trenches, mountains) still show the old ground until they are repainted or patched.
