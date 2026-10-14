# Original Amiga placement and adjacent terrain — 2026-10-15

The supplied Amiga program statically establishes a **16-pixel column step,
24-pixel row step and 12-pixel downward offset on odd columns** for its 24 × 24
terrain. A ground-only reconstruction now resolves original FIN indices through
`PARTS.DAT` / `PARTW.DAT` and places adjacent named resources. This is evidence for
G0 and subsequent seam work, not the remake's runtime geometry contract, a
populated game capture or approval of any redraw. No original executable was run.

## Evidence and coordinate system

The investigated `data/HL.DWN` has 92,219 packed bytes, SHA-256
`88628ff0ea6e318d9ab77f03737395b4f45cdfcd238aaff2dcc7ef853418eecb`.
Amiga big-endian TPWM unpacking consumes the entire input and produces 196,284 bytes,
SHA-256 `f6d925e6edaaeec5fa1ccc3ddb6e81b4d78d3405e627ec528c166229581245c6`.
It is flat 68k program data; the WHDLoad slave is a separate loader. All offsets
below refer to **unpacked HL.DWN**, not packed-file offsets or runtime addresses.

| Unpacked offset | Observed behavior | Consequence |
| --- | --- | --- |
| `0x8146–0x817e` | Loads raw FIN bytes 1 and 3 into internal dimension words at `a4+0x1ba4` and `a4+0x1ba6`; stores `a2+4` as the cell buffer; multiplies dimensions by two | FIN starts with two BE16 dimensions, columns then rows; cells have two bytes |
| `0x8dde–0x8e12` | Divides half the cell byte index by internal column count; remainder is column, quotient is row; repeats for viewport origin | Row-major FIN ordering |
| `0x8e7c–0x8eb6`, repeated `0x14fce–0x15008` | Subtracts viewport origin; tests local-column parity; multiplies local row by 24, adds 12 for odd local column; shifts local column left four | Tile origin `x=16*c`, `y=24*r+12*(c mod 2)` relative to the viewport |
| `0x8ebe`, `0x8ef4` | Adds 160 for the second viewport; pushes record pointer, y, x before drawing | Two viewport placement paths; not a full clipping/margin specification |
| `0x8bb2–0x8d52` | Independent full-view loop visits nine columns/seven rows, advances y by 24 and x by 16, starts alternate columns at y=12, uses `2*columns` cell row stride | Independent confirmation of placement and FIN stride |
| `0x25b78–0x25bba` | Draw entry loads x then y; aligns x to 16 and converts to planar byte offset; y indexes a row-offset table; reads graphic plane mask at byte 9 and data offset at word 10; advances destination rows by 40 bytes | Coordinate labels connect to actual native graphic drawing into 320-pixel planar rows |

Representative instructions are `BTST #0` at `0x8e8e` (`082e0000fff7`),
`MOVEQ #12` at `0x8e98` (`700c`), `MOVEQ #24; MULU.W; ADD.W` at `0x8ea4`
(`7018c0eefff4d16efffe`) and `MOVE.W; LSL.W #4; ADD.W` at `0x8eb0`
(`302efff6e948d440`). These anchors support manual inspection. A linear disassembly
also interprets embedded data as instructions; LLVM's m68k decoder leaves some
instructions unknown. Neither its listing nor these traced windows is a complete
decompilation.

For whole-map reference reconstruction choose origin `(0,0)`. The old hypothesis
`x=32*u+16*(v mod 2), y=12*v` yields the same infinite lattice, but different
logical cell labels. Do not apply row-stagger indexing to successive FIN cells.
The reference tool uses the original column-stagger indexing. Viewport-relative
parity and scrolling still need populated capture checks.

For an even column, screen-space neighbor displacements are `(0,-24)`,
`(16,-12)`, `(16,12)`, `(0,24)`, `(-16,12)`, `(-16,-12)`.
For an odd column the world displacements are the same, but neighboring logical
row indices differ. These are spatial directions; they do not assign the remake's
numeric road ports, compass letters or unit facings. The 384-pixel native mask
covers this lattice without overlapping covered pixels. Physical pixel aspect
and final display correction remain open; no flattening or rotation is applied.

## FIN resources and concrete adjacency

`LIB/PARTS.DAT` and `LIB/PARTW.DAT` unpack to identical 1,400-byte ordered tables of
175 eight-byte names. Their packed SHA-256 is
`ff1175a45d125c2b7975a62add34aee1d34c5f69c230bb4aa3af9030d4d195d9`.
This table supplies FIN terrain indices; archive FAT order differs and must not
be used as map order. The selected pilot IDs are river `36=SCOAS071`, bridge
`107=SSTRE091`, road `130=SSTRE114`; `174=VERBOTEN` is retained as original ground.
The two-byte cell starts at `4+2*(row*columns+column)`. The first byte selects
terrain. The exporter records the second byte raw and omits unit/state rendering.
Although the header reserves words, the traced loader reads their low bytes;
the exporter rejects dimensions above 255 rather than inventing a wider format.

A local structural audit of all 72 supplied `MAP/00.FIN`–`71.FIN` files found exact
`4+2*columns*rows` decoded sizes, dimensions 16–64, and all terrain indices below
175. This validates this supplied corpus, not arbitrary saves or other editions.

| Original map / cell `(column,row)` | Resource / index | Native tile origin | Adjacent source evidence |
| --- | --- | --- | --- |
| `00 / (1,15)` | `SSTRE114 / 130` | `(16,372)` | Upper-right `(2,15)=SSTRE122`; below `(1,16)=SBUIL014` |
| `00 / (2,2)` | `SCOAS071 / 36` | `(32,48)` | Lower-right `(3,2)=SCOAS075`; upper-left `(1,1)=SSTRE100` |
| `06 / (20,9)` | `SSTRE091 / 107` | `(320,216)` | Upper-left `(19,8)=SSTRE113`; lower-right `(21,9)=SCOAS063`; above `(20,8)=SCOAS067` |

These are actual neighboring map resources, including building and coast connections.
They are not a declaration that every visible material must continue across every
edge. Ground-only whole-map inspection shows coherent roads, banks, towns and
bridges; matching exits for the new illustrated studies still require a dedicated
comparison in these scenes.

## Independent manual cross-check

The grayscale JPEG on PDF page 23 / object 150 repeats masked ground texture at
raw top-left origins `(32,24)`→`(64,24)` (Pearson r=0.99561), `(64,24)`→`(64,48)`
(r=0.99536), and `(16,12)`→`(32,24)` (r=0.97866). Each comparison uses the 384
native covered cells, with mean absolute gray differences 0.471, 0.497 and 1.919
out of 255. This independently supports the same lattice in populated imagery.

Its frame matches decoded `GAME.IFF` at identical coordinates: r=0.97963 over
7,540 nonblack frame pixels above row 183. Winter page 41 / object 300 gives
r=0.97970. One-pixel shifts reduce agreement to about 0.38–0.56. Native columns
312–319 are entirely black, consistent with the manual's 312 × 200 screenshot
being a crop of the native 320 × 200 frame. The PDF places those pixels with
approximately equal x/y scaling. This does not establish original monitor aspect,
platform identity or RGB. Source and PNG hashes remain in
[manual-screen-references.csv](manual-screen-references.csv).

## Reproduce private map references

From the repository root, with fresh output directories:

```sh
just gfx-references \
  -data 'original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data' \
  -libraries PARTS -map 00 -out build/gfx-references/placement-map00
just gfx-references \
  -data 'original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data' \
  -libraries PARTS -map 06 -out build/gfx-references/placement-map06
just gfx-references \
  -data 'original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data' \
  -libraries PARTW -map 00 -map-library PARTW \
  -out build/gfx-references/placement-map00-winter
```

Open [map 00's local gallery](../../build/gfx-references/placement-map00/index.html#maps-00)
or [map 06's local gallery](../../build/gfx-references/placement-map06/index.html#maps-06).
`maps/00-PARTS.png` is 264 × 588, SHA-256
`6a9579c31ec54bef49e34bea535a5efc852bbca25af835a44d73acdc18268edd`;
`maps/06-PARTS.png` is 648 × 972, SHA-256
`5803bb09142c2dc27ac57fa50e12af4fe68645c10129e1fa2b2ec88d7d8bc4c1`.
Each matching `*-cells.csv` records original FIN offset, ground index/name, raw
second byte and native pixel origin. `source-sha256.txt` records packed input hashes.

The chosen library and palette are explicit: default `PARTS`, `00.PAL`. Switching
`-map-library PARTW` illustrates that ground treatment; it does not establish the
campaign season or its runtime palette selection. No units, UI, viewport clipping,
animations or state overlays are reconstructed. Original border markers remain.
References stay in ignored `build/`; none are installed into a playable pack.

## Reproduce the code-byte anchors

This standalone inspection recipe uses the documented Amiga TPWM format and
verifies both hashes before printing the traced windows. It writes nothing and
runs no native code. Python 3 standard library is sufficient:

```sh
python3 - <<'PYCODE'
from pathlib import Path
import hashlib
b = Path('original/extracted/amiga/HistorylineDe/fsuae/Hard Drives/data/HL.DWN').read_bytes()
assert hashlib.sha256(b).hexdigest() == '88628ff0ea6e318d9ab77f03737395b4f45cdfcd238aaff2dcc7ef853418eecb'
assert b[:4] == b'TPWM'
size = int.from_bytes(b[4:8], 'big')
out, pos = bytearray(), 8
while len(out) < size:
    flags = b[pos]; pos += 1
    for shift in range(7, -1, -1):
        if len(out) == size: break
        if flags & (1 << shift):
            a, c = b[pos:pos+2]; pos += 2
            distance = ((a & 240) << 4) | c
            assert 0 < distance <= len(out)
            for _ in range(min((a & 15) + 3, size - len(out))):
                out.append(out[-distance])
        else:
            out.append(b[pos]); pos += 1
assert pos == len(b) and len(out) == 196284
assert hashlib.sha256(out).hexdigest() == 'f6d925e6edaaeec5fa1ccc3ddb6e81b4d78d3405e627ec528c166229581245c6'
for start, length in [(0x8146,58), (0x8dde,58), (0x8e7c,136), (0x25b78,52)]:
    print(hex(start), out[start:start+length].hex(' '))
PYCODE
```

## Next graphics batch

Use these named neighborhoods for deterministic original/redraw comparisons at
native/2×/4× density, including both column parities and three-tile vertices.
Check the material contact cells recorded in [connected-terrain-pilot.md](connected-terrain-pilot.md)
against actual neighbors before assigning runtime ports or authoring shared edge
profiles. Compare source texture boundaries as well as road/water exits: the original
already has intentional material discontinuities. Keep physical aspect, scrolling,
unit anchors/facings and layered overlay migration open. G0's full pilot and user
visual approval still precede G1's shared runtime compositor and bulk production.
