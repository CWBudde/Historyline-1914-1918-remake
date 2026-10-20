# DOS-inspired shore refresh

2026-10-20. [All nine variants](shoreline-review.html) ·
[PULSE comparison](../../build/gfx-review/pulse.html) ·
[Shore manifest](../../art/shore-terrain-v1.json).

All nine summer lakesides now use a rough pale limestone bank, a darker wet
lip and a blue-green submerged shelf. The open water has a muted blue-teal
base and broken light ripples, following the original DOS lake rather than
the previous narrow brown outline and muted grey water.

The built-in image generator painted two new 1254×1254 material masters from
the original DOS ground reconstruction and individual DOS water/shore tiles.
The original images remain in the ignored local reference export. The new
[water recipe](../../art/recipes/terrain/shore-reference-v2/water.json) and
[bank recipe](../../art/recipes/terrain/shore-reference-v2/bank.json) retain
the complete prompts, input roles, hashes and output dimensions.

The selected water master averages RGB (48.7, 94.1, 93.1), close to the DOS
open-water tile's (53.2, 98.3, 100.0). An earlier bright-cyan candidate was
discarded. These means check the palette; they do not measure visual fidelity.

The Go shore renderer applies these materials consistently to `lakeside12`,
`lakeside16`, `lakeside23`, `lakeside34`, `lakeside56`, `lakeside123`,
`lakeside126`, `lakeside156` and `lakeside3456`. It exports 192×160 (2×) and
384×320 (4×) tiles. The 2× outputs are Lanczos reductions of the overscanned
4× render, with the final hex ownership mask applied after filtering.

The bank outline has small irregularities from lattice-periodic variation and
the stone material. Shared crossing positions, the established curved coast
geometry and all land/water corner profiles remain consistent, including
`lakeside23`'s effective 234 mask. This is a DOS-inspired redraw, not a traced
reconstruction of each original shore outline. `lakeside12` remains
family-derived because it is absent from the current campaign maps.

Open water, rocky-water ground and connected river materials use the same
new water master and sampling phase. Both-parity shore patches, rock contexts
and the PULSE maps and game captures are regenerated from the exported assets.
Existing grass and rock artwork remain the basis for their respective regions.

The regression test checks that the stony bank remains readable inland at
all eighteen mixed-edge crossings across the nine variants. The existing
shore tests also check reciprocal profiles, filtered PNG guards, masks,
deterministic exports and actual PNG patch assembly. Visual acceptance of
the new appearance remains open for review in the linked galleries.

To rebuild the shore bundle independently:

```sh
go run ./tools/art-review -shore art/shore-terrain-v1.json -out <fresh-output-directory>
```

To refresh the maps and nine actual game captures after building the assets:

```sh
just gfx-study
```
