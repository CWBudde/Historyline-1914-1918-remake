package dosplay

import (
	"fmt"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// Domain is the kind of fighting a unit stands for in a battle vignette.
type Domain int

// The battle domains; the zero value is Infantry.
const (
	Infantry Domain = iota
	Armour
	Artillery
	Naval
	Air
)

var domainNames = [...]string{"infantry", "armour", "artillery", "naval", "air"}

func (d Domain) String() string {
	if d >= 0 && int(d) < len(domainNames) {
		return domainNames[d]
	}
	return fmt.Sprintf("Domain(%d)", int(d))
}

// Type-record bits the domain reads.
const (
	classWater = 0x08 // class word (+0x0d): a boat
	classAir   = 0x10 // class word: an aircraft
	flagFoot   = 0x01 // flags word (+0x11): the soldiers and horsemen, whom no enemy blocks (Movement)
)

// TypeDomain is the battle domain of UNIT.DAT type typ by its record:
//   - Air for the aircraft and the balloon, Naval for the boats, by the class
//     word;
//   - for a ground type, Infantry for the soldiers and horsemen (flags word
//     bit 0: infantry, elite infantry, sappers, the anti-tank gun, cavalry);
//   - Artillery for a ground type whose weapon reaches beyond the next hex
//     as the unit info shows it, against surface or air targets: the
//     artillery, the bunker, the AA guns and the gun trains;
//   - Armour for the other ground types: the tanks, armoured cars, the
//     supply vehicles and the construction unit.
//
// It is Infantry for a type the tables lack.
func TypeDomain(t *dosrules.Tables, typ int) Domain {
	if typ < 0 || typ >= len(t.Types) {
		return Infantry
	}
	rt := t.Types[typ]
	// The class word is also the type record's AttackMask.
	class, flags := rt.AttackMask, 0
	if t.Move != nil && typ < len(t.Move.Types) {
		class, flags = t.Move.Types[typ].Class, t.Move.Types[typ].Flags
	}
	switch {
	case class&classAir != 0:
		return Air
	case class&classWater != 0:
		return Naval
	case flags&flagFoot != 0:
		return Infantry
	case rt.SurfaceRange-1 >= 2 || rt.AirRange-1 >= 2:
		return Artillery
	}
	return Armour
}

// DomainOf is the battle domain of unit idx (live record), Infantry for no
// such unit.
func (s *Session) DomainOf(idx int) Domain {
	u, ok := s.Game.Units[idx]
	if !ok {
		return Infantry
	}
	return TypeDomain(s.Game.Profile.Tables, u.Type)
}
