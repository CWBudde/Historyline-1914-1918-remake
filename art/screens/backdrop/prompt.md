---
id: screens/backdrop
title: Menu backdrop
kind: screen
status: candidate
runtime_ids: backdrop/western-front
masters: western-front.png
---

## Look
Painted backdrop behind every menu screen: a quiet Western Front landscape after rain, dark and calm on
the left where the menu panels sit, with interest to the right and lower third.

## Base prompt
Text-free title and menu backdrop, one continuous painted illustration, wide 16:9, ideally 2560x1440.
Scene: Western Front countryside after rain, seen from a low rise: grassy earth and a timber-lined trench
in the lower foreground, a winding narrow road leading to a small distant village, low wooded hills fading
into mist. No fighting or casualties. Quiet, dark, cool sky and low-detail terrain on the left 45 percent;
horizon below the middle. Palette: cool slate #162630, muted blue-grey, moss and earth, restrained warm
brass #c4a773 highlights and cream mist. Opaque.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| western-front | backdrop/western-front | The scene above. |

## Post
`go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>` crops the master by the
manifest rectangle and reduces it with Lanczos3 to 2x (1280x800) and 1x (640x400). Promote into
`assets/screens/backdrop/` and re-index.

## Avoid
- words, logos, maps with labels or UI
- modern equipment
- bright detail in the left menu area
