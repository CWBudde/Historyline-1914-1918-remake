package main

import (
	"errors"
	"fmt"
	"html"
	"image"
	"image/draw"
	"strings"
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type ScreenScroll struct {
	Rect       [4]int `json:"rect"`
	Text       string `json:"full_text"`
	LineHeight int    `json:"line_height"`
	Offset     int    `json:"line_offset"`
	SingleLine bool   `json:"single_line,omitempty"`
}

func screenStateComponents() []Art {
	return []Art{
		{ID: "progress/track", Family: "screen-component", Width: 280, Height: 16, Shapes: []Shape{round(muted, 0, 0, 280, 16, 3), round("#0d1b24", 1, 1, 278, 14, 2)}},
		{ID: "progress/fill", Family: "screen-component", Width: 280, Height: 16, Shapes: []Shape{box(brass, 1, 3, 278, 10), box(cream, 1, 3, 278, 1)}},
		{ID: "scrollbar/track", Family: "screen-component", Width: 14, Height: 210, Shapes: []Shape{round("#0d1b24", 0, 0, 14, 210, 3)}},
		{ID: "scrollbar/thumb", Family: "screen-component", Width: 14, Height: 64, Shapes: []Shape{round(brass, 2, 1, 10, 62, 3), box(cream, 4, 4, 1, 56)}},
	}
}

func screenDecorationSources() []Art {
	all := catalog()
	all = append(all, widgetCatalog()...)
	return append(all, screenStateComponents()...)
}

func renderClippedDecoration(im *image.NRGBA, d ScreenDecoration) error {
	r := d.Rect
	c := [4]int{0, 0, r[2], r[3]}
	if d.Clip != nil {
		c = *d.Clip
	}
	if r[2] <= 0 || r[3] <= 0 || c[0] < 0 || c[1] < 0 || c[2] < 0 || c[3] < 0 || c[0]+c[2] > r[2] || c[1]+c[3] > r[3] {
		return errors.New("invalid decoration clip")
	}
	for _, source := range screenDecorationSources() {
		if source.ID == d.ArtID {
			layer := Art{Width: r[2], Height: r[3], Shapes: transform(source.Shapes, float64(r[2])/float64(source.Width), float64(r[3])/float64(source.Height), 0, 0, 0)}
			clip := image.Rect(c[0]*2, c[1]*2, (c[0]+c[2])*2, (c[1]+c[3])*2)
			draw.Draw(im, clip.Add(image.Pt(r[0]*2, r[1]*2)), layer.raster(2), clip.Min, draw.Over)
			return nil
		}
	}
	return fmt.Errorf("missing decoration %s", d.ArtID)
}

func screenStateCatalog() []Screen {
	var all []Screen
	add := func(id, inventory, title string) *Screen {
		all = append(all, Screen{ID: id, InventoryID: inventory, Title: title, Regions: []ScreenRegion{{"body", [4]int{100, 92, 440, 230}}}, Text: []ScreenText{{title, 40, 52, 22, true, cream}}})
		return &all[len(all)-1]
	}
	line := func(s *Screen, text string, x, y int) {
		s.Text = append(s.Text, ScreenText{text, x, y, 13, false, cream})
	}
	button := func(s *Screen, label string, x, y, w int, state string) {
		s.Controls = append(s.Controls, ScreenControl{label, [4]int{x, y, w, 30}, state})
	}
	dec := func(s *Screen, id string, x, y, w, h int) {
		s.Decorations = append(s.Decorations, ScreenDecoration{ArtID: id, Rect: [4]int{x, y, w, h}})
	}
	for _, pct := range []int{0, 50, 100} {
		s := add(fmt.Sprintf("loading-%03d", pct), "screen:title", "Loading mission")
		line(s, "Pulse", 132, 144)
		line(s, fmt.Sprintf("%d%%", pct), 132, 216)
		dec(s, "progress/track", 132, 174, 280, 16)
		dec(s, "progress/fill", 132, 174, 280, 16)
		s.Decorations[len(s.Decorations)-1].Clip = &[4]int{0, 0, 280 * pct / 100, 16}
		button(s, "Cancel", 132, 266, 112, "normal")
	}
	for _, v := range []struct{ id, inventory, title, message, detail string }{
		{"mission-error", "screen:title", "Unable to load mission", "The selected mission could not be read.", "Choose another mission or try again."},
		{"save-error", "screen:save-load", "Unable to save", "The game could not be saved.", "Choose another slot or try again."},
		{"load-corrupt", "screen:save-load", "Unable to load save", "This saved game cannot be opened.", "Choose another saved game."},
	} {
		s := add(v.id, v.inventory, v.title)
		line(s, v.message, 132, 150)
		line(s, v.detail, 132, 186)
		button(s, "Back", 132, 266, 112, "normal")
		button(s, "Try again", 360, 266, 148, "focus")
	}
	s := add("save-empty", "screen:save-load", "No saved games")
	line(s, "Your saved games will appear here.", 132, 154)
	line(s, "Start a mission to begin.", 132, 190)
	button(s, "Back", 132, 266, 112, "normal")
	button(s, "Start game", 360, 266, 148, "focus")
	s = add("save-progress", "screen:save-load", "Saving game")
	line(s, "Pulse · Slot 1", 132, 144)
	line(s, "Saving…", 132, 216)
	dec(s, "progress/track", 132, 174, 280, 16)
	dec(s, "progress/fill", 132, 174, 280, 16)
	s.Decorations[len(s.Decorations)-1].Clip = &[4]int{0, 0, 140, 16}
	button(s, "Please wait", 360, 266, 148, "disabled")
	for _, v := range []struct{ id, title, body, action string }{
		{"confirm-exit", "Leave this mission?", "Unsaved progress will be lost.", "Leave mission"},
		{"confirm-restart", "Restart this mission?", "Current progress will be lost.", "Restart"},
		{"confirm-overwrite", "Replace saved game?", "The existing save will be replaced.", "Replace save"},
		{"confirm-delete", "Delete saved game?", "This saved game will be removed.", "Delete save"},
	} {
		s = add(v.id, "screen:pause", v.title)
		if strings.Contains(v.id, "overwrite") || strings.Contains(v.id, "delete") {
			s.InventoryID = "screen:save-load"
		}
		line(s, v.body, 132, 160)
		line(s, "Confirm to continue.", 132, 198)
		button(s, "Cancel", 132, 266, 112, "focus")
		button(s, v.action, 348, 266, 160, "normal")
	}
	s = add("settings-controls", "screen:options", "Settings")
	s.Regions = []ScreenRegion{{"settings", [4]int{24, 88, 592, 238}}}
	line(s, "Player name", 44, 122)
	line(s, "Commander", 336, 122)
	dec(s, "widget/input/focus", 324, 96, 240, 40)
	line(s, "Reduced animation", 44, 174)
	line(s, "On", 376, 174)
	dec(s, "widget/toggle/on/normal", 324, 148, 40, 40)
	line(s, "Sound volume", 44, 226)
	dec(s, "widget/slider/track/normal", 324, 200, 220, 32)
	dec(s, "widget/slider/fill/normal", 324, 200, 220, 32)
	s.Decorations[len(s.Decorations)-1].Clip = &[4]int{0, 0, 110, 32}
	dec(s, "widget/slider/thumb/normal", 422, 200, 24, 32)
	line(s, "Rules profile", 44, 278)
	line(s, "Standard", 336, 278)
	dec(s, "widget/dropdown/closed/normal", 324, 252, 240, 40)
	button(s, "Back", 24, 346, 112, "normal")
	button(s, "Apply", 464, 346, 152, "focus")
	s = add("code-invalid", "screen:side-mission-selection", "Enter mission code")
	line(s, "Mission code", 132, 144)
	line(s, "PULSE-?", 144, 186)
	dec(s, "widget/input/invalid/focus", 132, 160, 240, 40)
	line(s, "Code not recognised. Check and try again.", 132, 232)
	button(s, "Back", 132, 266, 112, "normal")
	button(s, "Confirm", 360, 266, 148, "disabled")
	s = add("dropdown-open", "screen:options", "Choose a rules profile")
	dec(s, "widget/dropdown/open/focus", 132, 110, 240, 40)
	line(s, "Standard", 144, 136)
	dec(s, "widget/dropdown-list", 132, 150, 240, 130)
	for i, v := range []string{"Classic", "Standard", "Original DOS (unavailable)", "Back"} {
		state := "normal/normal"
		if i == 1 {
			state = "selected/focus"
		}
		if i == 2 {
			state = "normal/disabled"
		}
		dec(s, "widget/dropdown-option/"+state, 133, 151+i*32, 238, 32)
		line(s, v, 169, 173+i*32)
		if i == 2 {
			s.Text[len(s.Text)-1].Color = muted
		}
	}
	for _, v := range []struct{ id, inventory, title, body string }{
		{"briefing-long", "screen:briefing-campaign", "Mission briefing", "Capture the enemy headquarters.\n\nKeep a clear route between your units and the objective. Check the terrain and available actions before committing to a move.\n\n"},
		{"help-long", "screen:credits-help", "Controls and help", "Select a unit to inspect its information and available actions. The selected rules profile determines which commands are legal.\n\nF1 ends the turn in the Python reference.\n\n"},
		{"credits-long", "screen:credits-help", "Credits", "Historyline remake\nPython reference: PoulKalff\nGo remake: project contributors\nTypeface: Go fonts\n\n"},
	} {
		s = add(v.id, v.inventory, v.title)
		s.Regions = []ScreenRegion{{"scroll", [4]int{24, 88, 592, 238}}}
		s.Scroll = &ScreenScroll{Rect: [4]int{44, 104, 536, 204}, Text: strings.Repeat(v.body, 8), LineHeight: 20}
		dec(s, "scrollbar/track", 586, 102, 14, 210)
		dec(s, "scrollbar/thumb", 586, 102, 14, 64)
		button(s, "Back", 24, 346, 112, "focus")
	}
	s = add("name-overflow", "screen:save-load", "Name this saved game")
	line(s, "Save name", 132, 144)
	dec(s, "widget/input/focus", 132, 160, 240, 40)
	// Full text is retained; the field clips horizontally rather than losing data.
	s.Scroll = &ScreenScroll{Rect: [4]int{144, 168, 216, 24}, Text: "A very long saved-game name — München / Épernay / 1918 / final advance", LineHeight: 24, SingleLine: true}
	button(s, "Cancel", 132, 266, 112, "normal")
	button(s, "Save", 360, 266, 148, "focus")
	s = add("new-map-invalid", "screen:map-editor-new-map", "Create a map")
	line(s, "Map dimensions", 132, 144)
	line(s, "0 columns × 47 lines", 144, 186)
	dec(s, "widget/input/invalid/normal", 132, 160, 240, 40)
	line(s, "Choose a positive map width.", 132, 232)
	button(s, "Cancel", 132, 266, 112, "normal")
	button(s, "Create map", 360, 266, 148, "disabled")
	return all
}

// Word wrapping breaks overlong tokens by Unicode rune. ScreenScroll retains the
// original text; only layout whitespace at a line break is normalized here.
func wrapScreenText(f font.Face, text string, width int) []string {
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		runes := []rune(paragraph)
		for start := 0; start < len(runes); {
			end := start
			for end < len(runes) && font.MeasureString(f, string(runes[start:end+1])) <= fixed.I(width) {
				end++
			}
			if end == start {
				end++
			} // A single glyph can exceed a very narrow clip.
			cut := end
			if end < len(runes) {
				for i := end - 1; i > start; i-- {
					if unicode.IsSpace(runes[i]) {
						cut = i
						break
					}
				}
			}
			lines = append(lines, strings.TrimRightFunc(string(runes[start:cut]), unicode.IsSpace))
			start = cut
			for start < len(runes) && unicode.IsSpace(runes[start]) {
				start++
			}
		}
	}
	return lines
}

func renderScreenScroll(im *image.NRGBA, s ScreenScroll, f *opentype.Font) error {
	if s.Rect[2] <= 0 || s.Rect[3] <= 0 || s.LineHeight < 14 || s.Offset < 0 {
		return errors.New("invalid scroll region")
	}
	r := image.Rect(s.Rect[0]*2, s.Rect[1]*2, (s.Rect[0]+s.Rect[2])*2, (s.Rect[1]+s.Rect[3])*2)
	if !r.In(im.Bounds()) {
		return errors.New("scroll outside screen")
	}
	face, e := opentype.NewFace(f, &opentype.FaceOptions{Size: 26, DPI: 72, Hinting: font.HintingNone})
	if e != nil {
		return e
	}
	defer func() { _ = face.Close() }()
	lines := wrapScreenText(face, s.Text, s.Rect[2]*2)
	if s.SingleLine {
		lines = []string{s.Text}
	}
	d := font.Drawer{Dst: im.SubImage(r).(*image.NRGBA), Src: image.NewUniform(colour(cream)), Face: face}
	for i := s.Offset; i < len(lines); i++ {
		y := r.Min.Y + 26 + (i-s.Offset)*s.LineHeight*2
		if y > r.Max.Y+26 {
			break
		}
		bounds, _ := font.BoundString(face, lines[i])
		if fixed.I(y)+bounds.Max.Y > fixed.I(r.Max.Y) {
			break
		}
		d.Dot = fixed.P(r.Min.X, y)
		d.DrawString(lines[i])
	}
	return nil
}

func screenStateGallery(screens []Screen, arts []Art) string {
	h := screenGallery(screens)
	h = strings.Replace(h, "Sixteen separate views", fmt.Sprintf("%d screen-state studies", len(screens)), 1)
	for _, s := range screens {
		if s.Scroll != nil {
			r := s.Scroll.Rect
			extra := ""
			if s.Scroll.SingleLine {
				extra = "white-space:pre;overflow-wrap:normal;"
			}
			markup := fmt.Sprintf(`<div class="textflow" role="region" tabindex="0" aria-label="%s text" style="left:%dpx;top:%dpx;width:%dpx;height:%dpx;line-height:%dpx;%s">%s</div>`, html.EscapeString(s.Title), r[0], r[1], r[2], r[3], s.Scroll.LineHeight, extra, html.EscapeString(s.Scroll.Text))
			start := strings.Index(h, `<article id="`+s.ID+`"`)
			end := strings.Index(h[start:], `</div></div><details>`) + start
			h = h[:end] + markup + h[end:]
		}
	}
	var b strings.Builder
	b.WriteString(`<style>.textflow{position:absolute;overflow:auto;white-space:pre-wrap;overflow-wrap:anywhere;font:13px Go;color:#eee5d2;scrollbar-color:#c4a773 #162630;scrollbar-width:thin}.textflow:focus-visible{outline:2px solid #eee5d2}.widget-grid{display:flex;flex-wrap:wrap;gap:20px}.widget-grid figure{margin:0;width:260px;padding:12px;background:#162630}.widget-grid img{max-width:240px;height:auto}.widget-grid figcaption{font-size:12px;overflow-wrap:anywhere}.live-widgets{padding:24px;max-width:640px;background:#162630;display:grid;gap:24px}.live-widgets input,.live-widgets select{font:16px Go;color:#eee5d2}.live-widgets input[type=text],.live-widgets select{width:240px;height:40px;border:0;padding:8px 12px;background:transparent url('widget/input/normal-2x.png') center/100% 100% no-repeat}.live-widgets input[type=text]:hover{background-image:url('widget/input/hover-2x.png')}.live-widgets input[type=text]:focus-visible{background-image:url('widget/input/focus-2x.png');outline:2px solid #eee5d2}.live-widgets input[type=text]:disabled{background-image:url('widget/input/disabled-2x.png')}.live-widgets input[type=checkbox]{appearance:none;width:40px;height:40px;background:url('widget/toggle/off/normal-2x.png') center/100% 100% no-repeat;vertical-align:middle}.live-widgets input[type=checkbox]:checked{background-image:url('widget/toggle/on/normal-2x.png')}.live-widgets input[type=checkbox]:focus-visible{outline:2px solid #eee5d2}.live-widgets input[type=range]{appearance:none;width:220px;height:32px;background:url('widget/slider/track/normal-2x.png') center/100% 100% no-repeat}.live-widgets input[type=range]::-webkit-slider-thumb{appearance:none;width:24px;height:32px;background:url('widget/slider/thumb/normal-2x.png') center/100% 100% no-repeat}.live-widgets input[type=range]::-moz-range-thumb{border:0;width:24px;height:32px;background:url('widget/slider/thumb/normal-2x.png') center/100% 100% no-repeat}.live-widgets select{appearance:none;padding-right:44px;background-image:url('widget/dropdown/closed/normal-2x.png')}.live-widgets select option{background:#162630;color:#eee5d2}</style><h2>Native HTML control studies</h2><p>These fields demonstrate text, focus, keyboard, pointer and value presentation only; changing them changes no game state. Browser interaction remains an acceptance check.</p><div class="live-widgets"><label>Player name<br><input type="text" value="Commander" maxlength="100"></label><label>Unavailable value<br><input type="text" value="Unavailable" disabled></label><label><input type="checkbox" checked> Reduced animation</label><label>Sound volume<br><input type="range" min="0" max="100" value="50"></label><label>Rules profile<br><select><option>Classic</option><option selected>Standard</option><option disabled>Original DOS (unavailable)</option></select></label></div><h2>Every widget asset and interaction state</h2><div class="widget-grid">`)
	for _, a := range arts {
		if !strings.HasPrefix(a.ID, "widget/") {
			continue
		}
		fmt.Fprintf(&b, `<figure><img src="%s-2x.png" alt="%s" style="width:%dpx;height:%dpx"><figcaption>%s · <a href="%s.svg">SVG</a> · <a href="%s-1x.png">1×</a> / <a href="%s-4x.png">4×</a></figcaption></figure>`, a.ID, html.EscapeString(a.ID), a.Width, a.Height, html.EscapeString(a.ID), a.ID, a.ID, a.ID)
	}
	b.WriteString("</div>")
	return strings.Replace(h, "</html>", b.String()+"</html>", 1)
}
