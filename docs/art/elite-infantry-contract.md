# Elite-infantry sprite and portrait candidates

2026-10-15. [Review gallery](elite-infantry-review.html) ·
[Sprite manifest](../../art/unit-elite-infantry-v1.json) ·
[Portrait manifest](../../art/portrait-elite-infantry-v1.json).

`CP_eliteInf` and `EC_eliteInf` now have separate elite-infantry offline
bindings. This family is the reference's two-person machine-gun team, distinct
from the ordinary three-rifleman infantry. Both countries share the artwork;
the current owner independently selects CP, EC or neutral tactical colours.
The portrait is owner-neutral.

The continuing user brief requests the established painted treatment and about
90% resemblance at reference size after Lanczos reduction. The repository
portrait and icon supplied composition/identity; the existing authored infantry
art supplied painted style, camera and atlas conventions. The exact prompts and
input hashes are retained in the
[atlas recipe](../../art/recipes/units/reference-v1/elite-infantry-atlas-01.json)
and [portrait recipe](../../art/recipes/portraits/reference-v1/elite-infantry-01.json).
Both masters are unchanged built-in image generation outputs, not original-pixel
composites. No private installation bytes were copied.

The portrait keeps the sepia sky/field, horizontal left-facing machine gun,
assistant behind the receiver and crouched gunner on the right. The old UI
bevel is omitted. The 1326×1187 opaque master uses the exact existing 105:94
crop: 1260×1128 at (33,29). The unchanged premultiplied byte-RGB Lanczos3 chain
produces 840×752, then 420×376, then 210×188. The illustration has more detailed
clouds, grass and equipment than the reference. Approximate90% is a visual
target, not a computed score or final user approval. This is the repository's
210×188 reference, not a verified native Amiga96×77 illustration.

The transparent 1024×1536 tactical atlas has six 512-square authored panels,
row-major N,NW,SW,S,SE,NE. It reuses the existing fixed0.8 whole-panel framing,
96×96 logical canvas, (48,49) anchor and terrain-relative (0,-9) offset.
No finished art is rotated, mirrored or independently recropped. Magenta cloth
is technical input for the unchanged neutral/tint-energy separation. Independent
Lanczos3 plane reduction supplies 2×/4× CP/EC/neutral normal/spent sprites.
Spent state uses red-channel grayscale after tint. All source alpha is retained;
the sprite guard has at least9.5 world pixels clear, exceeding the required6.
The tiny repository icon supplies identity, not six verified native directions
or a validated physical pivot. Weapon/crew formation varies slightly by view.

The retained sprite bundle contains72 sprites,24 base/tint planes,4 sheets and
24 seven-cell terrain fixtures:124 candidate PNGs. A fresh build has134 files
including unchanged input snapshots. The portrait adds3 exports in a7-file
build. Every output pin and geometry record is retained in the
[sprite export recipe](../../art/recipes/unit-elite-infantry-v1.json) and
[portrait export recipe](../../art/recipes/portrait-elite-infantry-v1.json).
The gallery uses continuous image rendering and compares the portrait at logical
size with the repository reference.

```sh
go run ./tools/art-review -unit art/unit-elite-infantry-v1.json -out <fresh-unit-directory>
go run ./tools/art-review -portrait art/portrait-elite-infantry-v1.json -out <fresh-portrait-directory>
```

Verification: both complete bundles reproduce byte-for-byte across two fresh
builds (134+7 files). The focused unit/portrait/path suite passes, including
three-family actual saved-PNG bundles, palette/state reconstruction, decoded
sheet/scene assembly, registration, scalar Lanczos oracle, strict schema,
source pins, opacity, snapshots and elite portrait/class bindings. Vet and lint
pass. Source art, production sheets, logical portrait and grass/mud fixtures
were visually inspected: two crew and one gun are readable in all six headings,
with clean cutouts and independent owner colours.

Runtime integration, motion, exact physical contact, native A–F correspondence,
Blender comparison, full-map/fractional-zoom review and final visual acceptance
remain open. The existing Windows replacement-directory rollback failures in
Phase2.1 are excluded from this focused verification and remain open; this is
not a full-suite or `just ci` pass.
