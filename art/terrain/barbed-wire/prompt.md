---
id: terrain/barbed-wire
title: Barbed wire
kind: terrain
status: revised
runtime_ids: barbedWire
masters: barbed-wire-shared-v1.png, barbed-wire-belts-shared-v1.png
---

## Look
A contained patch of weathered wooden stakes and rusted wire coils on true transparency.

## Base prompt
Create ONE reusable WWI BARBED WIRE OBSTACLE terrain-overlay game asset, trueRGBAtransparentbackground. Strict orthographicTOPDOWN NADIR view withNO horizon orperspectivetilt; commonwarmupperleftsoftdaylight. Subject ONE small irregular patch of two looselyparallel rusted barbedwire coils strung between six short weatheredwoodstakes, with a few crossingwires, shortshadowpaintedonlywithinmaterial thickness. No people,weapons,vehicles orscene. Naturalpaintedfinematerialdetail, restraineddarkrustgreysteel and beigeweatheredwood, crispantialiasededges readableat96x80, scaleabout50metresperhextile soentirewirepatchabout25metreswide. This is for seamless reusablehexterrain: no ground orsurfacemat anywhere, fullalpha0betweenwires/loops/stakesandoutside; all wireends finishinsidepatchratherthantileborders. Entireobjectinsidecentral65%width AND central55%height, generousalphamargins onall4sides. Singlepatch, NOTmultiplepanels orviews. Canvas1536x1280landscape6:5ideal, preserving completepatchinsidecentred6:5crop. HandpaintedillustrationmatchinglushwarmWWI strategyterrain, naturaltexturesnotphotograph/pixelart/vector/3Drender. No paintedgrass,earth,castshadow,haze,glow,checkerboard,hexoutline,borders,text,labels,insignia orwatermark.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| contained obstacle | barbedWire | Six stakes, wire loops and transparent gaps; no edge ports. |

## Edges & joins
The obstacle ends inside the hex; all margins and gaps are transparent over the map's own ground.
It is an independent patch rather than a connected fence, reusable over summer and winter ground.

## Post

### Look 4 R6 barbed-wire candidate (2026-10-23)

`barbed-wire-belts-shared-v1.png` (`exec-23240f29`, native 1536×1024 RGBA, unchanged) holds three
plan-view entanglements in 512 px columns: a screw-picket belt, two concertina coils and two
knife-rest frames. The entry's 746 × 320 px patch cannot fit a 512 px column, so each patch stands
vertically (about 300 × 820 px, margins ≥ 92 px). Faults: about 10 % taller than asked, and every try
carried a faint brown backing (alpha 1–31) between the wires, covering 20–46 % of each patch, so cutting
needs an alpha threshold. Not selected. No attachments. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Transparent RGBA 1536x1024 sheet, strict nadir, three separate WWI barbed-wire entanglement patches in three 512x1024 columns, each about 70x30 world px (746x320 px): (1) a belt of twisted iron screw pickets seen from directly above as small dark round heads with short shadows, strung with criss-crossing rusty wire; (2) two rusted concertina coils lying side by side, loops seen from above; (3) two timber knife-rest (cheval-de-frise) frames with wire wrapped across them. All wire ends finish inside the patch. Rust brown, dark steel grey and weathered grey wood. Short soft shadows only within the wire thickness. Full alpha between wires and outside.
Avoid: side-view stakes or log ends, ground mats or grass, people, text, patches touching column borders.

Orientation (a 746 px long patch cannot fit a 512 px column, so): each patch lies VERTICALLY in its column, long axis top to bottom: about 320 px wide (62% of the column width) and 746 px tall (73% of the canvas height), centred in its column (column centres x=256, 768 and 1280, y=512), with at least 60 px of fully transparent margin to every column border and to the top and bottom edges. The three patches never touch or overlap.
Forms (as seen from directly above, no depth):
- (1) Screw pickets appear only as small dark round heads with tiny eye loops; no visible stems or stake sides (a previous attempt drew upright side-view stakes).
- (2) The concertina coils lie flat on the ground with their axes running top to bottom; seen from directly above, each coil shows as a long band about 140 px wide of many overlapping flattened wire loops along its length. Not round rings seen down the axis, not a 3D tube.
- (3) The knife-rests: each frame is one long timber spine running top to bottom with several pairs of crossed legs that appear only as thin straight X strokes across it, and barbed wire wrapped along the spine and leg tips. No round sawn log ends, no depth.
Corrections measured on three previous attempts:
- Every pixel is either wire, picket or timber, or fully transparent (alpha exactly 0); only 1-2 px of antialiasing at the object edges. No cast shadow, no ground shadow, no painted ground or dark backdrop behind the wires: all previous attempts carried a faint brown matte of alpha 1-31 over 18-50% of each patch area, plus specks outside.
- All previous patches came out 846-905 px tall, about 130 px too tall. Each patch is only 746 px tall (73% of the canvas height), from y=139 to y=885, with 139 px of empty transparency above and below, and 320 px wide (x=96-416, 608-928, 1120-1440).
```

`art-review -pack art/pipeline/overlay-barbed-wire-v1.json` preserves alpha through the centred 6:5 crop
and fixed Lanczos3 2×/4× exports (192×160 and 384×320). Runtime anchor (48,40).
Sapper placement and removal remain separate gameplay work.

## Avoid
- ground patches, opaque mattes, painted shadows or border-crossing wire ends
- text, insignia, perspective scenery or people
