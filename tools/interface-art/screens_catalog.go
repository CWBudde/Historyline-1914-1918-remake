package main

// Full-screen art uses the same editable path model as the interface catalog.
// Text, examples and controls are fixture data, never baked into runtime art.
type ScreenText struct {
	Text  string `json:"text"`
	X     int    `json:"x"`
	Y     int    `json:"baseline"`
	Size  int    `json:"size"`
	Bold  bool   `json:"bold"`
	Color string `json:"color"`
}
type ScreenRegion struct {
	ID   string `json:"id"`
	Rect [4]int `json:"rect"`
}
type ScreenControl struct {
	Label string `json:"label"`
	Rect  [4]int `json:"rect"`
	State string `json:"state"`
}
type ScreenDecoration struct {
	ArtID string `json:"art_id"`
	Rect  [4]int `json:"rect"`
}
type Screen struct {
	ID          string             `json:"id"`
	InventoryID string             `json:"inventory_id"`
	Title       string             `json:"title"`
	Privacy     bool               `json:"privacy"`
	Regions     []ScreenRegion     `json:"regions"`
	Text        []ScreenText       `json:"live_fixture_text"`
	Controls    []ScreenControl    `json:"fixture_controls"`
	Decorations []ScreenDecoration `json:"fixture_decorations,omitempty"`
}

func screenCatalog() []Screen {
	var all []Screen
	add := func(id, inventory, title string) *Screen {
		all = append(all, Screen{ID: id, InventoryID: inventory, Title: title})
		s := &all[len(all)-1]
		s.Text = append(s.Text, ScreenText{title, 40, 52, 22, true, cream})
		return s
	}
	region := func(s *Screen, id string, x, y, w, h int) {
		s.Regions = append(s.Regions, ScreenRegion{id, [4]int{x, y, w, h}})
	}
	label := func(s *Screen, text string, x, y int) {
		s.Text = append(s.Text, ScreenText{text, x, y, 13, false, cream})
	}
	button := func(s *Screen, text string, x, y, w int, state string) {
		s.Controls = append(s.Controls, ScreenControl{text, [4]int{x, y, w, 30}, state})
	}
	footer := func(s *Screen, primary string) {
		button(s, "Back", 24, 346, 112, "normal")
		if primary != "" {
			button(s, primary, 464, 346, 152, "focus")
		}
	}

	s := add("title", "screen:title", "HISTORYLINE 1914–1918")
	region(s, "menu", 24, 88, 248, 242)
	for i, name := range []string{"Start game", "Continue", "Settings", "Credits and help"} {
		state := "normal"
		if i == 0 {
			state = "focus"
		}
		if i == 1 {
			state = "disabled"
		}
		button(s, name, 40, 110+i*48, 216, state)
	}
	label(s, "A strategy remake", 304, 300)
	label(s, "No saved game selected", 40, 366)

	s = add("side-selection", "screen:side-mission-selection", "Choose a side")
	region(s, "cp-card", 24, 88, 284, 238)
	region(s, "ec-card", 332, 88, 284, 238)
	label(s, "Central Powers", 44, 124)
	label(s, "Entente", 352, 124)
	s.Decorations = []ScreenDecoration{{"country/germany", [4]int{44, 140, 88, 88}}, {"country/france", [4]int{352, 140, 88, 88}}}
	label(s, "Germany", 148, 166)
	label(s, "France", 456, 166)
	label(s, "Country and owner", 148, 204)
	label(s, "are separate", 148, 226)
	button(s, "Selected", 44, 274, 244, "focus")
	button(s, "Choose Entente", 352, 274, 244, "normal")
	footer(s, "Choose mission")

	s = add("mission-selection", "screen:side-mission-selection", "Choose a mission")
	region(s, "mission-list", 24, 88, 248, 238)
	region(s, "details", 288, 88, 328, 238)
	for i, name := range []string{"Pulse · selected", "Civil · available", "Mouse · completed", "Campaign entry · locked"} {
		state := "normal"
		if i == 0 {
			state = "focus"
		}
		if i == 3 {
			state = "disabled"
		}
		button(s, name, 40, 108+i*46, 216, state)
	}
	label(s, "Pulse", 308, 120)
	label(s, "Mission 1 · Central Powers", 308, 148)
	label(s, "Details and map identity", 308, 190)
	label(s, "come from the selected level", 308, 212)
	label(s, "Code entry", 308, 254)
	button(s, "Invalid code", 308, 272, 288, "pressed")
	footer(s, "Read briefing")

	s = add("briefing", "screen:briefing-campaign", "Mission briefing")
	region(s, "narrative", 24, 88, 370, 238)
	region(s, "objectives", 410, 88, 206, 238)
	for i, line := range []string{"Pulse · Mission 1", "Briefing text is supplied by mission data.", "The current level format has no narrative.", "Long briefings use a scrollable text region.", "No historical orders are invented here."} {
		label(s, line, 44, 120+i*30)
	}
	label(s, "Objectives", 430, 120)
	label(s, "Capture the enemy HQ", 430, 156)
	label(s, "Classic / Standard", 430, 204)
	label(s, "rules shown explicitly", 430, 226)
	footer(s, "Begin mission")

	s = add("campaign", "screen:briefing-campaign", "Campaign progression")
	region(s, "progression", 24, 88, 592, 238)
	for i, name := range []string{"Pulse · completed", "Civil · selected", "Mouse · available"} {
		button(s, name, 44, 116+i*58, 252, []string{"normal", "focus", "normal"}[i])
	}
	label(s, "Progression fixture", 332, 128)
	label(s, "The three maps are enabled.", 332, 166)
	label(s, "Unlocks and chronology await", 332, 206)
	label(s, "the campaign content phase.", 332, 228)
	footer(s, "Read briefing")

	s = add("options", "screen:options", "Settings")
	region(s, "settings", 24, 88, 592, 238)
	for i, line := range []string{"Display and UI scale", "Music and sound", "Animation", "Rules profile"} {
		label(s, line, 44, 122+i*50)
	}
	for i, line := range []string{"Windowed · 100%", "Not yet available", "Reduced motion · on", "Classic / Standard"} {
		button(s, line, 324, 102+i*50, 272, []string{"normal", "disabled", "focus", "normal"}[i])
	}
	footer(s, "Apply")

	s = add("pause", "screen:pause", "Paused")
	region(s, "modal", 148, 88, 344, 238)
	for i, name := range []string{"Resume", "Save / load", "Settings", "Exit confirmation…"} {
		button(s, name, 172, 106+i*50, 296, []string{"focus", "disabled", "normal", "normal"}[i])
	}
	footer(s, "")

	for _, id := range []string{"save", "load"} {
		title := "Save game"
		if id == "load" {
			title = "Load game"
		}
		s = add(id, "screen:save-load", title)
		region(s, "slots", 24, 88, 300, 238)
		region(s, "metadata", 340, 88, 276, 238)
		for i, name := range []string{"Slot 1 · current", "Slot 2 · empty", "Slot 3 · incompatible", "Slot 4 · unreadable"} {
			button(s, name, 44, 108+i*46, 260, []string{"focus", "normal", "disabled", "disabled"}[i])
		}
		label(s, "Pulse · round 3", 360, 120)
		label(s, "Central Powers · Standard", 360, 150)
		label(s, "Example metadata only", 360, 194)
		label(s, "No save operation is enabled", 360, 216)
		label(s, "Naming / overwrite / errors", 360, 270)
		label(s, "require the save feature", 360, 292)
		footer(s, title)
	}

	s = add("handover", "screen:hot-seat-handover", "Pass to the next player")
	s.Privacy = true
	region(s, "privacy", 100, 100, 440, 214)
	label(s, "Entente", 132, 154)
	label(s, "The previous map and unit details are hidden.", 132, 198)
	label(s, "Confirm when the next player is ready.", 132, 230)
	button(s, "Ready", 164, 264, 312, "focus")

	for _, id := range []string{"victory", "defeat"} {
		title := "Victory"
		reason := "Enemy headquarters captured"
		primary := "Continue"
		if id == "defeat" {
			title = "Defeat"
			reason = "Your headquarters was captured"
			primary = "Retry"
		}
		s = add(id, "screen:victory-defeat", title)
		region(s, "outcome", 100, 88, 440, 238)
		label(s, reason, 128, 142)
		label(s, "Pulse · round 3 · Standard", 128, 186)
		label(s, "The engine supplies the actual outcome.", 128, 232)
		label(s, "This is a presentation fixture.", 128, 260)
		footer(s, primary)
	}

	s = add("credits", "screen:credits-help", "Credits")
	region(s, "attribution", 24, 88, 592, 238)
	for i, line := range []string{"Historyline remake", "Python reference: PoulKalff", "Go remake: project contributors", "Typeface: Go fonts (licence retained)", "Artwork credits and provenance remain live.", "Release attribution review remains open."} {
		label(s, line, 44, 120+i*32)
	}
	footer(s, "")

	s = add("help", "screen:credits-help", "Controls and help")
	region(s, "help", 24, 88, 592, 238)
	for i, line := range []string{"Keyboard / mouse reference", "F1 · end turn in the Python reference", "Select a unit, then choose a legal action.", "Classic preserves Python compatibility.", "Standard enables documented rule fixes.", "The DOS rules profile is a separate track."} {
		label(s, line, 44, 120+i*32)
	}
	footer(s, "")

	s = add("editor", "screen:map-editor", "Map editor")
	region(s, "palette", 24, 88, 176, 238)
	region(s, "workspace", 216, 88, 400, 238)
	for i, name := range []string{"Terrain", "Infrastructure", "Units", "Containers"} {
		button(s, name, 40, 106+i*46, 144, []string{"focus", "normal", "normal", "normal"}[i])
	}
	label(s, "Map workspace", 236, 124)
	label(s, "Selected: grass · summer", 236, 162)
	label(s, "Compositor and placement preview", 236, 218)
	label(s, "are supplied by the editor runtime.", 236, 240)
	footer(s, "New map")

	s = add("new-map", "screen:map-editor-new-map", "Create a map")
	region(s, "fields", 24, 88, 592, 238)
	for i, name := range []string{"Map name", "Dimensions", "Side / season", "Validation"} {
		label(s, name, 44, 122+i*50)
	}
	for i, name := range []string{"Untitled map", "8 columns · 47 lines", "Central Powers · summer", "Invalid dimensions example"} {
		button(s, name, 292, 102+i*50, 304, []string{"focus", "normal", "normal", "pressed"}[i])
	}
	footer(s, "Create map")
	return all
}

func screenFrame(s Screen) Art {
	a := Art{ID: "screen/" + s.ID, Family: "screen-frame", Width: 640, Height: 400}
	base := "#162630a8"
	if s.Privacy {
		base = slate
	}
	a.Shapes = append(a.Shapes, box(base, 0, 0, 640, 400))
	a.Shapes = append(a.Shapes, border(brass, 8, 8, 624, 384, 1)...)
	a.Shapes = append(a.Shapes, box("#162630ed", 24, 20, 592, 50), box(brass, 24, 70, 592, 2))
	for _, r := range s.Regions {
		x, y, w, h := float64(r.Rect[0]), float64(r.Rect[1]), float64(r.Rect[2]), float64(r.Rect[3])
		a.Shapes = append(a.Shapes, round("#162630ed", x, y, w, h, 5))
		a.Shapes = append(a.Shapes, border(muted, x+1, y+1, w-2, h-2, 1)...)
		a.Shapes = append(a.Shapes, box(brass, x+8, y+8, 22, 2))
	}
	return a
}

func screenControlShapes(c ScreenControl) []Shape {
	r := c.Rect
	return transform(controlFrame(160, 40, c.State), float64(r[2])/160, float64(r[3])/40, float64(r[0]), float64(r[1]), 0)
}
