package mapcompose

import (
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// campaignLevel is a setup of a shipped scenario, "monthly/1915-02:cp".
func campaignLevel(t *testing.T, name string) *level.Level {
	t.Helper()
	l, err := campaignLevelErr(name)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func campaignLevelErr(name string) (*level.Level, error) {
	file, mode, _ := strings.Cut(name, ":")
	return scenario.ShippedLevel(file, scenario.Mode(mode))
}

// allLevels are every shipped scenario map once (scenario.Shipped(scenario.CP)).
func allLevels(t *testing.T) []scenario.Named {
	t.Helper()
	sh, err := scenario.Shipped(scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	return sh
}
