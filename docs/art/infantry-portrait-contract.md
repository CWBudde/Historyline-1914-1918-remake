# Infantry portrait crop and panel contract

2026-10-15. [Gallery](infantry-portrait-review.html) ·
[Manifest](../../art/portrait-infantry-v1.json) ·
[Recipe and output pins](../../art/recipes/portrait-infantry-v1.json).

`CP_infantry` and `EC_infantry` share one new owner-neutral opaque illustration.
This follows the actual Python picture aliases at `hlrData.py:238` and `:253`.
The reference panel's picture is 210×188 logical pixels; it is separate from the
96×96 map sprite, faction recolouring, moved-state grayscale and ownership flags.
Those controls and all frame/text/icons remain independent UI layers.

The new art shows three anonymous infantrymen in neutral grey/sage uniforms,
rounded steel helmets, leather and wood, against a quiet painted field. It is
not a national uniform reconstruction. Initial source 01's photographic skin,
wool and background were an Important style-review finding. The built-in image
tool repainted that newly authored source into visible painted forms while
retaining the composition, anatomy and equipment. Source 01 remains a rejected
study; source 02 is the export input. Both generated files remain byte-identical
to their tool outputs. No original game image or close reconstruction was input.

The selected opaque source is **1254×1254**, despite the initial prompt requesting
1024 square. The exporter derives the exact crop from actual decoded dimensions:

```
k = min(width/105, height/94)       # integer division; require k >= 8
crop = (105*k, 94*k)
offset = floor((sourceSize-crop)/2)
```

Here `k=11`, the crop is **1155×1034** at **(49,110)**. Odd extra pixels fall on
the right/bottom. Copy that exact rectangle without resampling, then apply the
unchanged alpha-aware byte-RGB Lanczos3 chain:

| Stage | Dimensions | Meaning |
| --- | --- | --- |
| 4× | 840×752 | Crop reduced once; no enlargement or aspect stretch |
| 2× | 420×376 | Whole 4× PNG image reduced |
| 1× | 210×188 | Whole 2× image reduced; logical-size review |

These are three density versions of one illustration, not faction or spent-state
variants. Filtering uses signed Lanczos lobes, clipped/renormalized edge support,
premultiplied byte RGB and half-up output rounding. There is no linear-light
conversion, nearest-neighbor upsampling, independent object crop or alpha edit.
The complete source, including discarded overscan, must be fully opaque. Original
16-bit alpha is checked before conversion can round faint transparency to 255.
All three exported PNGs remain fully opaque.

The strict `-portrait` manifest fixes schema 1 and family `infantry`, with a safe
source path and canonical SHA256. Unsupported families/aliases, unknown or null
fields, duplicate keys, wrong types, unsafe paths and mismatched pins reject.
Actual exports require `-out` and exclude other explicit flags/arguments. Global
`-h`/`-help` remains read-only help and creates no export. Sources are read once;
the exact original bytes and manifest are snapshotted. Existing output paths are
preserved; failed writes roll back only the directory the exporter created.

The fresh bundle has **seven files**: unchanged source snapshot, three exported
PNGs, exact manifest snapshot, report and HTML index. The report binds both class
IDs and records source dimensions/hash, crop, filtering chain and output pins.

Independent art review checks all three saved PNGs at their actual resolution.
All three heads/helmets, rifle tips and the complete visible hand survive the
crop; the 210×188 view remains legible. The portrait does not bake text, flags,
frame or other UI. It is ready for panel composition as a provisional candidate.

This does not establish the native Amiga 96×77 illustration binding, physical
display aspect, populated panel layout, runtime replacement or whole-UI acceptance.
The original installation is not a generation/export dependency. Remaining unit
portraits, weapons, panel art and runtime integration remain separate work.
