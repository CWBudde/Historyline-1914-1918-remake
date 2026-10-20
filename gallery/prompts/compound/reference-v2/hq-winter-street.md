# art/recipes/compound/reference-v2/hq-winter-street.json

Recipe: `art/recipes/compound/reference-v2/hq-winter-street.json`

- **date**: 2026-10-16
- **edit_target**: art/source/compound/reference-v2/hq-winter-natural.png
- **edit_target_sha256**: 37a03c5bdee507d2f14d9af1a9d703a1ab9785a3be531d96d860f0ecf08a407f
- **framing**: reference-layout; inset1; complementary whole-scene tile ownership; dirt courtyard and irregular shrubs retained
- **output**: art/source/compound/reference-v2/hq-winter-street.png
- **output_sha256**: 1fb27001f23797ac98f36b95baf9cc5b6f1f5459df32a18e980b485f1cc96204
- **schema_version**: 1
- **status**: selected candidate; winter entrance widened for the current marked road
- **tool**: built-in image_gen
- **transparent_background**: true

## Prompt

```text
Use case: precise-object-edit. Image1 is the selected WINTER fortress HQ master, the edit target. Image2 is the ACTUAL newly selected game road tile, supporting road surface/width reference only. Change ONLY the very TOP portion of the north entrance approach so it joins the real wider paved road. Keep its exact centerline x=621.5 on the1243wide canvas, touching the TOP EDGE y=0. At y=0 paint a fully opaque continuous grey-mauve paved surface spanning at least x=546..696 (150pixels total, including thin dark shoulders), so there is ample opaque carriageway where the regular24-world-pixel road meets it. Keep that full width through the first70pixels, then gently taper into the existing bridge approach by y=200. Match image2's grey surface, not beige gravel. The regular road's centre dashes may end before the bridge; no stone cap, no transparency gap, no parapets extending through the join. Preserve EXACT1243x1265canvas and the rest of image1 as unchanged as possible: all four towers/walls, red snowy gatehouse, bridge and water, dirt courtyard with patchy snow, irregular rear shrubs and exterior transparency. Do not redraw/reposition the fortress or bushes. Only broaden and retone the upper road endpoint. No text, grid, arrows or new objects. Background outside the road and fortress genuinely transparent.
```
