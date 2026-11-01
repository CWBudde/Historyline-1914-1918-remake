//go:build !origtiles

package main

// load reports why the local comparison art is unavailable in this build.
func (a *origArt) load() {
	a.err = "build with -tags origtiles (just play-original)"
}
