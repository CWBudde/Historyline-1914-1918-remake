package main

import (
	"fmt"
	"strconv"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
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

func panelFor(u *game.Unit, visible bool) Panel {
	if !visible || u == nil {
		return Panel{}
	}
	p := Panel{Available: true, Title: u.Def.Name, Country: u.Def.Country, Owner: u.Owner.String(), Rank: u.Experience, Tactical: map[game.Player]string{game.None: "neutral", game.CentralPowers: "cp", game.Entente: "ec"}[u.Owner]}
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
		u, e := game.NewUnit(set, typ)
		if e != nil {
			return nil, e
		}
		switch i {
		case 2:
			u.Owner = game.Entente
		case 3:
			u.Owner = game.None
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
