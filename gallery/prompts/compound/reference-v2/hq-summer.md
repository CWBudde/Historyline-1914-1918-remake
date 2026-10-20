# art/recipes/compound/reference-v2/hq-summer.json

Recipe: `art/recipes/compound/reference-v2/hq-summer.json`

- **date**: 2026-10-16
- **edit_target**: art/source/compound/reference-v2/hq-summer-02.png
- **framing**: full source; scene framing reference-layout; no old 8px shrink/fade; final complementary tile ownership
- **output**: art/source/compound/reference-v2/hq-summer.png
- **output_sha256**: 4c907dbacd4751cc1bf4f9a4a29294d788bfa05911d1e21347d66faf14a78c32
- **reference**: build/redraw-reference/hq-summer.png (private original assembly)
- **schema_version**: 1
- **semantics**: Seven joined HQ pieces, four pale stone towers and walls, red north gate, bridge and moat, two rear bushes, transparent courtyard
- **status**: selected candidate; actual 2x/4x seven-tile assembly verified; user visual approval pending
- **tool**: built-in image_gen

## Prompt

```text
Use case: precise-object-edit. Edit target image1 is the painted seven-tile fortress HQ master. KEEP the complete existing fortress architecture, exact four corner tower locations and sizes, square enclosure, stone and wood textures, red-roof north gatehouse, transparent courtyard, turquoise water and bridge and the two small round bushes at the bottom. Change ONLY the road's upper endpoint: extend the existing straight vertical gray entrance road UP to touch the exact TOP CANVAS EDGE at its current center x=50%, preserving its current width, pale edge stones, texture and vertical alignment. The road must connect to the next map tile at the top edge, not end before it; leave no transparent gap above the gray road. Keep all other artwork and the canvas exactly as close to unchanged as possible. Background and courtyard remain genuinely transparent, do not fill ground. This road extension is the only change; no new structures, no text, no tile grid.
```
