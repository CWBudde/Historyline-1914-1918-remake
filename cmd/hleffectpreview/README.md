# hleffectpreview

Native preview of unit movement and effects on a grass fixture, driven by `internal/presentation`. It walks a
unit through the six facings (`n nw sw s se ne`), then plays hit, death, capture and round. Interface PNGs
load through `internal/assets` (candidates opt-in).

```sh
go run ./cmd/hl preview effects -family infantry      # any unit family slug, e.g. armored-train
go test ./internal/presentation ./cmd/hleffectpreview
go run ./cmd/hl preview effects -self-check -mode normal|reduced|skip -out build/review/effect-preview-normal
```

Keys: Space skips the step, R toggles reduced motion, Enter restarts, P pauses, Escape quits. `-self-check`
runs at a fixed 60 Hz, saves screenshots and `evidence.json` to a new directory, and fails if playback stalls.
