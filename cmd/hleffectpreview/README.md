# Candidate movement and effects preview

This native Ebitengine command loads new-art disk candidates. It is isolated from
the gameplay rules and does not accept an external runtime pack. The reusable
scheduler in `internal/presentation` consumes immutable visual descriptions using caller time.
Selected interface PNGs now load through `internal/assets` with explicit candidate
opt-in; see [the runtime pack contract](../../docs/runtime-assets.md).

From the repository root:

```powershell
go run ./cmd/hleffectpreview -family infantry
```

The sequence moves through `n`, `nw`, `sw`, `s`, `se`, `ne` using the six existing
independently authored CP normal-state sprites, then plays hit, death, capture and
round in order. Sprites retain logical anchor `(48,49)` over terrain center
`(48,40)`. The white cross exposes that registration. No image is rotated.
The hit count, capture label and round label are live fixture text, not PNG text.

Controls: **Space** skips the active step; **R** switches reduced motion and
restarts; **Enter** restarts; **P** pauses; **Escape** closes. Reduced motion snaps
movement to its destination and uses the finite reduced effect definitions.
Nothing loops automatically.

`-family` accepts the delivered pack slugs: `infantry`, `elite-infantry`,
`cavalry`, `sappers`, `light-artillery`, `medium-artillery`, `heavy-artillery`,
`supply-car`, `bunker`, `armored-car`, `charron`, `patrol-boat`, `transport-train`,
`armored-train`, `train-artillery`. The default is `infantry`. Each uses the same
validated 96×96 logical canvas and anchor. Moving a bunker or displaying a boat
on grass here is a graphics diagnostic; it does not imply legal gameplay movement
or terrain suitability. The selected family appears on screen and in evidence.

## Bounded native evidence

Each output directory must be new and its parent must already exist. The
self-check opens a native window, advances at fixed simulated 60 Hz, saves real
GPU-rendered screenshots for each movement heading's early sample and midpoint
and each effect frame,
writes `evidence.json`, and exits. It fails after 1200 ticks if playback stalls.

```powershell
go test ./internal/presentation ./cmd/hleffectpreview
go build -o bin/hleffectpreview.exe ./cmd/hleffectpreview
bin/hleffectpreview.exe -self-check -mode normal -out build/effects-normal
bin/hleffectpreview.exe -self-check -mode reduced -out build/effects-reduced
bin/hleffectpreview.exe -self-check -mode skip -out build/effects-skip
bin/hleffectpreview.exe -family armored-train -self-check -mode normal -out build/effects-armored-train
```

The skip check skips each active step after 150 ms. Evidence records exact input
SHA-256s, ordered frame/position/facing snapshots, final state and tick count.
Interface frame PNGs, animation/binding metadata and the grass fixture are
validated against the candidate export's hashes. Unit export bundles do not yet
include output hash reports; those six PNGs are dimension-checked (192×192) and
fingerprinted. Hashes identify supplied candidates rather than certify authorship.
All reads use a repository-scoped `os.Root`; original and legacy path components
are rejected. No original bytes are embedded or copied to the evidence.

This is a small grass fixture and presentation smoke test. It does not establish
full-map contrast, subjective likeness, all-family acceptance, gameplay integration,
browser playback, accessibility review or native runtime-pack acceptance.
