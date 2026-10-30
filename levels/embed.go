// Package levels embeds the shipped levels so the game also runs where there is no
// file system to read them from (the browser build).
package levels

import "embed"

// Monthly holds the campaign's scenario files (internal/scenario), one per
// month, named like "monthly/1914-08.json", their backdrops and the
// months' newspaper front pages (internal/campaign), "monthly/news/1914-08.json".
//
//go:embed monthly
var Monthly embed.FS

// Skirmish holds the original's two-player maps as scenario files, named
// like "skirmish/01-track.json".
//
//go:embed skirmish/*.json
var Skirmish embed.FS
