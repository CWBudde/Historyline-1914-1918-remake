package main

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"

// shippedLevels are every shipped scenario map once
// (scenario.Shipped(scenario.CP)).
func shippedLevels() ([]scenario.Named, error) {
	return scenario.Shipped(scenario.CP)
}
