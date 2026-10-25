package dospanel

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// Drawing needs no running game, but the pixels of an ebiten image cannot be
// read back in a plain test. So the draw tests run every Draw function on an
// offscreen image larger than the panel and check what can be seen: every
// text lands on an image clipped to the panel, and the plates, painted on the
// CPU, cover the panel and nothing outside it.

func TestDrawStaysInPanel(t *testing.T) {
	portrait := ebiten.NewImage(64, 48)
	portrait.Fill(color.RGBA{0x80, 0x90, 0xA0, 0xFF})
	sprite := ebiten.NewImage(24, 24)
	sprite.Fill(color.RGBA{0x30, 0x60, 0x30, 0xFF})
	unit := UnitInfo{
		Name: "LIGHT GUN", Ground: Weapon{45, 3}, Water: Weapon{45, 3}, Armour: 25, Speed: 2,
		Weight: 3, Group: 6, Portrait: portrait, Sprite: sprite,
	}
	contents := Contents{
		Name: "DEPOT", Stored: 15, Total: 80, Selected: 1,
		Details: []string{"GROUP 6 OF 6", "READY"},
	}
	contents.Slots[0] = Slot{Occupied: true, Name: "SCOUT PLANE", Size: 1, Full: 1, Sprite: sprite}
	contents.Slots[1] = Slot{Occupied: true, Name: "CAVALRY", Size: 5, Full: 6, Sprite: sprite, Portrait: portrait}
	game := GameInfo{
		Round: 3, Level: "PULSE", Mode: "MOVEMENT", Best: 900, Score: 861,
		Sides: [2]Holdings{{14, 0, 0}, {14, 0, 0}}, Map: Holdings{30, 2, 0}, Income: [2]int{5, 5},
	}

	draws := map[string]func(dst *ebiten.Image, p image.Rectangle, s float64, tx Text){
		"unit info": func(dst *ebiten.Image, p image.Rectangle, s float64, tx Text) {
			DrawUnitInfo(dst, p, s, unit, tx)
		},
		"unit info blank": func(dst *ebiten.Image, p image.Rectangle, s float64, tx Text) {
			DrawUnitInfo(dst, p, s, UnitInfo{}, tx)
		},
		"game info": func(dst *ebiten.Image, p image.Rectangle, s float64, tx Text) {
			DrawGameInfo(dst, p, s, game, tx)
		},
		"contents": func(dst *ebiten.Image, p image.Rectangle, s float64, tx Text) {
			DrawContents(dst, p, s, contents, tx)
		},
		"contents empty": func(dst *ebiten.Image, p image.Rectangle, s float64, tx Text) {
			DrawContents(dst, p, s, Contents{Selected: -1}, tx)
		},
	}
	for _, c := range panelCases {
		dst := ebiten.NewImage(c.panel.Max.X+40, c.panel.Max.Y+40)
		for name, fn := range draws {
			t.Run(c.name+"/"+name, func(t *testing.T) {
				var tx boxText
				fn(dst, c.panel, c.scale, &tx)
				fn(dst, c.panel, c.scale, nil) // no text
				for _, d := range tx.draws {
					if !d.clip.In(c.panel) {
						t.Errorf("%q drawn on %v, outside the panel %v", d.s, d.clip, c.panel)
					}
					if !d.at.Min.In(c.panel) {
						t.Errorf("%q starts at %v, outside the panel %v", d.s, d.at.Min, c.panel)
					}
				}
			})
		}
		dst.Deallocate()
	}
}

func TestDrawText(t *testing.T) {
	panel := image.Rect(0, 0, 304, 360)
	dst := ebiten.NewImage(320, 400)
	tests := []struct {
		name string
		draw func(tx Text)
		want []string
	}{
		{"unit info", func(tx Text) {
			DrawUnitInfo(dst, panel, 2, UnitInfo{Name: "LIGHT GUN", Ground: Weapon{45, 3}, Armour: 25, Group: 6}, tx)
		}, []string{"LIGHT GUN", "AIR", "GROUND", "WATER", "ARMOUR", "SPEED", "WEIGHT", "MAX GROUP SIZE", "45", "25", "6"}},
		{"game info", func(tx Text) {
			DrawGameInfo(dst, panel, 2, GameInfo{
				Level: "PULSE", Mode: "MOVEMENT", Best: 900, Score: 861,
				Map: Holdings{Units: 30},
			}, tx)
		}, []string{
			"ROUND", "PULSE", "MOVEMENT", "900", "861", "ONE", "TWO", "MAP", "UNITS", "DEPOTS",
			"FACTORY UNITS", "RESOURCES TOTAL", "RESOURCES INCOME", "30", "MOVES", "LIMIT", "NO LIMIT",
		}},
		{"game info limit", func(tx Text) {
			DrawGameInfo(dst, panel, 2, GameInfo{Limit: 40, Moves: 7}, tx)
		}, []string{"40", "7"}},
		{"contents", func(tx Text) {
			c := Contents{Name: "DEPOT", Stored: 15, Total: 80, Selected: 0, Details: []string{"READY TO MOVE"}}
			c.Slots[0] = Slot{Occupied: true, Name: "CAVALRY", Size: 5, Full: 6}
			DrawContents(dst, panel, 2, c, tx)
		}, []string{"DEPOT", "RESOURCES", "TOTAL", "15", "80", "CAVALRY", "5", "READY TO MOVE"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tx boxText
			tt.draw(&tx)
			for _, s := range tt.want {
				if !tx.has(s) {
					t.Errorf("%q not drawn", s)
				}
			}
		})
	}
	var tx boxText
	DrawGameInfo(dst, panel, 2, GameInfo{Limit: 40}, &tx)
	if tx.has("NO LIMIT") {
		t.Error("NO LIMIT drawn with a limit")
	}
}

// The plates are painted on the CPU: they cover the whole panel, opaque,
// and composed onto a screen leave everything outside the panel clear.
func TestPlatesCoverPanel(t *testing.T) {
	for _, c := range panelCases {
		size := image.Rectangle{Max: c.panel.Size()}
		plates := map[string]*image.RGBA{
			"unit info": paintUnitInfo(UnitInfoLayout(size, c.scale)),
			"contents":  paintContents(ContentsLayout(size, c.scale)),
		}
		for name, img := range plates {
			if img.Bounds() != size {
				t.Errorf("%s/%s: bounds %v, want %v", c.name, name, img.Bounds(), size)
				continue
			}
			for i := 3; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 0xFF {
					t.Fatalf("%s/%s: pixel %d not opaque", c.name, name, i/4)
				}
			}
			screen := image.NewRGBA(image.Rect(0, 0, c.panel.Max.X+40, c.panel.Max.Y+40))
			draw.Draw(screen, c.panel, img, image.Point{}, draw.Over)
			if a := screen.RGBAAt(c.panel.Max.X+10, c.panel.Max.Y+10).A; a != 0 {
				t.Errorf("%s/%s: pixel outside the panel painted", c.name, name)
			}
			if a := screen.RGBAAt(c.panel.Min.X, c.panel.Min.Y).A; a != 0xFF {
				t.Errorf("%s/%s: panel corner not painted", c.name, name)
			}
		}
	}
}

// Name plates are brown, the steel around them olive-grey.
func TestPlateColours(t *testing.T) {
	l := UnitInfoLayout(image.Rect(0, 0, 304, 360), 2)
	img := paintUnitInfo(l)
	x, y := centre(l.Name)
	if c := img.RGBAAt(x, y); int(c.R)-int(c.B) < 12 {
		t.Errorf("name plate %v is not brown", c)
	}
	if c := img.RGBAAt(l.Plate.Min.X-2, l.Plate.Min.Y-2); absInt(int(c.R)-int(c.B)) > 16 {
		t.Errorf("steel %v is not olive-grey", c)
	}
}

func TestSepia(t *testing.T) {
	m := sepia()
	for _, c := range []color.RGBA{{0x80, 0x80, 0x80, 0xFF}, {0x20, 0x60, 0xE0, 0xFF}, {0xE0, 0x40, 0x10, 0xFF}} {
		r, g, b, _ := m.Apply(c).RGBA()
		if r < g || g < b || r == b {
			t.Errorf("sepia of %v is %v %v %v", c, r>>8, g>>8, b>>8)
		}
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
