# art/recipes/compound/reference-v2/hq-winter-natural.json

Recipe: `art/recipes/compound/reference-v2/hq-winter-natural.json`

- **date**: 2026-10-16
- **edit_target**: art/source/compound/reference-v2/hq-winter-courtyard.png
- **edit_target_sha256**: 82d28d02765f651860c2812555794578c2de8f4b1ab0326183300ff0c529697d
- **framing**: reference-layout; inset1; final complementary tile ownership
- **output**: art/source/compound/reference-v2/hq-winter-natural.png
- **output_sha256**: 37a03c5bdee507d2f14d9af1a9d703a1ab9785a3be531d96d860f0ecf08a407f
- **schema_version**: 1
- **semantics**: Dirt courtyard, irregular rear shrubs, edge-reaching gravel road on the existing north centerline. Existing stone fortress, bridge/moat and red gate preserved.
- **status**: selected candidate; visual acceptance pending
- **tool**: built-in image_gen
- **transparent_background**: true

## Prompt

```text
Use case: precise-object-edit. Edit target is the winter fortress with a brown dirt courtyard and sparse patchy snow. Change ONLY the TWO shrubs at the very bottom/rear of the fortress. They currently look like identical perfectly circular clipped topiary balls. Replace them with two small naturally grown irregular leafy bushes: asymmetrical branching forms, loose broken silhouettes, a few protruding twigs and leaf clusters, varied dark green foliage with sparse snow on upward-facing leaves, different heights and widths, subtle mutual overlap and natural contact shadows. Preserve their approximate two positions and total footprint below the rear wall, with clear grass/transparent space around them. They should read as two organic low shrubs seen from the same overhead angle, not spherical ornament balls, rosettes or symmetrical trees. Keep the entire rest of the image EXACTLY as close to unchanged as possible: 1243x1265 canvas, patchy-snow dirt courtyard, snow-covered four towers, all walls, red gatehouse, north edge road, bridge and moat, walkways, painting style and lighting. Keep exterior background genuinely transparent. No new objects, text or grid. Also make the very TOP END of the north entrance road match the adjoining real street shown in image 2 (supporting material/width reference). Keep the road exactly centered at x=621.5 on the 1243-pixel-wide canvas and touching y=0. At the top edge use a 74-pixel-wide warm gray-beige gravel carriageway plus approximately 10-pixel earth shoulders on each side (94 pixels total), corresponding to the game's 14-world-pixel road and 2-world-pixel shoulders. Gently taper into the existing wider bridge approach farther down. Do not terminate, offset or cap this road; do not keep projecting stone parapets at its top endpoint. Preserve bridge railings near the water. Match image 2's fine gravel/earth colors at the seam so the entrance joins the regular road.
```
