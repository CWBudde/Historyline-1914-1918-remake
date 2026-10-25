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

func TestMonths(t *testing.T) {
	for month, p := range Months {
		if PeriodOf(month) < 0 || p.Headline == "" || len(p.Lines) == 0 {
			t.Errorf("month %s: %+v", month, p)
		}
		for _, l := range p.Lines {
			if len(l) > 44 {
				t.Errorf("month %s line too wide for the news column: %q", month, l)
			}
		}
	}
	if p, ok := News("1914-09"); !ok || p.Label != "September 1914" {
		t.Errorf("1914-09: %+v", p)
	}
	if p, ok := News("1914-08"); !ok || p.Label != Periods[0].Label {
		t.Errorf("1914-08: %+v, want the first period", p)
	}
	if _, ok := News("1918-08"); ok {
		t.Error("1918-08 has news")
	}
}

func TestPeriodOf(t *testing.T) {
	for month, want := range map[string]int{
		"1914-08": 0, "1914-09": 0, "1914-10": 1, "1915-01": 2, "1918-06": 23, "1918-07": 23,
		"1914-07": -1, "1918-08": -1, "1914-8": -1, "x": -1, "1914-13": -1,
	} {
		if got := PeriodOf(month); got != want {
			t.Errorf("PeriodOf(%q) = %d, want %d", month, got, want)
		}
	}
}
