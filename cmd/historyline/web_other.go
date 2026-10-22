//go:build !js

package main

// pageReady does nothing outside the web build.
func pageReady() {}

// pageFetched does nothing outside the web build.
func pageFetched(int64) {}
