package main

// WidgetGeometry describes composition in logical pixels. Artwork deliberately
// contains no labels, editable strings, numeric values or application behavior.
type WidgetGeometry struct {
	InputContent          [4]int   `json:"input_content_xywh"`
	InvalidInputContent   [4]int   `json:"invalid_input_content_xywh"`
	ToggleIndicator       [4]int   `json:"toggle_indicator_xywh"`
	SliderTrack           [4]int   `json:"slider_track_xywh"`
	SliderCenters         [2]int   `json:"slider_thumb_center_x_min_max"`
	SliderCenterY         int      `json:"slider_thumb_center_y"`
	SliderThumbSize       [2]int   `json:"slider_thumb_size"`
	DropdownContent       [4]int   `json:"dropdown_content_xywh"`
	DropdownOptionContent [4]int   `json:"dropdown_option_content_xywh"`
	DropdownRows          [][4]int `json:"dropdown_list_rows_xywh"`
	Composition           []string `json:"composition"`
}

func widgetGeometry() WidgetGeometry {
	return WidgetGeometry{
		InputContent:          [4]int{12, 8, 216, 24},
		InvalidInputContent:   [4]int{12, 8, 188, 24},
		ToggleIndicator:       [4]int{9, 9, 22, 22},
		SliderTrack:           [4]int{12, 12, 196, 8},
		SliderCenters:         [2]int{12, 208},
		SliderCenterY:         16,
		SliderThumbSize:       [2]int{24, 32},
		DropdownContent:       [4]int{12, 8, 188, 24},
		DropdownOptionContent: [4]int{36, 4, 188, 24},
		DropdownRows:          [][4]int{{1, 1, 238, 32}, {1, 33, 238, 32}, {1, 65, 238, 32}, {1, 97, 238, 32}},
		Composition: []string{
			"Place live labels outside the 40x40 toggle; off/on/mixed remain distinct without color.",
			"Input text is clipped to the stated content rectangle; invalid reserves space for an exclamation symbol and underline.",
			"Compose slider track, fill and thumb separately. Clamp value to [0,1]; clip fill at x=12+196*value and draw none at zero.",
			"Place slider thumb anchor at (12+196*value,16). The 24x32 thumb stays wholly within the 220x32 slider at both ends.",
			"Dropdown list is a separate 240x130 surface; overlay four 238x32 option rows at the listed positions. Labels are live text.",
			"Hover, focus, pressed and disabled are artwork states only; validation, selection, editing, pointer and keyboard behavior are supplied by the caller.",
		},
	}
}

func widgetInk(state string) string {
	switch state {
	case "hover":
		return "#ddc494"
	case "focus":
		return cream
	case "pressed":
		return "#9f824f"
	case "disabled":
		return "#52616b"
	default:
		return brass
	}
}

// widgetCatalog is shared by exports and review fixtures. Every semantic value
// has all five interaction states, so focus and disabled never imply a value.
func widgetCatalog() []Art {
	var out []Art
	add := func(id string, w, h int, shapes []Shape) {
		out = append(out, Art{ID: "widget/" + id, Family: "widgets", Width: w, Height: h, Anchor: [2]float64{float64(w) / 2, float64(h) / 2}, Shapes: shapes})
	}
	for _, state := range states {
		ink := widgetInk(state)
		add("input/"+state, 240, 40, controlFrame(240, 40, state))
		invalid := controlFrame(240, 40, state)
		warning := "#dc8974"
		if state == "disabled" {
			warning = "#966b63"
		}
		invalid = append(invalid, box(warning, 12, 36, 216, 2), poly(warning, 218, 9, 228, 29, 208, 29), box(slate, 217, 16, 2, 6), box(slate, 217, 24, 2, 2))
		add("input/invalid/"+state, 240, 40, invalid)

		for _, value := range []string{"off", "on", "mixed"} {
			shapes := controlFrame(40, 40, state)
			shapes = append(shapes, round(ink, 9, 9, 22, 22, 3))
			switch value {
			case "off":
				shapes = append(shapes, round("#0d1b24", 11, 11, 18, 18, 2))
			case "on":
				shapes = append(shapes, line(slate, 13, 20, 18, 25, 3), line(slate, 18, 25, 27, 15, 3))
			case "mixed":
				shapes = append(shapes, box(slate, 13, 18, 14, 4))
			}
			add("toggle/"+value+"/"+state, 40, 40, shapes)
		}

		track := controlFrame(220, 32, state)
		track = append(track, round("#070f15", 12, 12, 196, 8, 4), box(muted, 12, 15, 1, 2), box(muted, 207, 15, 1, 2))
		add("slider/track/"+state, 220, 32, track)
		add("slider/fill/"+state, 220, 32, []Shape{round(ink, 12, 12, 196, 8, 4), box("#eee5d266", 16, 13, 188, 1)})
		thumb := controlFrame(24, 32, state)
		thumb = append(thumb, box(ink, 9, 10, 2, 12), box(ink, 13, 10, 2, 12))
		add("slider/thumb/"+state, 24, 32, thumb)

		for _, value := range []string{"closed", "open"} {
			shapes := controlFrame(240, 40, state)
			shapes = append(shapes, box(muted, 202, 8, 1, 24))
			if value == "closed" {
				shapes = append(shapes, poly(ink, 211, 16, 227, 16, 219, 24))
			} else {
				shapes = append(shapes, poly(ink, 211, 24, 227, 24, 219, 16))
			}
			add("dropdown/"+value+"/"+state, 240, 40, shapes)
		}
		for _, value := range []string{"normal", "selected"} {
			shapes := controlFrame(238, 32, state)
			if value == "selected" {
				shapes = append(shapes, line(ink, 10, 16, 15, 21, 2), line(ink, 15, 21, 25, 11, 2), box(ink, 1, 6, 2, 20))
			}
			add("dropdown-option/"+value+"/"+state, 238, 32, shapes)
		}
	}
	add("dropdown-list", 240, 130, controlFrame(240, 130, "normal"))
	return out
}
