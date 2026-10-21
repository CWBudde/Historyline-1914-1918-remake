# Shared style

Every `art/<key>/prompt.md` base prompt is pasted after the preamble below and the section for its kind.

## Preamble

```text
Newly authored artwork for a turn-based World War I hex strategy game set in rural Europe, 1914-1918.
Continuous hand-painted digital illustration with fine, believable material detail: wool, leather, wood,
steel, concrete, earth, grass, stone and water. Restrained, natural palette and calm values so that
markers and units stay readable on top. Soft, diffuse daylight from the upper left with short, soft
shadows to the lower right. Crisp, antialiased natural edges. Not pixel art, not a pixel enlargement,
not a flat vector, not a cartoon, not a photograph or 3D render. No text, letters, numbers, labels,
logos, watermarks, flags, national insignia, grid lines, hex outlines, frames, borders, arrows,
vignettes or checkerboard patterns. Do not copy or trace any existing game's graphics.
```

## Terrain

The target look is the PULSE reference painting (`art/reference/pulse/`): a lush, sunlit aerial map.

```text
Strict top-down orthographic (nadir) view, no horizon and no perspective tilt. World scale: one map hex
is about 50 metres across; a single tree crown is about a fifth of a hex, a hut about a quarter. Ground
is a warm olive to yellow-green meadow (average close to RGB 96,104,31) with fine blade texture, tiny
straw scratches, small soil flecks and sparse pale seed heads; never neon lime, never a flat fill.
Trees are dense rounded broadleaf and conifer crowns in deep moss to dark teal green with yellow-green
sunlit tops. Water is a clear, saturated blue-teal with small white glints. Paved roads are medium grey
asphalt with dashed off-white centre lines and sandy gravel shoulders; tracks are pale tan earth.
Rocks are warm grey-beige with olive moss. Materials and features fill the canvas edge to edge, or
sit on true alpha transparency for cut-out objects, as each asset states.
```

Rules: same world scale on every tile; no visible tiling, rounded "pillow" repeats or big light/dark
blobs; roads, streams and banks continue straight through their declared ports; winter versions keep
the summer geometry and only change season (settled snow, bare deciduous trees, cool shadows).

Terrain must form a reusable tile set. Use the reference for materials, scale and lighting; paint shared
surfaces and reusable props, then compose them through the declared edge masks. A lake must work in
different compatible shoreline arrangements. Do not replace it with a painting selected for one map's
exact outline. Review individual tiles, several rearrangements and the composed map together.

## Units

```text
Six individually painted views of the same unit on one transparent RGBA atlas of 2 columns by 3 rows of
512x512 cells, no gutters, row-major headings N (up), NW (up-left), SW (down-left), S (down),
SE (down-right), NE (up-right). The front of the unit points toward each cell's heading. Fixed elevated
three-quarter overhead strategy camera, the same in all cells; light from the screen's upper left in
all cells. Same object and person count, equipment and scale in every view; never rotate or mirror a
finished view. Centre each view on (256,261) and keep everything inside the central 65 percent of its
cell, with fully transparent margins. Uniform cloth or vehicle paint that marks ownership is a
saturated, naturally shaded MAGENTA technical colour; leather, skin, wood, horses, concrete and metal
stay natural. Strong light/dark separation so the silhouette reads at 48 pixels. Real alpha
transparency outside the objects and between limbs: no ground, shadow, haze, glow or backdrop.
```

## Portraits

```text
One opaque, full-bleed landscape illustration at 105:94 aspect, ideally 1680x1504, of the unit in a quiet
period scene. Subdued ochre, sepia, khaki, olive and sage values under a pale cloudy sky; painted with
broad connected brushwork and selective fine mechanical or cloth detail, a little stylized like polished
tactical-game illustration. Owner-neutral: no national colours, magenta or faction tint. Keep the
essential subject inside the centred 105:94 crop and readable at 210x188. No frame, bevel or UI.
```

## Interface

```text
Slate and brass interface vocabulary: dark slate panels (#162630), muted blue-grey
hairlines (#6e8088), aged brass accents (#c4a773) and cream highlights (#eee5d2). Flat, clean shapes with
small corner radii and thin bevels; icons are simple solid glyphs in brass or slate. Painted items
(weapons, explosions) use the painted medium above on true transparency. Art never contains labels or
values; all text is drawn live.
```
