package dosrules

// hexDistance is the distance helper (0e40:2421 in the demo) on its own
// argument order: two (x, y) pairs of a map whose odd x sit half a cell
// lower, where x is the column. The movement code calls it that way.
func hexDistance(x1, y1, x2, y2 int) int {
	if y1 > y2 {
		x1, x2, y1, y2 = x2, x1, y2, y1
	}
	var lo, hi int
	if x2 >= x1 {
		lo, hi = y1-x1/2, y2-x2/2
	} else {
		x1, x2 = x2, x1
		lo, hi = y1-(x1/2+x1%2), y2-(x2/2+x2%2)
	}
	return x2 - x1 + max(0, hi-lo)
}

// Distance is the battle distance between two cells of a map of the given
// width.
func (p Profile) Distance(width, a, b int) int {
	if p.TransposedDistance {
		return hexDistance(a/width, a%width, b/width, b%width)
	}
	return hexDistance(a%width, a/width, b%width, b/width)
}

// LayoutDistance is the hex distance between two cells of a map of the given
// width on the map's own layout, whatever the profile's battle distance.
func (Profile) LayoutDistance(width, a, b int) int {
	return hexDistance(a%width, a/width, b%width, b/width)
}
