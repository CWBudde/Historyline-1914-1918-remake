// Package faction names the two sides of the war, numbered as in the level
// files' owner field, and the side each unit's country fights on.
package faction

import "fmt"

// Player is a side.
type Player int

const (
	None          Player = 0
	CentralPowers Player = 1
	Entente       Player = 2
)

var names = map[string]Player{"None": None, "Central Powers": CentralPowers, "Entente Cordial": Entente}

// ByName maps a level file's player string to a Player.
func ByName(name string) (Player, error) {
	p, ok := names[name]
	if !ok {
		return None, fmt.Errorf("unknown player %q", name)
	}
	return p, nil
}

func (p Player) String() string {
	for name, q := range names {
		if p == q {
			return name
		}
	}
	return fmt.Sprintf("Player(%d)", int(p))
}

// Opponent is the other side; None has none.
func (p Player) Opponent() Player {
	switch p {
	case CentralPowers:
		return Entente
	case Entente:
		return CentralPowers
	}
	return None
}

// Of returns the side of a unit from its country.
func Of(country string) Player {
	switch country {
	case "Germany", "Austria", "Bulgaria", "Ottoman":
		return CentralPowers
	}
	return Entente
}
