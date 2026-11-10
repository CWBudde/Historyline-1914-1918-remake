//go:build !js

package main

import "path/filepath"

// pageReady does nothing outside the web build.
func pageReady() {}

// pagePauser returns nil: a background load has threads of its own outside
// the web build.
func pagePauser() func() { return nil }

// pageFetched does nothing outside the web build.
func pageFetched(int64) {}

// unitCatalog is the unit art: the pack just packs builds.
func unitCatalog() string { return filepath.Join(packsDir, "units") }
