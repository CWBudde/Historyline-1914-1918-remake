# ground-rocky

Recipe: `art/recipes/terrain/distinct-v1/ground-rocky.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source preserved unchanged. Exact centered 6:5 crop then Lanczos3 exports; no original reference or mask. Rectangular overscan; hex clipping, texture phase and connections remain unaccepted.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: ground-rocky
- **review**: Batch02 review in docs/art/distinct-terrain-batch-02.md; source candidacy does not establish runtime replacement or seamlessness.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/ground-rocky.png
- **source_sha256**: f59005d7033d85c7f22e2195a169114ea930a4646384847d2194408d102974bd
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/meadow.png` | newly-authored-material-style-only | `8c8935efbbd463e39d49208e42ca0e87b140268a1b11941764dfa7c29c4fff85` |

## Prompt

```text
Create ONE original high-resolution terrain illustration for a European1914 strategy game. Image1 is solely the NEWLY AUTHORED meadow's fine continuous painted material/style reference, not an edit target. Strict vertically overhead orthographic terrain view, landscape canvas6:5, no horizon or perspective, consistent diffuse overcast light with subtle upperleft highlights. Entire opaque filled rectangle with overscan, quiet at192x160 but finely detailed at full resolution. No game UI, border, vignette, hex silhouette, labels, pixel art, blocky shapes or legacy art reconstruction. Subject ROCKY ROUGH GROUND: muted grey limestone and warm greige rubble interleaved with dusty muted earth and sparse sage grass under15percent. Low irregular exposed stone slabs and clusters of tiny angular fragments at aerial map scale, diverse sizes, no giant central rock or isolated boulder, no mountains or cliffs. Rock material occupies about70percent of field; evenly distributed irregular broad shallow patches, crisp subtle fissures with limited dark gaps. Read as distinct stone terrain at192x160, no black shadows. Broad exposed ground surface not gravel road or beach, no tracks or human objects, no geometric tiling or flagstones. Match reference quiet lighting and finesse, invent fresh spatial composition.
```
