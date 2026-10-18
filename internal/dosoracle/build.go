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
	// unit table, the word the support routine reads from type record 0 and
	// the two support result bytes. In the demo that word is the target-flag
	// word (+6) of the records at 32e2:09bb. DS is the module data segment of
	// the code segment Code.
	Data, Width, Units, TypeFlags, Results int
	DS, Code                               int

	// Offsets in Code: support caller, experience window and its two hooked
	// near-call helpers.
	Support, ExpFirst, ExpLast, ExpDestroy, ExpRefresh int

	// Handler is the order handler that runs the battles; nil where it has
	// not been located.
	Handler *Handler

	// Move holds the movement routines; nil where they have not been located.
	Move *Move
}

// Move holds the addresses of the movement queries (segment 0b06 and the
// end-cell check in 10cc in the demo). Offsets are in Build.Data.
type Move struct {
	// Reach fills the remaining budget per cell and the side's reach bits,
	// Destinations lists the reached cells the unit may end on, Path finds a
	// route over the reach bits and EndCell is the check Destinations calls.
	Reach, Destinations, Path, EndCell [2]int

	// The 69-byte type records, the 6-byte terrain records, the reach-bit
	// array (64 cells per row), the result count, the action record that
	// holds the moving unit, the map far pointers of side 0 and side 1 and
	// the far pointer to the map of the side whose turn it is.
	Types, Terrains, Bits, Count, Action, Current int
	Maps                                          [2]int

	// Order plans a move to a destination (the route and the reach bits of
	// its cells), Start begins it and Step moves the unit one cell. Sides
	// holds the per-side records (0x17 bytes: the buffer far pointer, the
	// order list far pointer and its count byte), Anim the animation flags,
	// Saved the two occupants a moving unit covers, and Containers the
	// 0x1c-byte container records a transporter's cargo sits in, followed
	// by the building tables (Buildings: flag 0x40, 0x100 and 0x400).
	Order, Start, Step                [2]int
	Sides, Anim, Saved, Containers    int
	Buildings                         [3]int
	Redraw, RouteDraw, Refresh, Sound [2]int    // drawing and sound, hooked
	View                              [2][2]int // view centring and view redraw, hooked
	Message                           [2]int    // the status message, hooked and recorded

	// Exit is the main loop's branch that takes a unit out of a container
	// and runs its reach search, ExitEnd the offset after it in Exit[0].
	Exit    [2]int
	ExitEnd int
}

// Handler holds the addresses of the order handler that resolves one side's
// attack orders (02dc:0004 in the demo).
type Handler struct {
	// Offsets in Build.Code: entry, the end of the first battle (the last
	// instruction of the experience window) and the exit after the last order.
	Entry, Stop, Done int

	// Offsets in Build.Data: the battle context the casualty routine reads,
	// the 69-byte type records, the 6-byte terrain records, the side
	// selector (2 runs side 0's orders), the two order lists of 0x17 bytes
	// (far pointer to live unit indices, count byte at +4), the two map far
	// pointers, the global flag word and the far pointer to the tile table.
	Ctx, Types, Terrains, Turn, Orders, Maps, Flags, Tiles int

	// The battle-animation switch byte, the battle-screen routine that runs
	// the casualty routine when its third argument is zero, and the far
	// pointer to the screen rectangle that routine draws into.
	Option, Caller, Screen [2]int

	// Routines that only draw, scroll the view or play sound.
	Present [][2]int

	// The game's RNG: the seed routine that its initialisation binds to the
	// context's callback at +0x6a, the draw routine, and the 32-bit state
	// both use (state*0x015a4e35+1, the Borland C rand).
	Srand, Rand, RandState [2]int
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
	Handler: &Handler{
		Entry: 0x0004, Stop: 0x0ABF, Done: 0x0D58,
		Ctx: 0x1812, Types: 0x09BB, Terrains: 0x05A1, Turn: 0x4D9E, Orders: 0x4D3A, Maps: 0x17A8,
		Flags: 0x4C9A, Tiles: 0x0151,
		Option: [2]int{0x37C1, 0x011B}, Caller: [2]int{0x1307, 0x178F}, Screen: [2]int{0x3CC5, 0x0092},
		Present: [][2]int{
			{0x109A, 0x0001}, {0x109A, 0x0123}, // map cursor and scrolling
			{0x0E40, 0x177E}, {0x0E40, 0x0D46}, {0x0E40, 0x0F38}, // view centring and redraw (32e2:4d88)
			{0x293D, 0x000C}, {0x28A6, 0x0004}, {0x29DA, 0x0000}, // mouse and screen updates
			{0x2F03, 0x1054}, // sound
		},
		Srand: [2]int{0x0000, 0x07AB}, Rand: [2]int{0x0000, 0x07C3}, RandState: [2]int{0x3CC5, 0x11F6},
	},
	Move: &Move{
		Reach: [2]int{0x0B06, 0x0331}, Destinations: [2]int{0x0B06, 0x1383}, Path: [2]int{0x0B06, 0x0E12},
		EndCell: [2]int{0x10CC, 0x0F74},
		Types:   0x09BB, Terrains: 0x05A1, Bits: 0x3B50, Count: 0x17E9, Action: 0x1805, Current: 0x17A4,
		Maps:  [2]int{0x17AC, 0x17A8},
		Order: [2]int{0x10CC, 0x02DD}, Start: [2]int{0x10CC, 0x05C3}, Step: [2]int{0x10CC, 0x071E},
		Sides: 0x4D36, Anim: 0x4C9E, Saved: 0x1807, Containers: 0x310C,
		Buildings: [3]int{0x38B4, 0x38EC, 0x3A04},
		Redraw:    [2]int{0x0E40, 0x145C}, RouteDraw: [2]int{0x0E40, 0x1219}, Refresh: [2]int{0x0E40, 0x0F38},
		Sound: [2]int{0x0CA0, 0x0279}, Message: [2]int{0x109A, 0x0001},
		View: [2][2]int{{0x0E40, 0x177E}, {0x0E40, 0x0D46}},
		Exit: [2]int{0x0661, 0x2F0A}, ExitEnd: 0x30FE,
	},
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
