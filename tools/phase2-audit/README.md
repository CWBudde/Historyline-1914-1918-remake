# phase2-audit

Checks that every enabled runtime ID has its candidate art: ledgers, recipe pins, export dimensions and
bundle reports (terrain, units, portraits, interface, screens v1/v2). Writes `build/phase2-audit/`
(JSON + Markdown) and fails on any missing file, stale pin, wrong size or duplicate binding.

```sh
go run ./tools/phase2-audit
go test ./tools/phase2-audit
```

A clean report means the candidates exist, not that they are approved.
