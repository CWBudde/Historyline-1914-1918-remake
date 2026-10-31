package candidatereview

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		c    candidate
		want string
		lock bool
	}{
		{"unrelated", candidate{Kind: "other", Family: "pinball"}, "delete", false},
		{"unit earlier step", candidate{Kind: "unit", Family: "dh2", Near: "art/units/dh2/atlas.png", NearDist: 3.5}, "delete", false},
		{"unit near another family", candidate{Kind: "unit", Family: "dh2", Near: "art/units/dh4/atlas.png", NearDist: 3.5}, "", false},
		{"unit alternative", candidate{Kind: "unit", Family: "dh2", Near: "art/units/dh2/atlas.png", NearDist: 8}, "", false},
		{"interface near-identical", candidate{Kind: "interface", Near: "art/interface/rank/07.png", NearDist: 1.2}, "delete", false},
		{"interface alternative", candidate{Kind: "interface", Near: "art/interface/rank/07.png", NearDist: 4}, "", false},
		{"terrain named keeper", candidate{Kind: "terrain", Exec: "exec-c8db8578-0000", Family: "grass"}, "keep", true},
		{"terrain pixel redraw", candidate{Kind: "terrain", Exec: "exec-1", Note: "Amiga-faithful pixel redraw"}, "delete", true},
		{"terrain other", candidate{Kind: "terrain", Exec: "exec-2", Family: "water", NearDist: 20}, "delete", true},
	}
	for _, tc := range cases {
		c := tc.c
		classify(&c)
		if c.Default != tc.want || c.Locked != tc.lock {
			t.Errorf("%s: got %q locked=%v, want %q locked=%v", tc.name, c.Default, c.Locked, tc.want, tc.lock)
		}
		if c.Default != "" && c.Reason == "" {
			t.Errorf("%s: decision without reason", tc.name)
		}
	}
}

func TestReadDecisions(t *testing.T) {
	dir := t.TempDir()
	if d, err := readDecisions(filepath.Join(dir, "missing.tsv")); err != nil || d != nil {
		t.Fatalf("missing file: %v %v", d, err)
	}
	good := filepath.Join(dir, "good.tsv")
	if err := os.WriteFile(good, []byte("exec-a\tkeep\tW1 reference\nexec-b\tdelete\tpool blue\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := readDecisions(good)
	if err != nil || d["exec-a"] != [2]string{"keep", "W1 reference"} || d["exec-b"][0] != "delete" {
		t.Fatalf("got %v %v", d, err)
	}
	bad := filepath.Join(dir, "bad.tsv")
	if err := os.WriteFile(bad, []byte("exec-a\tmaybe\tx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readDecisions(bad); err == nil {
		t.Fatal("want error for an unknown decision")
	}
}
