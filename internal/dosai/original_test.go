package dosai

import (
	"fmt"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/origview"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

// originalTurns bounds a played-out original map.
const originalTurns = 40

// original is a session on imported original map i, both sides computer.
func original(tb testing.TB, i int) *dosplay.Session {
	tb.Helper()
	raw, err := levels.Original.ReadFile(fmt.Sprintf("original/full/%02d.json", i))
	if err != nil {
		tb.Fatal(err)
	}
	l, err := finmap.Parse(raw)
	if err != nil {
		tb.Fatalf("map %02d: %v", i, err)
	}
	b, err := dosbind.FromOriginal(l, dataSet(tb))
	if err != nil {
		tb.Fatalf("map %02d: %v", i, err)
	}
	lv, err := origview.Level(l)
	if err != nil {
		tb.Fatalf("map %02d: %v", i, err)
	}
	s, err := dosplay.NewBoard(b, lv, dataSet(tb), dosplay.Options{Computer: [2]bool{true, true}})
	if err != nil {
		tb.Fatalf("map %02d: %v", i, err)
	}
	return s
}

// The computer plays both sides of every original map for originalTurns
// resolutions (with -short every sixth map) without an engine error or a
// refused move.
func TestPlayOriginalMaps(t *testing.T) {
	for i := range 72 {
		if testing.Short() && i%6 != 0 {
			continue
		}
		s := original(t, i)
		var evs []dosplay.Event
		end, err := playOut(s, [2]*Planner{{}, {}}, originalTurns, func(e []dosplay.Event) { evs = append(evs, e...) })
		if err != nil {
			t.Fatalf("map %02d: %v", i, err)
		}
		for _, ev := range evs {
			if ev.Kind == dosplay.Refused {
				t.Fatalf("map %02d: a move of unit %d to %v was refused: %s", i, ev.Unit, ev.To, ev.Text)
			}
		}
		state := "open"
		if end != nil && end.Over {
			state = fmt.Sprintf("winner %d (%v)", end.Winner, end.Won)
		}
		t.Logf("map %02d: %s after turn %d; %d moves, %d attack orders, %d battles", i, state, s.Turn(),
			count(evs, dosplay.Moved), count(evs, dosplay.Ordered), count(evs, dosplay.Battle))
	}
}
