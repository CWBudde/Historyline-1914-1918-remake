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

The movement, route, container and visibility models are not written apart: they live in
`internal/dosrules` (`move.go`, `route.go`, `container.go`, `visibility.go`) and the runners
below compare that one implementation with the executed routines. Their record-level functions
return the original records for that comparison. `board.go` reads them into typed values for
the command cycle: `Live` and `Hold` records, and `Profile.Destinations`, `Route`, `EndCheck`,
`Move`, `Exit` and `ChangePhase`. `TestTyped*` checks those against every recorded fixture.

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

## Movement harness

`Export.RunMovement` runs the demo's movement queries (`Build.Move`), and
`internal/dosrules/move.go` is their transcription:
- the reach search `0b06:0331` with its cost routine `0b06:09c1`;
- the destination list `0b06:1383` with the end-cell check `10cc:0f74`;
- the path search `0b06:0e12`.

The queries run on synthetic maps that hold the build's own type and terrain records
(`WW1GNDS.DAT.bin`), in the order the game calls them. Only the stack check is hooked. The
runner fails if a query writes past the cost buffer, and if a path search changes the reach
bits.

`TestMovementCoverage` compares them on every enumerated case:
- every distinct type on every distinct terrain;
- transitions between terrains of distinct mask and cost;
- own, enemy and neutral units in all six directions from both column parities and for both
  sides;
- units only on the other side's map;
- the terrain flag masks the callers pass;
- budgets from 0 to 255;
- map edges and random maps.

It also checks the rules these cases show on the executed outputs. The fixture recorder
writes the type and terrain fields the queries read into the profile's tables (`move` in
`internal/dosrules/demo.json`) and the cases into `movements.json` (schema
`docs/research/movement-fixture.schema.json`) beside the other dosrules fixtures.
`dosrules.TestMovementFixtures` replays them without the original files.

`Export.RunMove` runs a move order the way the player gives it: the reach search, the order code
`10cc:02dd`, the move start `10cc:05c3` and the step `10cc:071e` once per call until the unit
arrives (`10cc:0ae2`). Only the drawing, the sound and the status message are hooked; the
message codes are recorded. `internal/dosrules/route.go` is the transcription. `TestStepCoverage` compares
them on the routes of every movement class, the refused orders, an order changed before its
start, own units passed, hidden enemies on the route and the destination, a two-cell ship and an
empty transporter. It checks on the executed outcomes that live byte 0 is not spent, that every
route is walked to its end and that the other side's map is untouched. The recorder writes
`steps.json` (schema `docs/research/step-fixture.schema.json`), and `dosrules.TestStepFixtures`
replays it without the original files.

The container records (`Movement.Holds`, `Movement.Held`) are set up in both runners. `RunMove`
therefore also executes the end-cell check's container branches and the arrival into a
transporter or building. `Export.RunExit` runs the main loop's exit branch `0661:2f0a–30fe` on
a synthetic frame holding the order record, then the move order as `RunMove` does; the view
centring and redraw are hooked as well. `internal/dosrules/container.go` is the transcription (`EndCheck`,
`Exit`). `TestContainerCoverage` compares them on:
- loading every distinct type into each transporter type;
- slots, weights, a moved transporter and nested cargo;
- every building terrain and owner;
- carried cargo;
- exits from transporters and buildings.

It checks on the executed outcomes that no unit leaves a slot without an exit or a hand-over,
that capacities hold and that the exit's reach starts at 4 or 8. The recorder writes
`containers.json` (schema `docs/research/container-fixture.schema.json`), and
`dosrules.TestContainerFixtures` replays it without the original files.

`Export.RunPhaseChange` runs the tail of the order handler, `02dc:1a09–1cc6`, on a frame as the
handler leaves it after the attacking side's orders. This is where one side's view of the other
is brought up to date; the demo has no sight range. Both maps, the live units, the container
records and the per-side records are set up from a `PhaseChange`. The window calls only the
container sync, the side test and the map copy; the status message is hooked, and code in any
other segment fails the run. `internal/dosrules/visibility.go` is the transcription (`ChangePhase`), and
`TestVisibilityCoverage` compares them on:
- phase bytes and the turn counter;
- units of each side with views that agree or differ;
- alive, dying, destroyed and held units, and ships;
- differing terrain and building markers;
- container records of every owner;
- random maps.

It checks on the executed outcomes that both maps agree afterwards, that every unit on them
stands at its side-0 position word, and that the slot lists agree. `Engagement.Moved` gives the
order handler a different view for the other side, and `TestEngagementHidden` shows that attacks
resolve on the attacking side's view. The recorder writes `visibility.json` (schema
`docs/research/visibility-fixture.schema.json`), and `dosrules.TestVisibilityFixtures` replays it
without the original files.

`TestConsumables` looks for ammo, fuel and resupply. It repeats the engagement, move, container
and phase-change cases with the emulator's write watch on (`dos8086.Machine.OnWrite`, which
`Export.watch` installs on every machine a runner creates). It counts each write into the live
and type records by field, writing instruction and direction, and checks that:
- nothing writes a type record or live byte 0;
- only named fields are written;
- only group size, flags, positions, facings and targets go down;
- group size, alternate strength and the builder charges never go up;
- no unit field other than a position or facing is written per cell walked.

It needs the export but records no fixture: what it shows is an absence. The static side of
the census is in `docs/research/original-dos-rules.md` (*Ammo, fuel and resupply*).
