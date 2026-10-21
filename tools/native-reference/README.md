# Private original graphics reference exporter

Amiga remains the default mode. Use explicit `-platform dos` for original DOS
ground; the DOS and Amiga formats are decoded separately.

## Original DOS ground

```sh
go run ./tools/native-reference -platform dos \
  -data original/extracted/dos-full -libraries PARTS -map 00 \
  -out build/dos-reference
```

DOS mode requires `-map` and exactly its selected `-map-library` in `-libraries`
(`PARTS` by default; `PARTW` for explicit winter inspection). It exports the 175
original terrain tiles, a native ground-lattice reconstruction, a version aligned
to the remake geometry, an offline gallery, resource/cell metadata and input/image
SHA-256 pins in `reference.json`. No Python art or separately redrawn road/path
overlays are used. Packed unit sprites, state/ownership, runtime palette changes,
screens and briefings are outside this ground-only mode.

DOS TPWM sizes and the library FAT are little-endian. FIN dimension words remain
big-endian. Palettes unpack to 256 RGB byte triples and are preserved without
expansion. The supported ground record has an 18-byte header, `INFOILBM` signature,
colour-key byte at 8, encoding byte `0x55` at 9, and LE16 width/height at 14/16.
The payload contains four x-mod-4 VGA lanes: byte index
`18 + (x%4)*(width/4)*height + y*(width/4) + x/4` selects the palette colour.
Transparency uses each record's colour key, preserving opaque black and every
source edge pixel. The DOS unit `0x50` encoding is rejected.

The native reconstruction uses the terrain mask's 24×24 lattice at
`x=16*column, y=24*row+12*(column%2)`. This static reconstruction is not an executed
game capture or physical-display verification. The aligned view uses
`line=2*row+column%2, col=column/2` and nearest-neighbour rescales the complete
lattice uniformly in the original 24×20 hex aspect (×4 horizontally, ×10/3
vertically). The art is not stretched; cell centres drift from the remake's (column
pitch 64 vs 71 px), so wipe/difference against the remake is approximate. Scaling the
joined lattice preserves mask adjacency; original hex outlines still differ from the
remake's. The native PNG keeps its stored pixels.
The final original column and unused odd half-row lie outside the remake grid;
other original forbidden boundary tiles are retained. The comparison canvas clips
the outer edge. General aligned exports
require an even FIN width and stay within the image budgets.

The same fresh-output/private-directory restrictions below apply to DOS mode.

## Amiga mode

Run from the repository with your locally supplied **Amiga** data directory:

```sh
go run ./tools/native-reference \
  -data '/path/to/Amiga/Hard Drives/data' \
  -out build/native-reference
```

`-out` defaults to `build/native-reference`; it must be a **new** subdirectory
inside this repository's ignored `build/` or `original/` tree. Existing output
directories are rejected, and symlinks cannot redirect outputs outside those
private trees. Input and output directories must not overlap. Run again with
another private output directory. The exporter never modifies its input.

Default libraries: `PARTS,PARTW,UNIT,BIGUNIT`. Use `-libraries PARTS` for a
terrain-only export. Optional native graphic libraries: `EXP,UMAPDASA,UMAPDAS2,
UMAPDAWA,UMAPDAW2`. `-palette` selects `00` (default), `01`, or `02`. Each native
resource produces an intrinsic-size PNG under its library name. The exporter
also writes `HEXMSK00.png`, palette swatches in increasing index order,
`metadata.csv`, `source-sha256.txt`, and an offline `index.html` gallery.
The CSV contains archive offsets,
stored/unpacked lengths, dimensions, plane mask, data offset, alpha provenance,
and source/PNG SHA-256 hashes. Archive offsets refer to the unpacked library;
source hashes refer to the original packed bytes. Resource names are identifiers,
not semantic classifications: names beginning `SFORE` can contain cabins.

## Optional original screens and campaign images

Screen decoding is opt-in; default library exports keep their paths and CSV
schema. Add `-screen GAME.IFF` to decode that filename from the data folder.
Add `-briefings` to export exactly the 46 campaign image files
`TXT/H00.LBM` through `TXT/H45.LBM`; the tool does not scan for other files.
Either option can be used independently, or together:

```sh
go run ./tools/native-reference \
  -data '/path/to/Amiga/Hard Drives/data' \
  -screen GAME.IFF -briefings \
  -out build/native-reference-screens
```

The additional PNGs are `screens/GAME.IFF.png` and
`screens/TXT-H00.LBM.png` through `screens/TXT-H45.LBM.png`. These are
intrinsic-size references, not populated game captures or runtime assets.
Each screen uses its own file CMAP, independent of `-palette`.
`screen-metadata.csv` records original relative source names, platform,
format, dimensions, word-padded row bytes, planes, compression, coverage,
transparent-color index, CMAP count, BMHD aspect/page fields, CAMG,
stored BODY/decoded planar lengths, and source/PNG SHA-256 hashes.
Source hashes refer to original file bytes, including an outer TPWM wrapper
when present. `source-sha256.txt` also includes every selected screen source.
All selected inputs, including the final briefing image and output filename
collisions, are validated before any output directory is created.

## Offline gallery

Open the exported `index.html` directly in a browser. No server, network,
libraries, or fetched metadata are needed. Search by resource name/library,
filter a library (including `screens` and `briefings` when exported), and
choose integer inspection magnification of 1×, 2× (default), 4×, or 8×.
Checkerboard, dark and light backgrounds help inspect coverage. Magnification
uses nearest pixel rendering and adds no authored detail. Original intrinsic
dimensions, coverage source, palette and platform stay visible beside file
and metadata links. Resource permalinks have stable fragments such as
`#PARTS-SFORE150`, `#screens-GAME.IFF`, and `#briefings-TXT-H15.LBM`.
Reference values are escaped through Go's HTML templates.

The gallery does not assign semantic mappings, infer world spacing, apply
aspect correction, or close G0's 90/100 native-resemblance gate. The 2×/4×
inspection options do not constitute newly authored higher-resolution art.

## Optional original ground-map reconstruction

Add `-map 00` through `-map 71` to export one supplied Amiga FIN ground map.
Select its ground library explicitly with `-map-library PARTS` (default) or
`-map-library PARTW`; it must also appear in `-libraries`:

```sh
go run ./tools/native-reference \
  -data '/path/to/Amiga/Hard Drives/data' -libraries PARTS \
  -map 00 -out build/native-map00
```

The exporter resolves FIN indices through `LIB/PARTS.DAT` or `LIB/PARTW.DAT`,
never archive FAT order. It places masked 24×24 ground at
`x=16*column`, `y=24*row+12*(column mod 2)`, using the original Amiga placement
traced in [original-placement.md](../../docs/art/original-placement.md).
`maps/00-PARTS.png` and `maps/00-PARTS-cells.csv` provide a ground-only scene and
each cell's FIN offset, original index/name, raw second byte and native position.
The gallery links both; the source manifest includes the FIN and index table.
All selected inputs are validated before output creation, with a four-megapixel
map limit and the existing combined 16-megapixel limit.

This is a **static reconstruction**, not a game capture. The selected palette
(`-palette 00` by default) and summer/winter library are explicit inspection
choices, not inferred campaign state. Units, UI, clipping, animations and state
overlays are omitted; original ground border markers remain. No physical pixel
aspect correction is applied. This utility does not establish the remake's
runtime geometry, road-port numbering or unit anchors.

**Keep source archives, unpacked data, and exported art private and ignored.**
This tool contains only code and synthetic tests. It is an analysis utility,
not a redistribution license, asset import pipeline, or fidelity score.

## Verified native layout

Amiga TPWM starts with `TPWM`, followed at byte 4 by a **big-endian** unpacked
length. Starting at byte 8, each flag byte describes up to eight tokens from
most-significant bit to least. A zero bit means one literal byte. A one bit
means two bytes `a,b`, with distance `((a & 0xf0) << 4) | b` and count
`(a & 0x0f) + 3`. Backreferences copy forward and can overlap. The final copy is
clipped to the declared output size. Individually compressed resources can
have one arbitrary alignment byte when the consumed stream length is odd.
Whole-library streams must consume their complete input.

After whole-library decompression, byte 0 contains the big-endian 32-bit FAT
offset. Each 12-byte FAT entry contains an eight-byte name followed by a
big-endian 32-bit resource offset. `BIGUNIT` packs its resources individually.
Records use the following custom header; this is **not** standard FORM ILBM:

| Record offset | Encoding | Meaning |
| --- | --- | --- |
| 0 | eight bytes | `ILBM    ` |
| 8 | BE16 | stored-plane bitmask, bits 0–4 |
| 10 | BE16 | pixel-data offset |
| 12–17 | three words | unresolved; ignored |
| 18 | BE16 | 16-bit words per row |
| 20 | BE16 | width in pixels |
| 22 | BE16 | height in pixels |

Pixels are conventional Amiga bits, MSB first within each byte, with each row
padded to a 16-bit boundary. Complete planes are stored consecutively in
increasing bit significance, including only the planes selected by the mask.
An omitted plane contributes zero. Pixel data starts at byte 24 for terrain;
a larger offset must equal `24 + rowBytes*height` and identifies one embedded
coverage plane preceding the pixels. Every record must match its expected
unpacked length exactly.

Palette files are **raw 64-byte** arrays of 32 big-endian `0RGB` words. Each
four-bit component is scaled to the full eight-bit range by multiplication by
17. DOS TPWM lengths are little-endian, and DOS palette/graphic records differ;
**use explicit `-platform dos` for DOS ground; the default Amiga decoder rejects it.**

`HEXMSK00.DAT` and `HEXMSK20.DAT` must unpack to complementary 96-byte planes.
The supported hex mask spans 24×24 pixels, with four stored bytes per row and
384 covered pixels. Row widths are:

```text
8 10 10 12 14 16 16 18 20 22 22 24 24 22 22 20 18 16 16 14 12 10 10 8
```

Terrain alpha comes exclusively from `HEXMSK00`; unit/portrait alpha comes
from its embedded plane. Black pixels inside coverage stay opaque. No pixel
aspect correction, resizing, map layout, or 24×20 flattening is inferred.

`STATS` and `CURSOR` remain outside the supported custom record format.

## Standard FORM/ILBM decoding

Optional screens use the separate standard IFF decoder, after bounded Amiga
TPWM decompression if present. FORM and chunk lengths/padding must match;
duplicate required/property chunks and properties after BODY are rejected.
Indexed images with 1–8 planes, uncompressed or ByteRun1 rows, and a CMAP
of up to 256 RGB triples are supported. Each scan line stores plane 0 (the
low index bit) through the highest plane, followed by an optional mask plane.
Rows are MSB-first and word-padded. ByteRun1 packets must fit a single plane
row; literals, repeats and signed -128 no-ops are handled without crossing
rows. A mask bit of zero is transparent; BMHD transparent-color masking
compares the source index. Otherwise index zero and black remain opaque.
HAM/EHB modes, deep RGB, lasso masking, DEST plane merging, unsupported
compression, missing/out-of-range palette values and trailing BODY data
are rejected. Color cycling and other ancillary chunks are not animated.

GAME.IFF is an empty split-view frame with 320×200 intrinsic dimensions,
five planes, ByteRun1 compression and BMHD aspect 10:11. The 46 supplied
campaign images are also 320×200, five planes and ByteRun1, with aspect 1:1.
Aspect metadata is recorded without resizing or establishing terrain spacing.
Screen CMAP channels are preserved byte-for-byte, including low nibbles;
they are not expanded from hardware 0RGB words like external library palettes.
The earlier forensic GAME PNG expanded its CMAP high nibbles by 17; its
colours therefore differ from the CMAP-preserving export (up to 15 per channel),
while geometry and coverage match. The CMAP high nibbles still match `02.PAL`.

Format cross-checks: [Ancient's TPWM implementation](https://github.com/temisu/ancient/blob/master/src/TPWMDecompressor.cpp),
[Electronic Arts/Commodore ILBM specification](https://wiki.amigaos.net/wiki/ILBM_IFF_Interleaved_Bitmap).
The custom record offsets, sparse-plane order, coverage planes, and hex mask
were verified against the locally supplied native files rather than assumed
from the DOS format.

## Validation

```sh
go test ./tools/native-reference
go test -race ./tools/native-reference
go test ./tools/native-reference -fuzz FuzzNativeFormats -fuzztime=5s
```

Tests construct small synthetic inputs in memory or temporary directories.
They cover literal and overlapping TPWM copies, malformed bounds, endian
rejection, native palette conversion, sparse/five-plane images, word padding,
embedded/external masks, black-shadow preservation, archive offsets and names,
resource-count limits, output confinement, source hashes, and PNG/CSV export.
Screen tests additionally cover plane interleaving, ByteRun1 row boundaries,
mask/transparent-color semantics, CMAP preservation, aspect metadata,
outer TPWM, malformed chunks, unsupported modes, fixed campaign selection,
late-input/collision rejection before writes, aggregate pixel limits,
combined metadata/hashes, default compatibility, and gallery escaping.
Input/unpacked files are capped at 32 MiB, individual images at four megapixels,
libraries at 10,000 resources, and a complete export at 16 megapixels. These
limits comfortably exceed the verified native reference libraries.
