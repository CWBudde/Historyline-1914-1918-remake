package dospanel

import (
	"image"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// Weapon is an attack on one target class: its power and range.
type Weapon struct{ Power, Range int }

// UnitInfo is what the unit info shows: the unit's name, its weapon
// against air, ground and water targets, armour, speed, weight and the full
// group, with the portrait (drawn in sepia) and the map sprite. Either image
// may be nil.
type UnitInfo struct {
	Name                         string
	Air, Ground, Water           Weapon
	Armour, Speed, Weight, Group int
	Portrait, Sprite             *ebiten.Image
}

// UnitInfoRects are the parts of the unit info in device pixels: the steel
// plate holding everything, the photo's recess and the photo, the hex plate
// with the sprite, the name plate and the stats plate, inside Panel. Scale
// is the device pixels per reference pixel the layout uses.
type UnitInfoRects struct {
	Scale                                     float64
	Panel                                     image.Rectangle
	Plate, PhotoWell, Photo, Hex, Name, Stats image.Rectangle
}

// The unit info in reference pixels: a 140×162 block with a 100×84 photo at
// the top left and the hex plate at its lower right.
const (
	unitInfoW, unitInfoH = 140, 162
	unitHexX, unitHexY   = 124, 78 // the hex plate's centre
	unitHexA, unitHexB   = 13, 12  // and half size
)

// UnitInfoLayout lays out the unit info in panel at scale.
func UnitInfoLayout(panel image.Rectangle, scale float64) UnitInfoRects {
	g := fit(panel, scale, unitInfoW, unitInfoH)
	return UnitInfoRects{
		Scale:     g.k,
		Panel:     panel,
		Plate:     g.rect(0, 0, unitInfoW, unitInfoH),
		PhotoWell: g.rect(4, 4, 110, 94),
		Photo:     g.rect(7, 7, 107, 91),
		Hex:       g.rect(unitHexX-unitHexA, unitHexY-unitHexB, unitHexX+unitHexA, unitHexY+unitHexB),
		Name:      g.rect(2, 97, 138, 108),
		Stats:     g.rect(2, 110, 138, 160),
	}
}

// paintUnitInfo paints the plates of the unit info.
func paintUnitInfo(l UnitInfoRects) *image.RGBA {
	img := image.NewRGBA(l.Panel)
	p := painter{img, l.Scale}
	p.steel(img.Rect, toneBack)
	p.plate(l.Plate, tonePlate)
	p.well(l.PhotoWell)
	p.plate(l.Photo, toneBlank)
	cx, cy, a, b := hexOf(l.Hex)
	p.hex(cx, cy, a, b, toneHex, true)
	p.namePlate(l.Name)
	p.namePlate(l.Stats)
	return img
}

// Labels of the unit info.
const (
	labelWeapons = "WEAPONS"
	labelPowRng  = "(POWER/RANGE)"
	labelAir     = "AIR"
	labelGround  = "GROUND"
	labelWater   = "WATER"
	labelArmour  = "ARMOUR"
	labelSpeed   = "SPEED"
	labelWeight  = "WEIGHT"
	labelGroup   = "MAX GROUP SIZE"
)

// DrawUnitInfo draws the unit info of u over panel at scale: the map gives
// way to a steel plate with the sepia photo, the sprite on its hex plate,
// the name plate and the stats.
func DrawUnitInfo(dst *ebiten.Image, panel image.Rectangle, scale float64, u UnitInfo, t Text) {
	size := panel.Size()
	l := UnitInfoLayout(image.Rectangle{Max: size}, scale)
	drawPlate(dst, panel, cachedPlate(unitInfoPlate, size, scale, func() *image.RGBA { return paintUnitInfo(l) }))

	l = UnitInfoLayout(panel, scale)
	drawPhoto(dst, l.Photo.Intersect(panel), u.Portrait)
	drawSprite(dst, l.Hex.Intersect(panel), u.Sprite, 0.8)

	g := fit(panel, scale, unitInfoW, unitInfoH)
	w := writer{clip(dst, panel), t}
	w.centre(u.Name, g.x(unitInfoW/2), g.y(102.5), cream)

	w.left(labelWeapons, g.x(8), g.y(117), cream)
	w.left(labelPowRng, g.x(8), g.y(123), cream)
	for i, r := range []struct {
		label string
		w     Weapon
	}{{labelAir, u.Air}, {labelGround, u.Ground}, {labelWater, u.Water}} {
		y := g.y(116 + 6*float64(i))
		w.left(r.label, g.x(70), y, cream)
		w.right(strconv.Itoa(r.w.Power), g.x(113), y, amber)
		w.centre("/", g.x(117), y, cream)
		w.left(strconv.Itoa(r.w.Range), g.x(121), y, amber)
	}
	for i, r := range []struct {
		label string
		v     int
	}{{labelArmour, u.Armour}, {labelSpeed, u.Speed}, {labelWeight, u.Weight}, {labelGroup, u.Group}} {
		y := g.y(137 + 6*float64(i))
		w.left(r.label, g.x(8), y, cream)
		w.right(strconv.Itoa(r.v), g.x(104), y, amber)
	}
}
