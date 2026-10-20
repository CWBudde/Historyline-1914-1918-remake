# grass-base-refdesign-v1

Recipe: `art/recipes/terrain/refdesign-v1/grass-base-01.json`

- **acceptance**: Candidate; full-map visual acceptance remains open.
- **generation_command**: go test -tags grassgfx ./tools/art-review -run TestGenerateGrassVariations -count=1 -v
- **generation_recipe**: art/recipes/terrain/refdesign-v1/grass-meadow-01.json
- **id**: grass-base-refdesign-v1
- **profile**: Shared canonical grass material across terrain families and all reciprocal grass variant guards; unmirrored sampleGrassMaterial lattice sampler with quarter-source cubic-window overlap at source edges, periods142,80, and average of f(x,y),f(x+71,y+40). Sixteen variant phases use the same selected material and feature scale, with world offsets(37*index mod142,23*index mod80). No tint correction, sharpening or procedural marks. Winter, water and unseen retain the classic sampler.
- **schema_version**: 1
- **source**: art/source/terrain/refdesign-v1/grass-base-01.png
- **source_sha256**: 8e751025b89fe32233f01a9f8090dd280fb702ebbd9954f279583ba732fa8a3b
- **status**: candidate
- **tool**: Go offline grassgfx exporter
- **transparent_background**: false
- **visual_proof**: build/grass-refdesign-v1/patch-game-1.5x.png

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/refdesign-v1/grass-meadow-01.png` | New generated material master. Canonical derivative is its768x480 center crop at source origin(384,272), normalized to zero image bounds, encoded without filtering or retouching. Selected from three crop-scale proofs at game zoom1.5 to preserve readable seed flecks and small tufts while removing the old coarse straw swirls. | `d8e2f74e2a36cad07a0dc9b251a5bd3bcbbd9dae4d73341af72a7fb5349ecd60` |
