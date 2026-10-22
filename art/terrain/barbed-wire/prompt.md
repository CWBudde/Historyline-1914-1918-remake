---
id: terrain/barbed-wire
title: Barbed wire
kind: terrain
status: revised
runtime_ids: barbedWire
masters: barbed-wire-shared-v1.png
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
`art-review -pack art/pipeline/overlay-barbed-wire-v1.json` preserves alpha through the centred 6:5 crop
and fixed Lanczos3 2×/4× exports (192×160 and 384×320). Runtime anchor (48,40).
Sapper placement and removal remain separate gameplay work.

## Avoid
- ground patches, opaque mattes, painted shadows or border-crossing wire ends
- text, insignia, perspective scenery or people
