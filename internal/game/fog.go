package game

// Fog is what a side knows about a hex.
type Fog uint8

const (
	Unseen  Fog = iota // never seen
	Seen               // seen before, not visible now
	Visible            // in sight now
)

func (f Fog) String() string {
	switch f {
	case Unseen:
		return "Unseen"
	case Seen:
		return "Seen"
	case Visible:
		return "Visible"
	}
	return "Fog(?)"
}
