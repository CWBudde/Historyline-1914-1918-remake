package main

import (
	"bytes"
	"fmt"
	"image"
	"strings"
	"testing"
)

func TestWidgetsHaveCompleteSemanticAndInteractionStates(t *testing.T) {
	known := map[string]Art{}
	for _, a := range widgetCatalog() {
		if _, duplicate := known[a.ID]; duplicate {
			t.Fatalf("duplicate widget: %s", a.ID)
		}
		known[a.ID] = a
	}
	if len(known) != 61 {
		t.Fatalf("got %d widget candidates, want 61", len(known))
	}
	for _, family := range []string{"input", "input/invalid", "toggle/off", "toggle/on", "toggle/mixed", "slider/track", "slider/fill", "slider/thumb", "dropdown/closed", "dropdown/open", "dropdown-option/normal", "dropdown-option/selected"} {
		seen := map[string]bool{}
		for _, state := range []string{"normal", "hover", "focus", "pressed", "disabled"} {
			id := "widget/" + family + "/" + state
			a, ok := known[id]
			if !ok {
				t.Fatalf("missing %s", id)
			}
			key := string(a.svg())
			if seen[key] {
				t.Fatalf("interaction state has no visual distinction: %s", id)
			}
			seen[key] = true
		}
	}
}

func TestWidgetPathsFitLogicalCanvasAndNativeReduction(t *testing.T) {
	for _, a := range widgetCatalog() {
		t.Run(a.ID, func(t *testing.T) {
			if strings.Contains(string(a.svg()), "<text") {
				t.Fatal("runtime asset contains baked label")
			}
			for _, s := range a.Shapes {
				for _, c := range s.Commands {
					for i := 0; i < len(c.Points); i += 2 {
						x, y := c.Points[i], c.Points[i+1]
						if x < 0 || x > float64(a.Width) || y < 0 || y > float64(a.Height) {
							t.Fatalf("path control point outside canvas: %g,%g", x, y)
						}
					}
				}
			}
			four := a.raster(4)
			two, err := resizeLanczos(four, a.Width*2, a.Height*2)
			if err != nil {
				t.Fatal(err)
			}
			one, err := resizeLanczos(two, a.Width, a.Height)
			if err != nil || one.Bounds() != image.Rect(0, 0, a.Width, a.Height) {
				t.Fatalf("invalid native reduction: %v", err)
			}
		})
	}
}

func TestWidgetValuesRemainDistinctWithoutColor(t *testing.T) {
	known := map[string]Art{}
	for _, a := range widgetCatalog() {
		known[a.ID] = a
	}
	for _, family := range []struct {
		prefix string
		values []string
	}{{"toggle", []string{"off", "on", "mixed"}}, {"dropdown", []string{"closed", "open"}}, {"dropdown-option", []string{"normal", "selected"}}} {
		for _, state := range states {
			seen := map[string]bool{}
			for _, value := range family.values {
				a := known[fmt.Sprintf("widget/%s/%s/%s", family.prefix, value, state)]
				a.Shapes = append([]Shape(nil), a.Shapes...)
				// Remove colors so this checks geometry, rather than hue alone.
				for i := range a.Shapes {
					a.Shapes[i].Fill = brass
				}
				key := string(writeJSON(a.Shapes))
				if seen[key] {
					t.Fatalf("value differs only in color: %s", a.ID)
				}
				seen[key] = true
			}
		}
	}
	for _, state := range states {
		normal := known["widget/input/"+state]
		invalid := known["widget/input/invalid/"+state]
		if bytes.Equal(normal.svg(), invalid.svg()) || len(invalid.Shapes) <= len(normal.Shapes) {
			t.Fatalf("validation has no separate non-color indicator: %s", state)
		}
	}
}

func TestWidgetSliderBoundsAndDropdownRows(t *testing.T) {
	g := widgetGeometry()
	if g.SliderCenters[0] != g.SliderTrack[0] || g.SliderCenters[1] != g.SliderTrack[0]+g.SliderTrack[2] {
		t.Fatal("slider track and center travel disagree")
	}
	for _, value := range []float64{0, .25, .5, .75, 1} {
		center := float64(g.SliderCenters[0]) + float64(g.SliderTrack[2])*value
		left, right := center-float64(g.SliderThumbSize[0])/2, center+float64(g.SliderThumbSize[0])/2
		if left < 0 || right > 220 || g.SliderCenterY-g.SliderThumbSize[1]/2 < 0 || g.SliderCenterY+g.SliderThumbSize[1]/2 > 32 {
			t.Fatalf("thumb clipped at value %g", value)
		}
	}
	for i, row := range g.DropdownRows {
		r := image.Rect(row[0], row[1], row[0]+row[2], row[1]+row[3])
		if !r.In(image.Rect(1, 1, 239, 129)) {
			t.Fatalf("row escapes list: %v", r)
		}
		if i > 0 && row[1] != g.DropdownRows[i-1][1]+g.DropdownRows[i-1][3] {
			t.Fatal("dropdown rows overlap or leave unplanned gaps")
		}
	}
	if g.InvalidInputContent[0]+g.InvalidInputContent[2] >= 208 {
		t.Fatal("editable input content overlaps validation symbol")
	}
}
