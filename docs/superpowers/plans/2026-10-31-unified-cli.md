# Unified CLI Implementation Plan

> **For agentic workers:** Use executing-plans to implement this plan inline.

**Goal:** Expose all developer tools through `hl` and commit directly to main.

**Architecture:** A standard-library command tree dispatches to importable Go
tool packages. Native graphics and tagged test generators run as explicit
child processes; ordinary tools call libraries directly.

**Tech Stack:** Go 1.25, standard `flag`, existing Go packages.

**Spec:** `docs/superpowers/specs/2026-10-31-unified-cli-design.md`

## Global Constraints

- Preserve every existing capability and flag; no new dependencies.
- No art, assets, scenarios or generated catalogues change in this migration.
- Keep build/tool caches outside `/mnt/Projekte`.
- Commit directly to `main`; preserve other work.

## Review Focus

- Help for every command must succeed with DISPLAY unset and have no side effects.
- Invalid arguments must not invoke exporters or mutate files.
- Moved fixtures and embedded pages must resolve from their new paths.
- Candidate review must work without the previous command binaries or Go builds.
- Rejected LLM orders must preserve the game and return exit code 2.

### Task 1: Command routing and package migration

**Files:** `cmd/hl/main.go`, `internal/hlcli/`, moved `cmd/` tooling packages,
and existing offline `tools/` packages.

**Interfaces:** `hlcli.Run(ctx, args, stdin, stdout, stderr) int` is the process
boundary; tool `Run(args, stdout, stderr) error` functions never exit.

- [x] Write CLI integration tests for all help paths, invalid commands and flags,
  asset validation, and an LLM session/rejection; run before implementation.
- [x] Move tool implementations, tests, fixtures and embedded files, retaining
  directory depth. Replace global flags and process exits with local flag sets
  and returned errors; thread output writers through command boundaries.
- [x] Implement the command tree and `cmd/hl` entry point. Run focused tests.

### Task 2: Native preview and tagged generators

**Files:** `internal/hlcli/`, `internal/effectconfig/`,
`cmd/hleffectpreview/main.go`.

**Interfaces:** Shared preview flags provide headless help. Named generators
resolve to fixed Go package, build tag and exact test name.

- [x] Test generator discovery/invalid names, preview argument validation and
  help without spawning graphics or generators.
- [x] Implement explicit child processes with context and inherited streams;
  prefer an installed preview helper, otherwise use the source checkout.
- [x] Run CLI tests with DISPLAY unset; never execute asset generators.

### Task 3: Consumers and documentation

**Files:** `tools/candidate-review/`, subprocess tests, `justfile`, `.github/`,
`README.md`, `AGENTS.md`, tooling docs (excluding `docs/manual/`).

**Interfaces:** Candidate review calls `artreview.Run`, `unitpanel.Run` and
`interfaceart.Run`; no build/run subprocess for these tools.

- [x] Replace candidate-review subprocess exports with library calls.
- [x] Update just recipes, CI commands, command examples, fixture references
  and subprocess tests. Add a tool-build recipe and CLI usage guide.
- [x] Run package tests and inspect all remaining old command references.

### Task 4: Verification and commit

**Files:** Only files belonging to this migration.

- [x] Run the repository checks, including the existing slow tool tests, lint,
  desktop/WASM builds, art and reference checks. Inspect failures and reports.
- [x] Run race tests for the new command boundary and inspect the full diff.
- [x] Review the migration against the spec and fix material issues.
- Commit only migration changes directly to `main`, then push origin main
  according to repository instructions after saving this verification record.

Local verification passed for the CLI/race tests, moved asset-pack tests, slow
art-review exports, other offline exporters, desktop/WASM builds, catalogue and
reference checks. Full checks retain existing failures for missing local
reference data, a stale unit-panel source pin, and the game's original-art
`SA4023` warning. Selected art, assets and levels are unchanged.
