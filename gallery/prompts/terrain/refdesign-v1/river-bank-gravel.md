# river-bank-gravel

Recipe: `art/recipes/terrain/refdesign-v1/river-bank-gravel.json`

- **art_direction**: refdesign-v1
- **consumer**: connected stream bank only, width18 border3.5 with canonical eight-world-pixel port guards and 1.1-world-pixel interior bank wear; nine lakeside variants keep their existing full parent master
- **date**: 2026-10-20
- **generation_command**: go test -tags riverbankgfx ./tools/art-review -run '^TestGenerateRiverBankGravelCrop$' -count=1
- **id**: river-bank-gravel
- **method**: Decode the pinned 1254x1254 parent PNG. image.Draw Src copies rectangle (400,640)-(784,960) to a 384x320 NRGBA canvas; encode PNG. All 122880 decoded output RGBA pixels are checked against the corresponding decoded parent pixel. No interpolation, recoloring, retouching or raster painting. Horizontal material features become 1254/384=3.265625 times coarser and vertical features 1254/320=3.91875 times coarser under the unchanged lattice material sampler.
- **review**: 384x320 crop selected after comparison with full master and 512x384 crop at 2x and 1.5x assembled scale; larger irregular loose stone detail without a regular paving pattern. Integrated map visual acceptance remains open.
- **schema_version**: 1
- **source**: art/source/terrain/refdesign-v1/river-bank-gravel.png
- **source_sha256**: bec7afdd82d829aa9f3ae08f8e5e06eb8e1b3d1dbec3d8f36c19edfbc99ba0f0
- **status**: selected-source-awaiting-connected-bundle-review
- **tool**: Go image/png and image/draw; deterministic unchanged-pixel crop
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/shore-reference-v2/bank.png` | new-authored-parent-material; exact decoded RGBA crop, no original raster input | `d43b40a2ed64e95cbb6afdacca4a384a39ec59c8a54d05bdcbe77a60e6aa6613` |
