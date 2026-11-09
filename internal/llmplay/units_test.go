package llmplay

import (
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// A level ref with UnitsSuffix plays the extension units in; the ref stays
// in the header, so a replay starts the same game.
func TestUnitsRef(t *testing.T) {
	month, mode, err := ParseLevelRef("1918-02:cp" + UnitsSuffix)
	if err != nil || month != "1918-02" || mode != scenario.CP || !ExtendedRef("1918-02:cp+units") || ExtendedRef("1918-02:cp") {
		t.Fatalf("ParseLevelRef: %q %q %v", month, mode, err)
	}
	g, err := StartGame("1918-02:cp+units", t.TempDir(), DriverLLM, false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := g.Close(); err != nil {
			t.Error(err)
		}
	}()
	if g.S.Game.Profile.Tables != dosrules.ExtendedTables {
		t.Errorf("tables %q", g.S.Game.Profile.Tables.Build)
	}
	if strings.Contains(Observe(g.S, 0, "", Window{}), "GAS (") {
		t.Error("a game without a cloud lists gas")
	}
}
