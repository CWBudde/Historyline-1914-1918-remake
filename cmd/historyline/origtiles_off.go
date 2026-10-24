//go:build !origtiles

package main

import "errors"

// loadOrigTiles is not in this build: only -tags origtiles reads the original
// tiles.
func loadOrigTiles() (*origTiles, error) {
	return nil, errors.New("build with -tags origtiles (just play-original)")
}
