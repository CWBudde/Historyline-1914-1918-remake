package dosoracle

// Build holds the addresses the runners use in one identified executable.
// Pairs are segment:offset relative to the load image; single values are
// segments or offsets as named. Each address was located by instruction
// structure in that build (docs/research/original-dos-rules.md).
type Build struct {
	Name, Exe, SHA256 string

	// Casualty routine, its data segment with the terrain terms, the far
	// pointer to the battle context, the Borland stack limit and stack check,
	// and the hex distance helper that the runner replaces.
	Combat, Terrain, Ctx, StackLimit, StackCheck, Distance [2]int
	CombatDS                                               int

	// Data is the segment holding the map width (height follows), the live
	// unit table, the flag word of type record 0 (+0x0f) and the two support
	// result bytes. DS is the module data segment of the code segment Code.
	Data, Width, Units, TypeFlags, Results int
	DS, Code                               int

	// Offsets in Code: support caller, experience window and its two hooked
	// near-call helpers.
	Support, ExpFirst, ExpLast, ExpDestroy, ExpRefresh int
}

// HLSHA256 identifies the investigated HL.EXE build.
const HLSHA256 = "fe4fa44699ef36df0f1152786557e790651f44b836f9f6bffd3ca5dcc38a9b92"

// Full is the investigated HL.EXE of the full game.
var Full = Build{
	Name: "full", Exe: "HL.EXE", SHA256: HLSHA256,
	Combat: [2]int{0x2466, 0x000E}, Terrain: [2]int{0x3AC5, 0x0008}, Ctx: [2]int{0x3AC4, 0x0010},
	StackLimit: [2]int{0x3CD8, 0x11FA}, StackCheck: [2]int{0x0000, 0x1FC3}, Distance: [2]int{0x0E47, 0x2414},
	CombatDS: 0x3AC5,
	Data:     0x32F7, Width: 0x4CA8, Units: 0x18A4, TypeFlags: 0x09CB, Results: 0x1854,
	DS: 0x32E2, Code: 0x02DC,
	Support: 0x262A, ExpFirst: 0x0908, ExpLast: 0x0ABF, ExpDestroy: 0x2083, ExpRefresh: 0x254B,
}

// Demo is HL14-18.EXE of the free DOS demo. Its runtime and the code
// segment 02dc sit where the full build has them; the casualty routine, the
// distance helper and the data moved.
var Demo = Build{
	Name: "demo", Exe: "HL14-18.EXE", SHA256: "83fc4e3736d1d6b0bd30ddebaf304ae381ab28ee2ded3589322c466e71f5b846",
	Combat: [2]int{0x2451, 0x0004}, Terrain: [2]int{0x3AB2, 0x000E}, Ctx: [2]int{0x3AAE, 0x0045},
	StackLimit: [2]int{0x3CC5, 0x11FA}, StackCheck: [2]int{0x0000, 0x1FC3}, Distance: [2]int{0x0E40, 0x2421},
	CombatDS: 0x3AB2,
	Data:     0x32E2, Width: 0x4C96, Units: 0x1892, TypeFlags: 0x09C1, Results: 0x1842,
	DS: 0x32CD, Code: 0x02DC,
	Support: 0x262A, ExpFirst: 0x0908, ExpLast: 0x0ABF, ExpDestroy: 0x2083, ExpRefresh: 0x254B,
}

// Builds lists the identified executables.
var Builds = []*Build{&Full, &Demo}

// BuildBySHA returns the identified build with that executable hash, or nil.
func BuildBySHA(sha string) *Build {
	for _, b := range Builds {
		if b.SHA256 == sha {
			return b
		}
	}
	return nil
}
