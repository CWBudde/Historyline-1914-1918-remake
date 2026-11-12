package main

// fogState is what the shown side knows about a hex.
type fogState uint8

const (
	fogUnseen  fogState = iota // never seen
	fogSeen                    // seen before, not visible now
	fogVisible                 // in sight now
)
