// Package levels embeds the shipped level folders (internal/scenario) so the
// game also runs where there is no file system to read them from (the
// browser build).
package levels

import "embed"

// Levels holds the level folders: the campaign months, "1914-08/map.json"
// with a setup file per mode, the backdrop "1914-08/backdrop.jpg" and the
// month's newspaper front page "1914-08/news.json" (internal/campaign). The
// backdrops' lossless sources (backdrop.png) stay out.
//
//go:embed 19*/*.json 19*/backdrop.jpg
var Levels embed.FS
