package main

import (
	"os"
	"strings"
	"testing"
)

func TestPulseLoadsSharedGrassVariants(t *testing.T) {
	r, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	c := collector{root: r}
	contexts, err := c.pulseContexts()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for id, pictures := range contexts {
		if !strings.HasPrefix(id, "tile/grass-detail-") {
			continue
		}
		count++
		for _, density := range []string{"2", "4"} {
			p := pictures[density]
			if p.SHA == "" || !strings.HasSuffix(p.Source, "-"+density+"x.png") {
				t.Errorf("missing pinned %sx export for %s", density, id)
			}
		}
	}
	if count != 16 {
		t.Fatalf("PULSE loaded %d grass details, want sixteen", count)
	}
}
