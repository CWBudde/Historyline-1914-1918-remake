package interfaceart

import (
	"fmt"
	"math"
)

var (
	states = []string{"normal", "hover", "focus", "pressed", "disabled"}
	owners = []struct{ ID, Color string }{{"neutral", "#929b9b"}, {"cp", "#b49579"}, {"ec", "#acb5c2"}}
)

func star(c string, x, y, r float64) Shape {
	var p []float64
	for i := 0; i < 10; i++ {
		a := float64(i)*math.Pi/5 - math.Pi/2
		rr := r
		if i%2 == 1 {
			rr *= .43
		}
		p = append(p, x+rr*math.Cos(a), y+rr*math.Sin(a))
	}
	return poly(c, p...)
}

func glyph(id, c string) []Shape {
	switch id {
	case "move":
		return []Shape{poly(c, 3, 10, 14, 10, 14, 4, 23, 12, 14, 20, 14, 14, 3, 14)}
	case "attack":
		return []Shape{poly(c, 4, 20, 17, 3, 23, 1, 22, 7, 8, 23), line(c, 3, 13, 14, 22, 2)}
	case "contain":
		return []Shape{poly(c, 2, 7, 12, 3, 22, 7, 12, 12), poly(c, 2, 9, 11, 14, 11, 23, 2, 18), poly(c, 13, 14, 22, 9, 22, 18, 13, 23)}
	case "exit", "invalid":
		return []Shape{line(c, 5, 5, 19, 19, 3), line(c, 19, 5, 5, 19, 3)}
	case "capture":
		return []Shape{box(c, 4, 2, 2, 21), poly(c, 6, 3, 21, 3, 17, 8, 21, 13, 6, 13)}
	case "fuel":
		return []Shape{round(c, 3, 3, 14, 20, 2), box(slate, 6, 6, 8, 6), poly(c, 17, 6, 21, 8, 23, 12, 23, 22, 19, 22, 19, 13, 17, 12)}
	case "ammo":
		return []Shape{poly(c, 7, 3, 10, 0, 13, 3, 13, 20, 7, 20), box(c, 5, 21, 10, 2), poly(c, 16, 7, 19, 4, 22, 7, 22, 20, 16, 20)}
	case "infinity":
		return []Shape{ring(c, 7, 12, 5, 2), ring(c, 17, 12, 5, 2), line(c, 9, 8, 15, 16, 2), line(c, 9, 16, 15, 8, 2)}
	case "empty":
		return []Shape{box(c, 5, 11, 14, 2)}
	case "skill-unknown":
		return []Shape{poly(c, 12, 2, 22, 12, 12, 22, 2, 12), poly(slate, 12, 6, 18, 12, 12, 18, 6, 12)}
	case "sight":
		return []Shape{poly(c, 1, 12, 6, 7, 12, 5, 18, 7, 23, 12, 18, 17, 12, 19, 6, 17), poly(slate, 4, 12, 8, 9, 12, 8, 16, 9, 20, 12, 16, 15, 12, 16, 8, 15), disk(c, 12, 12, 3)}
	case "size":
		return []Shape{disk(c, 12, 5, 3), round(c, 7, 10, 10, 12, 3), disk(c, 4, 9, 2), box(c, 1, 13, 4, 8), disk(c, 20, 9, 2), box(c, 19, 13, 4, 8)}
	case "armour":
		return []Shape{poly(c, 3, 3, 12, 1, 21, 3, 20, 14, 17, 19, 12, 23, 7, 19, 4, 14), poly(slate, 6, 5, 12, 4, 18, 5, 17, 13, 15, 17, 12, 19, 9, 17, 7, 13)}
	case "weight":
		return []Shape{round(c, 3, 8, 18, 14, 3), ring(c, 12, 7, 5, 2)}
	case "air":
		return []Shape{poly(c, 11, 1, 13, 1, 14, 9, 23, 14, 23, 17, 14, 14, 14, 20, 18, 23, 6, 23, 10, 20, 10, 14, 1, 17, 1, 14, 10, 9)}
	case "ground":
		return []Shape{poly(c, 1, 21, 7, 9, 12, 15, 18, 5, 23, 21)}
	case "water":
		return []Shape{poly(c, 1, 12, 23, 12, 19, 19, 6, 19), box(c, 10, 4, 2, 8), poly(c, 12, 4, 19, 9, 12, 9), line(c, 2, 22, 8, 22, 1), line(c, 12, 22, 21, 22, 1)}
	case "repair":
		return []Shape{poly(c, 3, 2, 9, 2, 9, 9, 16, 16, 21, 16, 23, 21, 20, 24, 15, 22, 15, 17, 8, 10, 2, 10), disk(slate, 19, 20, 1)}
	case "hit":
		return []Shape{disk(c, 12, 9, 9), round(c, 7, 12, 10, 10, 2), disk(slate, 8, 9, 3), disk(slate, 16, 9, 3), poly(slate, 12, 12, 10, 16, 14, 16), box(slate, 9, 18, 1, 5), box(slate, 14, 18, 1, 5)}
	case "turn":
		return []Shape{poly(c, 3, 7, 17, 7, 17, 3, 23, 9, 17, 15, 17, 11, 7, 11, 7, 15, 1, 9), poly(c, 21, 17, 7, 17, 7, 21, 1, 15, 7, 9, 7, 13, 17, 13, 17, 9, 23, 15)}
	}
	panic("unknown glyph " + id)
}

func hexOutline(c string, inset, w float64) []Shape {
	pts := [][2]float64{{30 + inset, 10 + inset}, {90 - inset, 10 + inset}, {114 - inset, 50}, {90 - inset, 90 - inset}, {30 + inset, 90 - inset}, {6 + inset, 50}}
	var s []Shape
	for i, p := range pts {
		q := pts[(i+1)%6]
		s = append(s, line(c, p[0], p[1], q[0], q[1], w))
	}
	return s
}

func controlFrame(w, h int, state string) []Shape {
	edge, fill := muted, slate
	switch state {
	case "hover":
		edge, fill = brass, "#243743"
	case "focus":
		edge, fill = cream, "#243743"
	case "pressed":
		edge, fill = brass, "#0d1b24"
	case "disabled":
		edge, fill = "#46545b", "#19252c"
	}
	s := []Shape{round(edge, 0, 0, float64(w), float64(h), 5), round(fill, 1, 1, float64(w-2), float64(h-2), 4)}
	if state == "focus" {
		s = append(s, border(brass, 4, 4, float64(w-8), float64(h-8), 1)...)
	}
	if state == "pressed" {
		s = append(s, box("#080e13", 3, 3, float64(w-6), 2))
	}
	return s
}

func catalog() []Art {
	var out []Art
	var ownershipAtlas []Shape
	add := func(id, family string, w, h int, s ...Shape) {
		out = append(out, Art{id, family, w, h, [2]float64{float64(w) / 2, float64(h) / 2}, s})
	}
	for ownerIndex, o := range owners {
		strip := []Shape{round("#28333b", 0, 0, 35, 6, 1), box(o.Color, 1, 1, 33, 3), box(cream, 1, 1, 33, 1), box("#111e27", 1, 4, 33, 1)}
		add("ownership/"+o.ID, "ownership", 35, 6, strip...)
		add("ownership/"+o.ID+"-factory", "ownership", 24, 34, transform(strip, 1, 1, 6, 0, math.Pi/3)...)
		ownershipAtlas = append(ownershipAtlas, transform(strip, 1, 1, 0, float64(ownerIndex*6), 0)...)
	}
	add("ownership/depotOwnership", "ownership", 35, 18, ownershipAtlas...)
	add("country/germany", "information", 88, 88, round(muted, 2, 2, 84, 84, 4), box("#151719", 5, 5, 78, 26), box(cream, 5, 31, 78, 26), box("#9e403c", 5, 57, 78, 26))
	add("country/france", "information", 88, 88, round(muted, 2, 2, 84, 84, 4), box("#34526e", 5, 5, 26, 78), box(cream, 31, 5, 26, 78), box("#b45248", 57, 5, 26, 78))
	for rank := 0; rank <= 10; rank++ {
		var s []Shape
		if rank >= 5 {
			s = append(s, ring(brass, 44, 44, 30, 2))
			if rank == 10 {
				s = append(s, ring(brass, 44, 44, 25, 1))
			}
			for i := 0; i < 8; i++ {
				a := math.Pi * float64(i) / 7
				for _, side := range []float64{-1, 1} {
					x, y := 44+side*math.Sin(a)*33, 44+math.Cos(a)*33
					s = append(s, transform([]Shape{poly(brass, 0, 0, 8, -4, 10, 0, 3, 3)}, 1, 1, x, y, side*a)...)
				}
			}
		}
		n := rank % 5
		if rank == 0 {
			s = append(s, star(muted, 44, 44, 13))
		} else if n == 0 {
			s = append(s, star(brass, 44, 44, 19))
		} else {
			for j := 0; j < n; j++ {
				x := 44 + float64(2*j-n+1)*9
				s = append(s, star(brass, x, 44, 10))
			}
		}
		add(fmt.Sprintf("rank/%02d", rank), "information", 88, 88, s...)
	}
	for _, id := range []string{"capture", "skill-unknown", "fuel", "ammo", "infinity", "empty", "sight", "size", "armour", "weight", "air", "ground", "water", "repair", "turn"} {
		add("symbol/"+id, "information", 24, 24, glyph(id, brass)...)
	}
	add("bar/track", "information", 300, 20, round(muted, 0, 0, 300, 20, 3), round("#0d1b24", 1, 1, 298, 18, 2))
	for _, o := range []struct{ ID, Color string }{{"positive", "#78aa87"}, {"negative", "#c37263"}, {"neutral", brass}} {
		add("bar/"+o.ID, "information", 15, 26, round("#101f29", 0, 0, 15, 26, 2), box(o.Color, 2, 2, 11, 22), box(cream, 2, 2, 11, 1))
	}
	add("bar/unavailable", "information", 24, 24, glyph("invalid", muted)...)
	for _, id := range []string{"cursor", "selected", "from", "attack", "reachable", "attackable", "invalid", "path", "editor", "camera"} {
		c := brass
		switch id {
		case "reachable", "path", "from":
			c = "#81b49a"
		case "attack", "attackable", "invalid":
			c = "#dc8974"
		case "editor", "camera":
			c = cream
		}
		s := hexOutline("#10202bdc", 0, 6)
		s = append(s, hexOutline(c, 0, 2)...)
		switch id {
		case "from", "path":
			s = append(s, transform(glyph("move", c), 1.6, 1.6, 41, 31, 0)...)
		case "attack":
			s = append(s, transform(glyph("attack", c), 1.6, 1.6, 41, 31, 0)...)
		case "invalid":
			s = append(s, transform(glyph("invalid", c), 1.6, 1.6, 41, 31, 0)...)
		case "reachable":
			s = append(s, disk(c, 60, 50, 4))
		case "attackable":
			s = append(s, ring(c, 60, 50, 10, 2), line(c, 60, 35, 60, 43, 2), line(c, 60, 57, 60, 65, 2))
		case "selected":
			s = append(s, hexOutline(cream, 4, 1)...)
		}
		if id == "camera" {
			s = border("#10202b", 4, 4, 112, 92, 4)
			s = append(s, border(cream, 5, 5, 110, 90, 2)...)
		}
		add("map/"+id, "map", 120, 100, s...)
	}
	for _, state := range states {
		c := brass
		offset := 0.0
		if state == "disabled" {
			c = muted
		}
		if state == "pressed" {
			offset = 1
		}
		for _, id := range []string{"move", "attack", "contain", "exit"} {
			s := controlFrame(62, 52, state)
			s = append(s, transform(glyph(id, c), 1.4, 1.4, 14+offset, 9+offset, 0)...)
			add("action/"+id+"-"+state, "controls", 62, 52, s...)
		}
		s := border(c, 1, 1, 50, 50, 2)
		if state == "focus" {
			s = append(s, border(cream, 5, 5, 42, 42, 1)...)
		}
		if state == "pressed" {
			s = append(s, box("#c4a77328", 4, 4, 44, 44))
		}
		add("slot/"+state, "controls", 52, 52, s...)
		add("weapon-row/"+state, "controls", 654, 50, controlFrame(654, 50, state)...)
		add("button/"+state, "controls", 160, 40, controlFrame(160, 40, state)...)
	}
	frame := func(id string, w, h int) []Shape { return controlFrame(w, h, "normal") }
	terrain := frame("terrain", 654, 110)
	terrain = append(terrain, box(brass, 112, 10, 1, 90), box(brass, 375, 10, 1, 90))
	for _, y := range []float64{18, 48, 78} {
		terrain = append(terrain, round("#0d1b24", 390, y, 246, 20, 2))
	}
	add("panel/terrain", "panels", 654, 110, terrain...)
	battle := frame("battle", 840, 156)
	battle = append(battle, round("#0d1b24", 20, 14, 800, 14, 3), box(brass, 419, 40, 2, 100))
	for _, x := range []float64{20, 284, 468, 732} {
		battle = append(battle, round("#0d1b24", x, 46, 88, 88, 4))
	}
	add("panel/battle", "panels", 840, 156, battle...)
	weapon := frame("weapon", 670, 286)
	for _, y := range []float64{58, 111, 164, 217} {
		weapon = append(weapon, round("#20333f", 8, y, 654, 50, 3))
	}
	add("panel/weapon", "panels", 670, 286, weapon...)
	content := frame("content", 527, 670)
	content = append(content, round("#0d1b24", 290, 56, 210, 188, 4))
	for row := 0; row < 9; row++ {
		for col := 0; col < 2; col++ {
			content = append(content, round("#0d1b24", float64(20+col*58), float64(140+row*56), 52, 52, 3))
		}
	}
	add("panel/content", "panels", 527, 670, content...)
	for _, id := range []string{"capture", "round"} {
		s := frame(id, 480, 112)
		g := "capture"
		if id == "round" {
			g = "turn"
		}
		s = append(s, transform(glyph(g, brass), 2.6, 2.6, 22, 22, 0)...)
		s = append(s, box(brass, 98, 20, 1, 72))
		add("banner/"+id, "panels", 480, 112, s...)
	}
	add("effect/hit", "effects", 72, 72, transform(glyph("hit", cream), 2.5, 2.5, 6, 6, 0)...)
	add("effect/hit-reduced", "effects", 72, 72, ring(brass, 36, 36, 22, 3))
	return out
}
