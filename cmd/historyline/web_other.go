//go:build !js

package main

// pageReady does nothing outside the web build.
func pageReady() {}

// pagePauser returns nil: a background load has threads of its own outside
// the web build.
func pagePauser() func() { return nil }

// pageFetched does nothing outside the web build.
func pageFetched(int64) {}
