# art/recipes/compound/reference-v2/hq-winter.json

Recipe: `art/recipes/compound/reference-v2/hq-winter.json`

- **date**: 2026-10-16
- **edit_target**: art/source/compound/reference-v2/hq-summer.png
- **framing**: reference-layout; inset1; no old8pxshrink/fade; finalcomplementarytileownership
- **invariants**: Exact summer geometry, four towers and enclosing walls, red gate, north road/bridge/moat, two rear bushes, transparent courtyard
- **output**: art/source/compound/reference-v2/hq-winter.png
- **output_sha256**: 9a9b09325ff426c4b4c092d26dd98c6ce4c6cf093d8794a0e9e7307f39757583
- **reference**: build/redraw-reference/hq-winter.png (private original assembly; snow/color reference)
- **schema_version**: 1
- **status**: selected candidate; actual 2x/4x seven-tile assembly verified; user visual approval pending
- **tool**: built-in image_gen
- **transparent_background**: true

## Prompt

```text
Use case: lighting-weather seasonal game sprite edit. Image1 is the selected NEW summer fortress HQ master, the edit target: preserve its exact canvas, positions, four gray stone corner towers, crenellated connecting walls, red north gatehouse, road extending to the TOP edge, bridge, moat and two small round rear bushes. Image2 is the original assembled WINTER HQ, a snow/color reference only; do not copy its low-resolution pixels or change Image1's structure. Create the WINTER counterpart of Image1 in the same detailed continuous painted style and same high resolution, with powder snow on upper-facing battlements, tower roofs, stone edges, the red gate roof and the two rear bush crowns. Stone shadows stay dark, the road remains readable under light patchy snow, bridge route stays clear and turquoise water remains visible just as in the original winter reference. Keep enough red roof showing for the north gate to read at native size. Keep two distinct green-and-snow bushes. No snow slab across the courtyard: the courtyard opening and every background area MUST stay transparent so the game's winter ground shows through. Same orthographic perspective and geometry across both seasons. No added buildings, no text or hex grid or border. Genuine transparent background.
```
