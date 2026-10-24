//go:build !origtiles

package main

import "errors"

// loadOrigArt is unavailable without -tags origtiles: the original art is a
// local comparison aid and never ships.
func loadOrigArt() (*origArt, error) {
	return nil, errors.New("build with -tags origtiles (just play-original)")
}
