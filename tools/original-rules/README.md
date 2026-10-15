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

`internal/dos8086` is a small 8086 interpreter (integer subset, hooks for far addresses;
unknown opcodes return an error). `internal/dosoracle` loads the exported load image with
its relocations and runs the real routines on synthetic contexts, next to independent
transcriptions of the same code:

- `Export.Battle` runs the casualty routine `2466:000e`; `Resolve` is the transcription.
- `Export.RunExperienceWindow` runs the experience awards after a battle (`02dc:0908-0abb`) with
  its near-call helpers hooked; `Award` is the transcription. The window can run on its own
  because it only reads the battle context and two saved strengths from the caller's frame.
- `Export.RunSupport` runs the pincer and blocking caller `02dc:262a` on a synthetic map,
  executing the real neighbour routine `0b06:0157` and side test `0ca6:020a`; `Support` is the
  transcription. The ranged case with an off-map neighbour reads an uninitialised stack byte,
  so the transcription returns `ErrUndefined` there and the comparison skips it.

```sh
go test ./internal/dos8086 ./internal/dosoracle -v
```

The interpreter and transcription tests need no original files; the agreement tests read
`build/original-rules/dos-full` and are skipped when it is absent. Agreement shows the routine's
arithmetic as written, not how the game fills its inputs.
