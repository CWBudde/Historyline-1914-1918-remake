package campaign

import "testing"

func TestPeriodsRunThroughTheWar(t *testing.T) {
	if len(Periods) != 24 {
		t.Fatalf("%d periods, want 24 two-month layers", len(Periods))
	}
	if Periods[0].Label != "August – September 1914" || Periods[23].Label != "June – July 1918" {
		t.Errorf("first %q, last %q", Periods[0].Label, Periods[23].Label)
	}
	for i, p := range Periods {
		if p.Headline == "" || len(p.Lines) == 0 {
			t.Errorf("period %d (%s) has no news", i, p.Label)
		}
		for _, l := range p.Lines {
			if len(l) > 44 {
				t.Errorf("period %d line too wide for the news column: %q", i, l)
			}
		}
	}
}

func TestMonthsStartWithSeptember1914(t *testing.T) {
	if len(Months) != 1 {
		t.Fatalf("%d monthly layers, want the first one so far", len(Months))
	}
	p, ok := Month(0)
	if !ok || p.Label != "September 1914" || p.Headline == "" || len(p.Lines) == 0 {
		t.Fatalf("month 0: %+v ok %v", p, ok)
	}
	for _, l := range p.Lines {
		if len(l) > 44 {
			t.Errorf("monthly line too wide for the news column: %q", l)
		}
	}
	if _, ok := Month(1); ok || len(Periods) != 24 {
		t.Error("month 1 must not exist yet, and the two-month layers stay 24")
	}
}

func TestForMapsEachCampaignThroughThePeriods(t *testing.T) {
	for mapIndex, want := range map[int]int{0: 0, 1: 1, 23: 23, 24: 0, 47: 23, 48: 0, 71: 23} {
		p, ok := For(mapIndex)
		if !ok || p.Label != Periods[want].Label {
			t.Errorf("map %d: %+v ok %v, want period %d", mapIndex, p, ok, want)
		}
		if got := Index(mapIndex); got != want {
			t.Errorf("map %d index %d, want %d", mapIndex, got, want)
		}
	}
	if _, ok := For(72); ok || Index(-1) != -1 {
		t.Error("map 72 must have no period")
	}
}
