package dosoracle

import (
	"fmt"
	"reflect"
	"testing"
)

// R3 order entry: the attack order (0ad1:000c) and the cancel (10cc:0001)
// run on the state the menu leaves, and EnterAttack and Cancel, which
// transcribe them, must give the same state afterwards.

// diffDesk names the first part where two states differ.
func diffDesk(got, want Desk) string {
	for i, b := range got.Live {
		if !reflect.DeepEqual(b, want.Live[i]) {
			return fmt.Sprintf("live %d:\nemulator % x\nmodel    % x", i, b, want.Live[i])
		}
	}
	for k := range 2 {
		for c := range got.Maps[k] {
			if got.Maps[k][c] != want.Maps[k][c] {
				return fmt.Sprintf("map %d byte %d: emulator %#x, model %#x", k, c, got.Maps[k][c], want.Maps[k][c])
			}
		}
	}
	for i := range got.Bits {
		if got.Bits[i] != want.Bits[i] {
			return fmt.Sprintf("bit byte %d: emulator %#x, model %#x", i, got.Bits[i], want.Bits[i])
		}
	}
	for i, b := range got.Holds {
		if !reflect.DeepEqual(b, want.Holds[i]) {
			return fmt.Sprintf("container %d:\nemulator % x\nmodel    % x", i, b, want.Holds[i])
		}
	}
	if !reflect.DeepEqual(got.Record, want.Record) {
		return fmt.Sprintf("order record:\nemulator % x\nmodel    % x", got.Record, want.Record)
	}
	if !reflect.DeepEqual(got.Saved, want.Saved) {
		return fmt.Sprintf("saved record:\nemulator % x\nmodel    % x", got.Saved, want.Saved)
	}
	if len(got.Live) != len(want.Live) || len(got.Holds) != len(want.Holds) {
		return fmt.Sprintf("%d/%d live records, %d/%d containers", len(got.Live), len(want.Live), len(got.Holds), len(want.Holds))
	}
	return fmt.Sprintf("site %d/%d, list %v/%v, anim %#x/%#x, wait %#x/%#x, depots %d/%d, messages %v/%v, result %d/%d",
		got.Site, want.Site, got.List, want.List, got.Anim, want.Anim, got.Wait, want.Wait, got.Depots, want.Depots,
		got.Messages, want.Messages, got.Result, want.Result)
}

func TestEntryCoverage(t *testing.T) { eachExport(t, entryCoverage) }

func entryCoverage(t *testing.T, e *Export) {
	needMove(t, e)
	if e.B.Cycle == nil {
		t.Skip("the order-entry routines are not located in this build")
	}
	tab := moveTablesOf(e)
	seen := map[string]int{}

	// Attack orders: an attacker, a depot builder and a sapper (live +6
	// 0x4000/0x2000) of each side, on an enemy unit's cell and an empty
	// cell, with no target bit, the side's, the other side's or both, on
	// depot terrain or not, with 9 or 10 depots under way and an order
	// count of 0 or 3.
	target := holdMiddle + 1
	for side := range 2 {
		for _, f6 := range []int{0, 0x4000, 0x2000} {
			for _, occupied := range []bool{true, false} {
				for _, bits := range []int{0, 4 << side, 4 << (1 - side), 0x0C} {
					for _, onSite := range []bool{true, false} {
						for _, depots := range []int{9, 10} {
							for _, count := range []int{0, 3} {
								g := holdMovement(tab, uLight, side, holdMiddle)
								p := g.Units[0]
								p.Flags6 |= f6
								g.Units[0] = p
								if occupied {
									g.Units[2] = piece(tab, uLight, target, 1-side)
								}
								site := g.Terrain[target]
								if !onSite {
									site++
								}
								a := AttackOrder{Cell: target, Bits: map[int]int{target: bits}, Count: count, Depots: depots, Site: site}
								what := fmt.Sprintf("side %d +6 %#x occupied %v bits %#x site %v depots %d count %d",
									side, f6, occupied, bits, onSite, depots, count)
								before, got, err := e.RunAttackOrder(g, a)
								if err != nil {
									t.Fatalf("%s: %v", what, err)
								}
								want := EnterAttack(before, side, g.Width)
								if !reflect.DeepEqual(got, want) {
									t.Errorf("%s: %s", what, diffDesk(got, want))
								}
								live := got.Live[0]
								switch {
								case got.Result == 0 && got.Messages[0][0] == 0x45 && bits&(4<<side) == 0:
									seen["refused: no target bit"]++
								case got.Result == 0 && got.Messages[0][0] == 0x45:
									seen["refused: no unit in the target cell"]++
								case got.Result == 0 && got.Messages[0][0] == 0x47:
									seen["refused: 10 depots under way"]++
								case live[0x13] == 9 && len(got.List) == count+1:
									seen["attack appended"]++
								case live[0x13] == 0x0D && len(got.List) == count:
									seen["depot laid, not appended"]++
								case live[0x13] == 0x0B && len(got.List) == count:
									seen["depot builder off site, not appended"]++
								case live[0x13] == 0x11 && len(got.List) == count:
									seen["sapper, not appended"]++
								}
								if got.Result == 1 && word(live, 4)&0x200 != 0 {
									seen["accepted: moved flag set"]++
								}
							}
						}
					}
				}
			}
		}
	}

	// Cancels: a unit's reach with no order, an order to a reached cell and
	// one out of reach, and exits from either of two slots of a transporter
	// and of an own building, each with the cancel mode and without.
	for side := range 2 {
		for _, mode := range []int{0x10, 0} {
			for _, goal := range []int{-1, holdMiddle + 1, 0} {
				g := holdMovement(tab, uLight, side, holdMiddle)
				g.Budget = 4
				what := fmt.Sprintf("side %d mode %#x plain goal %d", side, mode, goal)
				cancelAgrees(t, e, what, g, CancelOrder{Mode: mode, Goal: goal}, seen)
			}
			for _, from := range []string{"transporter", "building"} {
				for _, slot := range []int{0, 1} {
					g := holdMovement(tab, uLight, side, holdMiddle)
					delete(g.Units, 0)
					rec := transportRec
					if from == "transporter" {
						addTransporter(tab, g, uSupply, holdMiddle, transportRec, 0xFC0)
					} else {
						rec = addBuilding(tab, g, tBuilding, holdMiddle, 0)
					}
					fill(tab, g, rec, side, holdMiddle, uLight, uLight)
					g.Unit, g.Budget = heldFirst+slot, 8
					o := &ExitOrder{Hold: rec, Slot: slot, MoveOrder: MoveOrder{Goal: -1, Retarget: -1}}
					what := fmt.Sprintf("side %d mode %#x exit from %s slot %d", side, mode, from, slot)
					cancelAgrees(t, e, what, g, CancelOrder{Mode: mode, Exit: o}, seen)
				}
			}
		}
	}

	for _, name := range []string{"refused: no target bit", "refused: no unit in the target cell",
		"refused: 10 depots under way", "attack appended", "depot laid, not appended",
		"depot builder off site, not appended", "sapper, not appended", "accepted: moved flag set",
		"reach bits cleared", "selection cleared", "exit: unit back in the slot it left",
		"exit: building position word even", "exit: transporter back on the map", "exit: building marker back"} {
		if seen[name] == 0 {
			t.Errorf("no case with %s", name)
		}
	}
	t.Logf("%v", seen)
}

func cancelAgrees(t *testing.T, e *Export, what string, g *Movement, c CancelOrder, seen map[string]int) {
	t.Helper()
	before, got, err := e.RunCancel(g, c)
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
	want := Cancel(before, g.Side, c.Mode)
	want.Result = got.Result // the routine returns nothing
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: %s", what, diffDesk(got, want))
		return
	}
	reach, sel := byte(1<<g.Side), byte(0x40<<g.Side)
	var hadReach, hasReach, hadSel, hasSel bool
	for i := range got.Bits {
		hadReach, hasReach = hadReach || before.Bits[i]&reach != 0, hasReach || got.Bits[i]&reach != 0
		hadSel, hasSel = hadSel || before.Bits[i]&sel != 0, hasSel || got.Bits[i]&sel != 0
	}
	if hadReach && !hasReach {
		seen["reach bits cleared"]++
	}
	if hadSel && !hasSel {
		seen["selection cleared"]++
	}
	if c.Exit == nil || c.Mode&0x10 == 0 {
		return
	}
	h := got.Holds[c.Exit.Hold]
	if int(h[7*g.Side+c.Exit.Slot]) == g.Unit {
		seen["exit: unit back in the slot it left"]++
	}
	// A building's first cell is the even map offset: the unit's position
	// word gets it too.
	cell := word(h, 0x0E)
	if word(got.Live[g.Unit], 0x0B+2*g.Side) == cell && cell&1 == 0 {
		seen["exit: building position word even"]++
	}
	switch v := got.Maps[0][cell|1]; {
	case word(h, 0x19)&0x20 != 0 && v == h[0x1B]:
		seen["exit: transporter back on the map"]++
	case v == 0xF4:
		seen["exit: building marker back"]++
	}
}
