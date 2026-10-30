package main

import (
	"cmp"
	"fmt"
	"image/color"
	"math"
	"slices"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/campaign"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/hajimehoshi/ebiten/v2"
)

// Before a campaign battle the player reads the month's front page of "The
// War Chronicle" (campaign.Articles): masthead, headline, an illustration
// (news/<month> of the screen catalog) with its caption and the battle's
// orders under it, the article in two columns and a column of short items; then, where the map gives their side
// unit types it has not had before, the page of new units, four a page.

var (
	paperTone = color.RGBA{232, 223, 202, 255}
	ink       = color.RGBA{38, 32, 26, 255}
	inkFaint  = color.RGBA{98, 88, 76, 255}
	shadow    = color.RGBA{0, 0, 0, 110}
)

// The front page's frame, in logical canvas pixels.
var (
	newsSheet   = rect{12, 6, 616, 388}
	newsPicture = rect{24, 116, 240, 160}
	newsButton  = rect{24, 354, 240, 26}
)

const (
	newsLeft, newsRight = 24, 616
	newsColB, newsColC  = 276, 452 // the article's columns, newsColW wide
	newsColW            = 164
	newsTop             = 125  // the columns' first baseline
	newsBottom          = 382  // the columns' last baseline at most
	newsOrders          = 305  // the rule over the orders, under the caption
	newsLead            = 12.5 // the article's line spacing
	briefLead           = 10.5
)

// printLine is a line of the front page: k.s in k's face and colour with
// its baseline at (x, y), aligned as drawKey aligns.
type printLine struct {
	k        textKey
	x, y, al float64
}

// newsPage is the laid-out front page.
type newsPage struct {
	lines     []printLine
	briefRule float64 // y of the rule over "In brief"
	fits      bool    // false when a part runs out of its room
}

// issue is the paper's number of month: its place among the front pages.
func issue(month string) int {
	months := make([]string, 0, len(campaign.Articles))
	for m := range campaign.Articles {
		months = append(months, m)
	}
	slices.Sort(months)
	return slices.Index(months, month) + 1
}

// layoutNews sets article a as the front page of issue n, with the
// battle's orders under the picture.
func layoutNews(a campaign.Article, n int, orders string) newsPage {
	p := newsPage{fits: true}
	add := func(k textKey, x, y, al float64) { p.lines = append(p.lines, printLine{k, x, y, al}) }
	centre := float64(newsLeft+newsRight) / 2
	width := float64(newsRight - newsLeft)
	// The masthead and the date line between its rules.
	add(textKey{s: "THE WAR CHRONICLE", serif: true, size: 27, c: ink}, centre, 38, 0.5)
	date := textKey{print: printRoman, size: 8.5, c: inkFaint}
	date.s = fmt.Sprintf("No. %d", n)
	add(date, newsLeft, 56, 0)
	date.s = strings.ToUpper(a.Label)
	add(date, centre, 56, 0.5)
	date.s = "PRICE ONE PENNY"
	add(date, newsRight, 56, 1)
	head := textKey{s: a.Headline, serif: true, size: 21, c: ink}
	deck := textKey{s: a.Deck, print: printItalic, size: 11, c: ink}
	p.fits = measure(head) <= width && measure(deck) <= width
	add(head, centre, 86, 0.5)
	add(deck, centre, 102, 0.5)
	// The caption under the picture, then the orders above the button.
	capt := textKey{print: printItalic, size: 8.5, c: inkFaint}
	captions := wrapText(a.Caption, capt, float64(newsPicture.W))
	for i, l := range captions {
		capt.s = l
		add(capt, newsLeft, float64(newsPicture.Y+newsPicture.H)+11+float64(i)*briefLead, 0)
	}
	p.fits = p.fits && len(captions) <= 2
	add(textKey{s: "ORDERS OF THE DAY", bold: true, size: 8, c: ink}, newsLeft, newsOrders+11, 0)
	ord := textKey{print: printRoman, size: 9, c: ink}
	for i, l := range wrapText(orders, ord, float64(newsPicture.W)) {
		ord.s = l
		y := newsOrders + 23 + float64(i)*11
		p.fits = p.fits && y <= float64(newsButton.Y)-6
		add(ord, newsLeft, y, 0)
	}
	// "In brief" at the foot of the last column, bullets hanging.
	item := textKey{print: printRoman, size: 8.5, c: ink}
	var briefs [][]string
	rows := 0
	for _, b := range a.Briefs {
		w := wrapText(b, item, newsColW-8)
		briefs = append(briefs, w)
		rows += len(w)
	}
	y := newsBottom - float64(rows-1)*briefLead - 3*float64(len(briefs)-1)
	heading := y - 14
	p.briefRule = heading - 11
	add(textKey{s: "IN BRIEF", bold: true, size: 8, c: ink}, newsColC, heading, 0)
	for _, w := range briefs {
		add(textKey{s: "•", print: printRoman, size: 8.5, c: ink}, newsColC, y, 0)
		for _, l := range w {
			item.s = l
			add(item, newsColC+8, y, 0)
			y += briefLead
		}
		y += 3
	}
	// The article, down the middle column and on into the last one above
	// "In brief"; paragraphs after the first indented.
	body := textKey{print: printRoman, size: 10, c: ink}
	x, y, last := float64(newsColB), float64(newsTop), float64(newsBottom)
	for i, para := range a.Body {
		indent := 0.0
		if i > 0 {
			indent = 9
		}
		for j, l := range wrapIndent(para, body, newsColW, indent) {
			if y > last {
				if x == newsColC {
					p.fits = false
					return p
				}
				x, y, last = newsColC, newsTop, p.briefRule-8
			}
			body.s = l
			if j == 0 {
				add(body, x+indent, y, 0)
			} else {
				add(body, x, y, 0)
			}
			y += newsLead
		}
	}
	return p
}

// drawNewsPage draws the month's front page and its button.
func (a *app) drawNewsPage(dst *ebiten.Image, d int) {
	s := a.shell
	m := s.maps[s.code]
	art, ok := m.news()
	if !ok {
		return
	}
	fillRect(dst, d, rect{newsSheet.X + 3, newsSheet.Y + 3, newsSheet.W, newsSheet.H}, shadow)
	a.drawPaper(dst, d, newsSheet)
	p := layoutNews(art, issue(m.month), m.briefing)
	hr := func(y, h int) { fillRect(dst, d, rect{newsLeft, y, newsRight - newsLeft, h}, ink) }
	hr(44, 2)
	hr(60, 1)
	hr(62, 1)
	hr(109, 1)
	bottom := newsBottom + 2 - newsPicture.Y
	fillRect(dst, d, rect{newsColB - 6, newsPicture.Y, 1, bottom}, inkFaint)
	fillRect(dst, d, rect{newsColC - 6, newsPicture.Y, 1, bottom}, inkFaint)
	fillRect(dst, d, rect{newsColC, int(math.Round(p.briefRule)), newsColW, 1}, ink)
	fillRect(dst, d, rect{newsLeft, newsOrders, newsPicture.W, 1}, ink)
	if img := a.art.image("news/"+m.month, d); img != nil {
		drawInto(dst, d, img, newsPicture)
	} else {
		fillRect(dst, d, newsPicture, color.RGBA{205, 196, 176, 255})
	}
	strokeRect(dst, d, newsPicture, ink)
	for _, l := range p.lines {
		drawKey(dst, d, l.k, l.x, l.y, l.al)
	}
	for i, b := range s.buttons() {
		a.drawButton(dst, d, b, i == s.focus, i == a.press)
	}
}

// drawPaper covers r with the newsprint surface, or its tone without art.
func (a *app) drawPaper(dst *ebiten.Image, d int, r rect) {
	tile := a.art.image("surface/newsprint", d)
	if tile == nil {
		fillRect(dst, d, r, paperTone)
		return
	}
	sheet := dst.SubImage(devRect(r, d)).(*ebiten.Image)
	tw, th := tile.Bounds().Dx(), tile.Bounds().Dy()
	for y := r.Y * d; y < (r.Y+r.H)*d; y += th {
		for x := r.X * d; x < (r.X+r.W)*d; x += tw {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			sheet.DrawImage(tile, op)
		}
	}
}

// drawInto draws img stretched over r.
func drawInto(dst *ebiten.Image, d int, img *ebiten.Image, r rect) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(float64(r.W*d)/float64(w), float64(r.H*d)/float64(h))
	op.GeoM.Translate(float64(r.X*d), float64(r.Y*d))
	dst.DrawImage(img, op)
}

// unitCard is a unit type on the page of new units.
type unitCard struct {
	id, family    string
	name, country string
	armour, speed int
	attack        string // "Ground 30 · Water 30"; "No weapons" for none
	reach         string // "Range 1–3"; "" for none
	note          string // campaign.UnitNotes
}

// unitsPerPage is how many cards the page of new units shows at once.
const unitsPerPage = 4

// newCard is the card of unit type id.
func newCard(set *data.Set, id string) unitCard {
	u := set.Units[id]
	family, _ := unitsprite.Family(id)
	c := unitCard{id: id, family: family, name: u.Name, country: u.Country, armour: u.Armour, speed: u.Speed, note: campaign.UnitNotes[family]}
	var power [3]int // air, ground, water
	lo, hi := math.MaxInt, 0
	for _, sl := range u.Weapons {
		if sl == nil || sl.Attack == nil || sl.Attack.Power == 0 {
			continue
		}
		at := sl.Attack
		for i, hits := range []bool{at.Air, at.Ground, at.Water} {
			if hits {
				power[i] = max(power[i], at.Power)
			}
		}
		lo, hi = min(lo, at.RangeMin), max(hi, at.RangeMax)
	}
	var parts []string
	for i, name := range []string{"Air", "Ground", "Water"} {
		if power[i] > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", name, power[i]))
		}
	}
	switch {
	case len(parts) == 0:
		c.attack = "No weapons"
	case lo == hi:
		c.attack, c.reach = strings.Join(parts, " · "), fmt.Sprintf("Range %d", lo)
	default:
		c.attack, c.reach = strings.Join(parts, " · "), fmt.Sprintf("Range %d–%d", lo, hi)
	}
	return c
}

// sideUnits are the unit types of mode's own side in setup s: on the map,
// in its buildings and among the builds.
func sideUnits(s *scenario.Setup, mode scenario.Mode) []string {
	prefix := "CP_"
	if mode == scenario.EC {
		prefix = "EC_"
	}
	var ids []string
	add := func(id string) {
		if strings.HasPrefix(id, prefix) && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	for _, u := range s.Units {
		add(u.ID)
	}
	for _, b := range s.Buildings {
		for _, id := range b.Units {
			add(id)
		}
	}
	for _, id := range s.Builds {
		add(id)
	}
	return ids
}

// newCards are the cards of ids that seen does not hold yet, in the unit
// table's order; it adds them to seen.
func newCards(set *data.Set, ids []string, seen map[string]bool) []unitCard {
	var cards []unitCard
	for _, id := range ids {
		if _, ok := set.Units[id]; ok && !seen[id] {
			seen[id] = true
			cards = append(cards, newCard(set, id))
		}
	}
	slices.SortFunc(cards, func(a, b unitCard) int {
		return cmp.Or(cmp.Compare(set.Units[a.id].Dos, set.Units[b.id].Dos), cmp.Compare(a.id, b.id))
	})
	return cards
}

// unitPages is the number of pages the cards fill.
func unitPages(cards []unitCard) int {
	return (len(cards) + unitsPerPage - 1) / unitsPerPage
}

// unitCardRect is card i's frame on the page.
func unitCardRect(i int) rect { return rect{24, 86 + i*76, 592, 70} }

// The page of new units' buttons, top right beside the title.
var (
	unitsNextRect   = rect{392, 28, 108, 26}
	unitsBattleRect = rect{508, 28, 108, 26}
)

// drawUnitsPage draws the current page of the map's new units.
func (a *app) drawUnitsPage(dst *ebiten.Image, d int) {
	s := a.shell
	m := s.maps[s.code]
	drawDisplay(dst, d, "NEW AT THE FRONT", 23, 24, 50, amber, false)
	fillRect(dst, d, rect{24, 60, 592, 1}, brass)
	label := m.month
	if art, ok := m.news(); ok {
		label = art.Label
	}
	drawText(dst, d, factionName(m.mode.Human())+" · "+label, false, 11, 24, 76, muted, false)
	if n := unitPages(m.newUnits); n > 1 {
		drawKey(dst, d, textKey{s: fmt.Sprintf("Page %d of %d", s.unitPage+1, n), size: 11, c: muted}, 616, 76, 1)
	}
	first := s.unitPage * unitsPerPage
	for i, c := range m.newUnits[first:min(first+unitsPerPage, len(m.newUnits))] {
		r := unitCardRect(i)
		dospanel.SteelWell(dst, devRect(r, d), float64(d))
		portrait := rect{r.X + 4, r.Y + 4, 69, 62}
		if img := a.art.image("portrait/"+c.family, d); img != nil {
			drawInto(dst, d, img, portrait)
		} else {
			strokeRect(dst, d, portrait, muted)
		}
		x, y := float64(r.X+84), float64(r.Y)
		drawText(dst, d, c.name, true, 11, x, y+15, amber, false)
		drawText(dst, d, c.country, false, 9.5, x, y+27, muted, false)
		drawText(dst, d, fmt.Sprintf("Armour %d · Speed %d", c.armour, c.speed), false, 9.5, x, y+39, cream, false)
		drawText(dst, d, c.attack, false, 9.5, x, y+51, cream, false)
		drawText(dst, d, c.reach, false, 9.5, x, y+63, cream, false)
		for j, l := range wrapText(c.note, textKey{size: 10}, unitNoteW) {
			drawText(dst, d, l, false, 10, float64(r.X+unitNoteX), y+17+float64(j)*13, cream, false)
		}
	}
	for i, b := range s.buttons() {
		a.drawButton(dst, d, b, i == s.focus, i == a.press)
	}
}

// The note's column on a unit card: from unitNoteX within the card,
// unitNoteW wide, four lines at most.
const unitNoteX, unitNoteW = 240, 344
