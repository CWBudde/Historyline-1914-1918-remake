# depot-summer-02

Recipe: `art/recipes/compound/distinct-v1/depot-summer-02.json`

- **date**: 2026-10-15
- **geometry_reference**: docs/art/compound-building-geometry.md
- **id**: depot-summer-02
- **output_manifest**: art/compound-depot-v1.json
- **schema_version**: 1
- **source**: art/source/compound/distinct-v1/depot-summer.png
- **source_sha256**: e1f78ec16fc736ae4bc0974b0a7ccc56b2d672ef98584def69572c410ce3e91a
- **status**: reviewed-source-for-provisional-compound-exports
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/compound/distinct-v1/studies/depot-summer-01.png` | newly-authored-style-only | `21a8b6c47cad3bcf3cec60bb183980d3e01e2f0c5267cfd6970700afaeb69175` |
| `art/source/compound/distinct-v1/guides/depot-layout-guide.png` | numeric-layout-guide-no-original-pixels | `93a1b6c8c57283509cb408c1d6fbed58b387cee0599fd45ea154551b23ff41fb` |

## Prompt

```text
Create corrected NEW depot master. Image1 is ONLY reference for freshly invented warehouse architecture/materialstyle; do NOT keep its old oversized roof placements. Image2 is the AUTHORITATIVE NUMERIC LAYOUT GUIDE. Output EXACT3:2 landscape1536x1024 with true transparentbackground. Repaint FOUR overheadroof buildings contained INSIDE the four colored rectangles in image2: blue=northroof, green=westroof, orange=eastroof,purple=southroof. Roof shadows, chimneys and small eaves mayextend only a fewpixels, still safely inside the lightgrey fourhex union. The blue andpurple roofs MUST be considerably NARROWER and SHORTER thanimage1 and moved inward to their guide rectangles. Centre/size for northroof x768,y260,width384,height132px; west x342,y512,width120,height300px; east x1194,y512,width120,height300px; south x768,y764,width384,height132px. Treatthese as maximum roofboxes, notoutlinedboxes to draw. Keep a coupleof pixels of margin INSIDE each. North/mainwarehouse hasone small courtyard-side canopy inside union. Courtyardroughsoil uses tan centralguide area and thinlinks tobuildings, with a few tinycrates/barrels nearwalls. STRICT NADIR straight-down orthographic roofplan, practically no visiblefacades. Continuous detailed muted reddishclay/slateroofs andtimber, quiet overcast upperleftlight, smallcloseshadows. Depotcirca1914. Replace ALL guide colors/gridlines/hexedges/greyfill by illustration or REALTRANSPARENCY; no visibleguideoroutlinedhex. Outerunusedarea completelytransparent, noopaquegrassorbackground. Inventsimpleperiodbuildings withroofdetailthatreadsclearlyat476x320. Nopeople,vehicles,text,flags,pixelart. Preserve3:2canvasgeometryandwidepadding; nevercutroofs atcanvasorunionbounds. Image2layout takespriorityoverImage1composition. Fullhighresolution, transparentcutout.
```
