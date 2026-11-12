package unitpanel

import (
	"fmt"
	"strconv"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/faction"
)

type Stat struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Icon  string `json:"icon"`
}
type Weapon struct {
	Empty  bool   `json:"empty"`
	Name   string `json:"name,omitempty"`
	Ammo   string `json:"ammo,omitempty"`
	Power  int    `json:"power,omitempty"`
	Min    int    `json:"range_min,omitempty"`
	Max    int    `json:"range_max,omitempty"`
	Air    bool   `json:"air,omitempty"`
	Ground bool   `json:"ground,omitempty"`
	Water  bool   `json:"water,omitempty"`
	Icon   string `json:"icon"`
}
type Panel struct {
	Available bool     `json:"available"`
	Title     string   `json:"title,omitempty"`
	Country   string   `json:"country,omitempty"`
	Owner     string   `json:"owner,omitempty"`
	Tactical  string   `json:"tactical,omitempty"`
	Rank      int      `json:"rank,omitempty"`
	Stats     []Stat   `json:"stats,omitempty"`
	Weapons   []Weapon `json:"weapons,omitempty"`
	Capture   bool     `json:"capture,omitempty"`
	Action    string   `json:"action,omitempty"`
	Moved     bool     `json:"moved,omitempty"`
	Attacked  bool     `json:"attacked,omitempty"`
}
type Fixture struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Panel Panel  `json:"panel"`
}

// unit is the state of one unit the panel shows.
type unit struct {
	Def        data.Unit
	Owner      faction.Player
	Size       int // current size, 0..MaxSize
	MaxSize    int
	Experience int // rank 0..10
	Fuel       int
	Weapons    [4]*data.Weapon // per weapon slot; nil for an empty slot
	Ammo       [4]*int         // per weapon slot; nil for an empty slot or unlimited ammo
	Moved      bool
	Attacked   bool
}

// newUnit is a fresh unit of type id at the full strength of its original
// group, with no experience and full fuel and ammo.
func newUnit(set *data.Set, id string) (*unit, error) {
	def, ok := set.Units[id]
	if !ok {
		return nil, fmt.Errorf("unknown unit %q", id)
	}
	u := &unit{Def: def, Owner: faction.Of(def.Country), Size: def.Group, MaxSize: def.Group, Fuel: def.Fuel}
	for i, sl := range def.Weapons {
		if sl == nil {
			continue
		}
		wd := set.SlotWeapon(sl)
		u.Weapons[i] = &wd
		if a := wd.Ammo; a != nil {
			v := *a
			u.Ammo[i] = &v
		}
	}
	return u, nil
}

func panelFor(u *unit, visible bool) Panel {
	if !visible || u == nil {
		return Panel{}
	}
	p := Panel{Available: true, Title: u.Def.Name, Country: u.Def.Country, Owner: u.Owner.String(), Rank: u.Experience, Tactical: map[faction.Player]string{faction.None: "neutral", faction.CentralPowers: "cp", faction.Entente: "ec"}[u.Owner]}
	p.Stats = []Stat{{"Sight", strconv.Itoa(u.Def.Sight), "sight"}, {"Speed", strconv.Itoa(u.Def.Speed), "speed"}, {"Size", fmt.Sprintf("%d/%d", u.Size, u.MaxSize), "size"}, {"Armour", strconv.Itoa(u.Def.Armour), "armour"}, {"Weight", strconv.Itoa(u.Def.Weight), "weight"}, {"Fuel", strconv.Itoa(u.Fuel), "fuel"}}
	for _, s := range u.Def.Skills {
		if s == 1 {
			p.Capture = true
		}
	}
	for i, w := range u.Weapons {
		if w == nil {
			p.Weapons = append(p.Weapons, Weapon{Empty: true, Icon: "empty"})
			continue
		}
		ammo := "Unlimited"
		if u.Ammo[i] != nil {
			ammo = strconv.Itoa(*u.Ammo[i])
		}
		icon := "firearm"
		if i == 0 {
			icon = "bayonet"
		}
		p.Weapons = append(p.Weapons, Weapon{Name: w.Name, Ammo: ammo, Power: w.Power, Min: w.RangeMin, Max: w.RangeMax, Air: w.Air, Ground: w.Ground, Water: w.Water, Icon: icon})
	}
	yes := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}
	p.Action = fmt.Sprintf("Moved: %s · Attacked: %s", yes(u.Moved), yes(u.Attacked))
	p.Moved, p.Attacked = u.Moved, u.Attacked
	return p
}

func makeFixtures() ([]Fixture, error) {
	set, err := data.Load()
	if err != nil {
		return nil, err
	}
	ids := []string{"cp-ready", "ec-ready", "ownership-stress", "neutral-cp", "moved-cp", "depleted-cp", "long-labels", "empty", "hidden"}
	labels := []string{"Central Powers · runtime defaults", "Entente Cordial · runtime defaults", "Synthetic ownership contract: CP type, EC owner", "Synthetic neutral CP owner", "Moved; normal tactical thumbnail", "Synthetic depletion: zero fuel/ammo, size 1, rank 10", "Synthetic long translated labels", "No selected unit", "Hidden selection; same empty product"}
	fs := make([]Fixture, 0, 9)
	for i, id := range ids {
		typ := "CP_infantry"
		if i == 1 {
			typ = "EC_infantry"
		}
		u, e := newUnit(set, typ)
		if e != nil {
			return nil, e
		}
		switch i {
		case 2:
			u.Owner = faction.Entente
		case 3:
			u.Owner = faction.None
		case 4:
			u.Moved = true
		case 5:
			u.Attacked = true
			u.Fuel = 0
			u.Size = 1
			u.Experience = 10
			for j := range u.Ammo {
				if u.Ammo[j] != nil {
					*u.Ammo[j] = 0
				}
			}
		case 6:
			u.Def.Name = "Infanterie de reconnaissance internationale — Überlange Übersetzung für die Vorschau"
			for j, w := range u.Weapons {
				if w != nil {
					v := *w
					v.Name = "Équipement d’infanterie — très longue désignation traduite " + strconv.Itoa(j+1)
					u.Weapons[j] = &v
				}
			}
		case 7:
			u = nil
		}
		fs = append(fs, Fixture{ID: id, Label: labels[i], Panel: panelFor(u, i != 8)})
	}
	return fs, nil
}
