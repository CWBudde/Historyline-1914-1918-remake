package main

import (
	"testing"
)

// Catalogue art can be exported before a unit has a gameplay-table binding.
// It must retain every palette/facing/state without inventing gameplay classes.
func TestUnboundTankBundle(t *testing.T) {
	path, _ := unitTestFamilyFixture(t, "a7v")
	in, err := loadUnitInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := buildUnitBundle(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.report.ClassIDs) != 0 || len(b.report.Variants) != 36 {
		t.Fatalf("classes=%v variants=%d", b.report.ClassIDs, len(b.report.Variants))
	}
	for _, palette := range []string{"cp", "ec", "neutral"} {
		for _, face := range []string{"n", "nw", "sw", "s", "se", "ne"} {
			for _, state := range []string{"normal", "spent"} {
				for _, density := range []int{2, 4} {
					name := unitSpriteName("a7v", palette, face, state, density)
					im := b.pixels[name]
					if im == nil || im.Bounds().Dx() != 96*density || im.Bounds().Dy() != 96*density {
						t.Fatalf("missing or invalid export %s", name)
					}
				}
			}
		}
	}
	for _, scene := range b.report.Scenes {
		for _, cell := range scene.Cells {
			if cell.RuntimeID != "" {
				t.Fatalf("invented gameplay binding %q", cell.RuntimeID)
			}
		}
	}
}

func TestRemainingPortraitCatalogue(t *testing.T) {
	for _, name := range []string{
		"anti-tank", "static-aa", "mobile-aa", "a7v", "st-chamond", "renault-ft17", "mark-i", "mark-iv",
		"torpedo-boat", "destroyer", "battleship", "transport-ship", "submarine", "static-balloon",
		"fokker-e-i", "fokker-e-iii", "albatros-d-iii", "fokker-dr-i", "fokker-d-vii", "junkers-j4",
		"gotha-g-vb", "zeppelin-staaken-r-vi", "dh2", "morane-saulnier-n", "nieuport-xvii", "spad-vii",
		"se5a", "spad-xiii", "sopwith-camel", "voisin-iii", "handley-page-0400", "dh4", "construction-unit",
	} {
		t.Run(name, func(t *testing.T) {
			definition, err := lookupUnitFamily(name)
			if err != nil {
				t.Fatal(err)
			}
			if definition.name != name || len(definition.classIDs) != 0 {
				t.Fatalf("unbound catalogue family: %+v", definition)
			}
		})
	}
}
