---
id: terrain/fog
title: Unexplored
kind: terrain
status: approved
runtime_ids: unseen
masters: fog-unseen-shared-v1.png
---

## Look
Never-explored hexes: a calm, dark, information-free surface that still belongs to the painted map.

## Base prompt
Opaque seamless square texture, 1024x1024: a quiet dark desaturated blue-grey charcoal surface around
muted midnight slate (#252e3a), very subtle broad smoky pigment variation between deep slate and dark
indigo, restrained fine matte paper grain, spatially even illumination, very low local contrast. It must
contain no recognizable terrain or marks and stay calm when tiled.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| unseen | unseen | The texture above. |

## Edges & joins
The texture is periodic, so unseen hexes form one continuous dark field; the frontier to explored
hexes is drawn by the renderer.

## Post
`art-review -fog art/pipeline/fog-terrain-v1.json` builds the opaque unseen tile and seasonal frontier
review patches. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- stars, clouds or smoke plumes
- vignettes or hot spots
