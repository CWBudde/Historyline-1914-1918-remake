---
id: screens/backdrop
title: Menu backdrop
kind: screen
status: candidate
runtime_ids: backdrop/western-front
masters: western-front.png, campaign.png, intermission.png, production-repair.png, editor.png
---

## Look
Painted Western Front scenes for menus and feature illustrations. Cool slate, moss and earth, quiet on
the left for live panels, with the main subjects to the right and in the lower third. Only the shared
Western Front backdrop has a runtime binding; the four feature paintings are candidate sources.

## Base prompt
Text-free World War I strategy-game backdrop, one continuous hand-painted digital illustration, opaque,
16:10, target native 2560x1600 for 4x at 640x400 logical. Rural Europe, 1914–1918, with believable wood,
canvas, steel, grass and earth, connected painterly detail and soft diffuse light from the upper left.
Palette: cool slate #162630, muted blue-grey, moss and earth, restrained warm brass #c4a773 highlights
and cream mist. Keep the left 40 percent dark, calm and low-detail for live panels; concentrate subjects
on the right and lower third, with the horizon below the middle. Quiet activity, no fighting or
casualties. No text, controls, frames, symbols, national insignia or modern equipment.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| western-front | backdrop/western-front | Countryside after rain, seen from a low rise: grassy earth and a timber-lined trench in the foreground, a winding narrow road to a small distant village, low wooded hills fading into mist. |
| campaign | — | Quiet command tent, with a field table on the right holding an unlabelled geographic map, brass dividers, a compass, binoculars and rolled paper. Canvas, timber and soft daylight; keep the left 40 percent dark and low-detail. No people, readable writing or route markings. |
| intermission | — | Old stone farmhouse and open barn on the right, with a horse-drawn supply wagon beneath a tree, crates, folded canvas and a wet lane. A restrained warm window glows against the cool countryside; keep the left 40 percent dark and quiet. No people. |
| production-repair | — | Barn workshop interior, with a World War I field gun on the right: wooden-spoked wheels, a long barrel and protective shield. A workbench, tool chest, spanners and spare wheel surround it; an open door looks into the countryside. Keep the left 40 percent dark and quiet. No people. |
| editor | — | Mapmaker's survey table beneath an open canvas shelter on the right: an almost blank paper sheet with faint unlabelled contours, brass dividers, a straightedge, compass and rolled maps. A brass survey instrument on a tripod overlooks fields, a stream, a village and wooded hills. Keep the left 40 percent dark and quiet. No people, grid or interface. |

## Post
`go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>` crops the master by the
manifest rectangle and reduces it with Lanczos3 to 2x (1280x800) and 1x (640x400). Promote into
`assets/screens/backdrop/` and re-index.
It also exports 4x (2560x1600) when the selected crop contains at least that many native pixels.

The other four masters are candidate paintings without runtime IDs or manifest bindings. Their current
1586x992 source size supports 2x; the native 2560x1600 target is not met. Exact-size API generation is
pending. Do not upscale them and label the result native 4x.

## Avoid
- words, logos, maps with labels or UI
- modern equipment
- bright detail in the left menu area
