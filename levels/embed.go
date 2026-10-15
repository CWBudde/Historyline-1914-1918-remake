// Package levels embeds the shipped levels so the game also runs where there is no
// file system to read them from (the browser build).
package levels

import "embed"

// FS holds the shipped level files, named like "level1_cp.json".
//
//go:embed level?_cp.json
var FS embed.FS
