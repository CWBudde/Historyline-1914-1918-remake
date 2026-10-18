# DOS original rules investigation

Reproduce the static disassembly and data extraction for the locally supplied
`original/extracted/dos-full/HL.EXE`. Findings and limitations are in
[`docs/research/original-dos-rules.md`](../../docs/research/original-dos-rules.md).
Go and GNU binutils `objdump` with `i8086` support are sufficient; no network,
DOS execution, or changes to the installation are needed.

From the repository root:

```sh
go test ./internal/dosexe ./internal/tpwm
go run ./cmd/dosexport
```

The default destination is ignored `build/original-rules/dos-full/`. Existing output
is refused; use `--out build/original-rules/another-run` to repeat. `--source` can
select another local copy of the same installation. The executable hash must match
an identified build (`dosoracle.Builds`) because the named code windows and runner
addresses are specific to each version.

The free DOS demo (`hl-dem.zip`, see the research notes) is the second identified build.
Its `HL14-18.EXE` is exported the same way, with five traced windows and only the summer
terrain table:

```sh
go run ./cmd/dosexport -source original/demo/HL -out build/original-rules/dos-demo
```

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
- `Export.RunDistance` runs the hex distance `0e40:2421`; `Distance` is the transcription.
- `Export.RunEngagement` runs the order handler `02dc:0004` for one attack order on a map
  holding the build's own type records, through the neighbour list, support call, battle
  screen, casualty routine and experience window, and stops after the first battle. Only
  drawing, scrolling, sound and the routines named above are hooked; the seed callback is the
  game's own and the RNG draw fails if called. `Engage` composes the transcriptions in the
  handler's order. The handler is located in the demo only (`Build.Handler`); the engagement
  tests skip a build without it.

`internal/dosrules` is the pure Go resolver of profile `dos-0.1`, written apart from the
transcriptions and free of the emulator. `dosrules.Demo` binds it to the demo's tables
(`internal/dosrules/demo.json`, numbers only); `Profile.TransposedDistance` switches the battle
distance between the executed (row, column) call and the layout distance. The engagement and
coverage tests compare it with the emulator on every case, and its own tests compare it with
fixtures recorded from the emulator, so CI checks it without the original files. To re-record
the tables and fixtures from a local export (and, without `-update`, to check them):

```sh
go test ./internal/dosoracle -run TestFixtures -update
```

```sh
go test ./internal/dos8086 ./internal/dosoracle -v
```

The interpreter and transcription tests need no original files; the agreement and coverage
tests run once per export present in `build/original-rules/dos-full` and `dos-demo` (subtests
`full` and `demo`) and are skipped when neither exists. Agreement shows the routine's
arithmetic as written, not how the game fills its inputs.

The coverage tests (R1) enumerate the input classes instead of sampling them: group sizes
0–8 and 255, experience 0–8, 127, 128 and 255, the four target-flag combinations, encoded
ranges 0–3, the build's maximum and 255 at every distance up to R+1, all 81 terrain-class
pairs, pincer 0–250 against blocking 0/80/150/160/255, the return-fire suppressions,
alternate strength, a power × armour grid, every type record against every other, all six
support directions for both column parities with each eligibility variant, and experience
0–6 against kill/loss/no-loss outcomes. A branch trace of `Resolve` (`Trace`) must show
each path taken both ways.

The engagement tests (R1, controlled battles) run real unit types through the handler:
duels in all six directions for both sides, all 81 terrain pairs, experience 0–6 with roles
swapped, an empty target, one or two blockers and every pincer step with eligible and
ineligible units, ranged attacks at layout distance 2–4, adjacent defenders that never
return fire, and kills and losses with and without the no-award bit.
