# art/recipes/units/reference-v2/charron-wide.json

Recipe: `art/recipes/units/reference-v2/charron-wide.json`

- **acceptance**: Authored candidate correction only. Original-size likeness, mechanical/facing consistency, native anchors and runtime approval remain open. No computed similarity percentage.
- **art_direction**: reference-faithful-painted-proportions-v3
- **date**: 2026-10-16
- **schema_version**: 1
- **source**: art/source/units/reference-v2/charron-wide.png
- **source_sha256**: c5f5f7730921fa5e91fa3b084bdebaa164faf34f587db29535beb8b55301fca3
- **status**: selected candidate
- **tool**: built-in image_gen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/units/reference-v2/charron.png` | edit target: previously authored six-view atlas | `541551cb38a492d7415a5599952fbfc4cee9e8c4bea5c96dd98f89cb80969020` |
| `legacy/python/gfx/units/icons/charron.png` | unchanged original proportions and overhead camera reference only; not installed | `2d99afb97654cf923487b33d42ee05de88dea78aab2f3af6e60f212a3f2b7833` |
| `art/source/units/reference-v2/facing-guide.png` | supporting six-direction layout guide only; not installed | `67eb2dcfc13430889b6ef4c35210a9f3049dc91c33f5ef724d4d76a77a020f90` |

## Prompt

```text
Use case: precise-object-edit. Edit image1, our six-view Charron atlas. Image2 is ONLY the original unit proportion/camera reference; image3 is ONLY a direction layout guide. Keep exact1024x1536 transparent canvas, 2columns and3rows of512squarecells, six complete independent overhead views. Row-major vehicle directions N, NW, SW, S, SE, NE. Current vehicle is too narrow: broaden the ENTIRE HULL and WHEELBASE about25percent while preserving current full LENGTH, small round steel turret, four wheels, all fittings, current upper-left lighting and detailed painted style. UP-panel target overall visible bounds x132..380 y52..470, centered256,261. The other five views same physical size, all within their own local x/y30..482. A long rectangular WWI armoured car seen from high overhead, tiny sidewall visibility; broad flat top with modest turret. Technical saturated magenta ownership recolour on armour, steel/tan/brass turret and wheel rims, dark restrained outline. Consistent steel highlights, no shiny green vehicle. All six views retain directions and all details from supplied atlas, no invented accessories. Every pixel outside equipment genuinely alpha0: no floor, shadow, background haze, scenery, text, grid or direction marks. Preserve transparent margins in EACH512square panel, no object crossing a panel boundary.
```
