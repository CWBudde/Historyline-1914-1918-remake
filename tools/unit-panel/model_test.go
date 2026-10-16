package main

import (
	"encoding/json"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"reflect"
	"testing"
)

func TestRealFixtures(t *testing.T) {
	fs, err := makeFixtures()
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 9 {
		t.Fatalf("fixtures %d", len(fs))
	}
	cp := fs[0].Panel
	if cp.Country != "Germany" || cp.Owner != "Central Powers" || cp.Rank != 0 || cp.Stats[2].Value != "6/6" || cp.Stats[5].Value != "30" {
		t.Fatalf("runtime defaults %+v", cp)
	}
	if cp.Weapons[0].Ammo != "Unlimited" || cp.Weapons[1].Ammo != "5" || !cp.Weapons[3].Empty {
		t.Fatalf("weapon semantics %+v", cp.Weapons)
	}
	if fs[1].Panel.Country != "France" || fs[1].Panel.Weapons[1].Name != "revolveurM1892" {
		t.Fatal("EC data lost")
	}
	if fs[2].Panel.Country != "Germany" || fs[2].Panel.Owner != "Entente Cordial" || fs[2].Panel.Tactical != "ec" {
		t.Fatal("owner confused with type")
	}
	if fs[3].Panel.Tactical != "neutral" {
		t.Fatal("neutral owner")
	}
	if fs[4].Panel.Action != "Moved: yes · Attacked: no" || fs[4].Panel.Tactical != "cp" {
		t.Fatal("moved must preserve normal sprite")
	}
	d := fs[5].Panel
	if d.Stats[1].Value != "4" || d.Stats[2].Value != "1/6" || d.Stats[5].Value != "0" || d.Rank != 10 || d.Weapons[0].Ammo != "Unlimited" || d.Weapons[1].Ammo != "0" {
		t.Fatalf("depleted %+v", d)
	}
	if !reflect.DeepEqual(fs[7].Panel, fs[8].Panel) {
		t.Fatal("hidden data leak")
	}
	set, _ := data.Load()
	u, _ := game.NewUnit(set, "CP_infantry")
	u.Def.Name = "SECRET"
	u.Fuel = 987
	hidden := panelFor(u, false)
	raw, _ := json.Marshal(hidden)
	if string(raw) == "" || !reflect.DeepEqual(hidden, panelFor(nil, true)) {
		t.Fatalf("guard leaks %s", raw)
	}
}
