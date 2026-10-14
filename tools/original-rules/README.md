# DOS original rules investigation

Reproduce the static disassembly and data extraction for the locally supplied
`original/extracted/dos-full/HL.EXE`. Findings and limitations are in
[`docs/research/original-dos-rules.md`](../../docs/research/original-dos-rules.md).
Python 3 and GNU binutils `objdump` with `i8086` support are sufficient; no network,
DOS execution, external Python packages, or changes to the installation are needed.

From the repository root:

```sh
python3 -m unittest discover -s tools/original-rules -v
python3 tools/original-rules/disassemble.py
```

The default destination is ignored `build/original-rules/dos-full/`. Existing output
is refused; use `--out build/original-rules/another-run` to repeat. `--source` can
select another local copy of the same installation. The executable hash must match
the investigated build because the named code windows are specific to that version.

Outputs include the header-free load image, a complete **linear sweep** listing,
123 overlapping segment listings, ten traced instruction windows, relocation and
immediate far-call indexes, printable strings, 51 unit records, both 175-record
terrain tables, and a source/decoded-data hash manifest. DOS TPWM lengths are little
endian. Unknown record fields remain raw, with every record retained as hex.

Linear sweep also interprets data and jump tables as instructions. Segment boundaries
come from relocation locations and are not function boundaries. The explicit windows
start at traced instruction boundaries; the far-call index excludes near/indirect
calls and is not a complete control-flow graph. No listing is decompiled source.

Keep originals, generated listings, strings and full data tables local under ignored
directories. Only the analysis recipe, synthetic tests and authored findings belong
in Git. The synthetic tests can run without the original installation.

## Combat harness

`emu8086.py` is a small 8086 interpreter (integer subset, hooks for far addresses; unknown
opcodes raise). `dos_combat.py` loads the exported load image with its relocations, builds a
synthetic battle context and runs `2466:000e`; `dos_combat_model.py` is an independent
transcription of the same routine, and `compare_combat.py` runs both on random cases:

```sh
python3 -I tools/original-rules/compare_combat.py --cases 300
```

The interpreter tests and the transcription tests need no original files; the comparison test
is skipped when `build/original-rules/dos-full` is absent. Agreement shows the routine's
arithmetic as written, not how the game fills its inputs.

`dos_experience.py` does the same for the experience awards after a battle
(`02dc:0908–0abb`): `award` is the transcription, `run_window` executes that code window with
its near-call helpers hooked. The window can be run on its own because it only reads the
battle context and two saved strengths from the caller's frame.
