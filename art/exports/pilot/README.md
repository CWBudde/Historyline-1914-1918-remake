# Unapproved pilot exports

These are deterministic reductions of retained AI studies, not an accepted runtime pack.
Original reference pixels and comparison sheets remain in ignored `original/` and `build/`.
Do not mix these export families by their filename or density suffix.

| Family | Reference | 2× | 4× |
|---|---|---|---|
| `repo-2x`, `repo-4x` grass/forest/water | Repository 96 × 80 sprites | 192 × 160 | 384 × 320 |
| `amiga-native-2x`, `amiga-native-4x` forest, road, river, bridge | Amiga `SFORE150`, `SSTRE114`, `SCOAS071`, `SSTRE091`, 24 × 24 | 48 × 48 | 96 × 96 |

See `art/recipes/` for exact prompts, inputs and hashes, and `docs/art/style-guide.md` for
the visual criteria. The generated images are retained under `art/source/pilot/` so exports
do not need another image-generation call.

Reproduce repository-based exports with the offline review utility, for example:

```sh
go run ./tools/art-review \
  -reference gfx/hexTypes/hex_forest.png \
  -candidate art/source/pilot/forest-study-01.png \
  -out build/art-review/forest-repo
```

The native recipe needs the privately decoded original reference described in
`docs/art/original-references.md`. Review output names are `candidate-2x.png` and
`candidate-4x.png`; these were copied to the corresponding pilot directories above.
Use a fresh output directory for another run. PNGs are alpha weighted and constrained to
reference coverage. This alone cannot make material or feature joins seamless.

Known issues: generated canopy/ground details drift, and repository seven-hex patches show
dark joins. Native aspect, connected features and visual approval remain open. The next
production step is shared edge/corner masters and faithful cleanup, before mass generation.

The connected terrain studies include road revisions 01/02, river 01 and bridge 01.
Road 02 is the stronger study; revision 01 remains as its recorded edit input. See
[the review notes](../../../docs/art/connected-terrain-pilot.md) for tentative scores,
opacity/palette defects and exact reproduction commands, or open
[the local comparison page](../../../docs/art/connected-terrain-review.html) with private
references and reductions present in `build/`. Clipping yields the original binary coverage
but preserves semi-transparent interior pixels; these exports still need explicit cleanup.
