// Package levels embeds the shipped levels so the game also runs where there is no
// file system to read them from (the browser build).
package levels

import "embed"

// FS holds the shipped level files, named like "level1_cp.json".
//
//go:embed level?_cp.json
var FS embed.FS

// Original holds the imported maps of the original game (just levels-import),
// named like "original/full/05.json", and their "original/index.json".
//
//go:embed original/index.json original/full/*.json
var Original embed.FS

// Monthly holds the campaign's scenario files (internal/scenario), one per
// month, named like "monthly/1914-08-pulse.json".
//
//go:embed monthly/*.json
var Monthly embed.FS

// Skirmish holds the original's two-player maps as scenario files, named
// like "skirmish/01-track.json".
//
//go:embed skirmish/*.json
var Skirmish embed.FS
