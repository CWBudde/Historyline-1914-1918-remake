# Private native Amiga reference exporter

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
`metadata.csv`, and `source-sha256.txt`. The CSV contains archive offsets,
stored/unpacked lengths, dimensions, plane mask, data offset, alpha provenance,
and source/PNG SHA-256 hashes. Archive offsets refer to the unpacked library;
source hashes refer to the original packed bytes. Resource names are identifiers,
not semantic classifications: names beginning `SFORE` can contain cabins.

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
**DOS files are unsupported and must not be decoded with this tool.**

`HEXMSK00.DAT` and `HEXMSK20.DAT` must unpack to complementary 96-byte planes.
The supported hex mask spans 24×24 pixels, with four stored bytes per row and
384 covered pixels. Row widths are:

```text
8 10 10 12 14 16 16 18 20 22 22 24 24 22 22 20 18 16 16 14 12 10 10 8
```

Terrain alpha comes exclusively from `HEXMSK00`; unit/portrait alpha comes
from its embedded plane. Black pixels inside coverage stay opaque. No pixel
aspect correction, resizing, map layout, or 24×20 flattening is inferred.

`STATS`, `CURSOR`, and standard FORM ILBM files such as `GAME.IFF` are outside
this exporter's supported record format. Separate forensic decoding established
that GAME.IFF is an empty split-view frame with 320×200 intrinsic dimensions,
five planes, ByteRun1 compression, and a BMHD aspect field of 10:11. Its CMAP
high nibbles match all 32 words of `02.PAL`; this does not establish runtime
terrain display geometry.

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
Input/unpacked files are capped at 32 MiB, individual images at four megapixels,
libraries at 10,000 resources, and a complete export at 16 megapixels. These
limits comfortably exceed the verified native reference libraries.
