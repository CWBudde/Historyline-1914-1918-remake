# meadow

Recipe: `art/recipes/terrain/distinct-v1/meadow.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source preserved unchanged. Exact center crop then Lanczos3 exports, no original mask or pixels. Rectangular overscan; runtime hex clipping and connections not yet accepted.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: meadow
- **review**: Independent batch review recorded in docs/art/distinct-terrain-batch-01.md. No original-fidelity acceptance sought under updated brief.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/meadow.png
- **source_sha256**: 8c8935efbbd463e39d49208e42ca0e87b140268a1b11941764dfa7c29c4fff85
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## Prompt

```text
Create ONE production game terrain illustration, not a sprite sheet. Original new artwork for an early twentieth-century European strategy game, a modern continuous finely painted naturalistic illustration seen STRICTLY vertically from directly above, orthographic, no horizon and no perspective. Landscape canvas aspect ratio 6:5. Subject: an uninterrupted muted sage and moss-green meadow, fine short grasses, tiny sparse clover leaves, subtle earth glimpses. Tiny organic detail, calm low-frequency irregular variation, no central focal object, no paths, no trees, no flowers or rocks. Consistent flat diffuse overcast light, neutral gentle colors, subtle short upper-left light, no dramatic shadows. Fill all four edges with the same material, no dark framing or vignette, approximately even edge values. It should be visually quiet under game units at 192x160, yet show crisp flowing blades at large source resolution. Broad natural connected shapes and continuous contours, no pixel art, no tracing of square pixels, no outlines, no hex shape, no border, no UI, no text, no watermark. Entire opaque image, substantial overscan for later clipping. Invent the illustration from scratch; do not emulate an existing game's art or palette.
```
