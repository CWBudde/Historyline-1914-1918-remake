# unit-panel

Builds the 400×640 logical unit information panel (2× and 4×) from `art/unit-panel-v1.json`: text-free
vector frame and icons, live Go-font text, nine fixtures from real `data.Load` / `game.NewUnit` state.

```sh
go run ./tools/unit-panel -manifest art/unit-panel-v1.json -out build/gfx-production/unit-panel/new
go test ./tools/unit-panel
```

Owner sets the tactical colour; country stays type data. Speed is base speed; rank, size, fuel and
ammunition come from runtime state; weapon slots distinguish empty, unlimited and zero.
