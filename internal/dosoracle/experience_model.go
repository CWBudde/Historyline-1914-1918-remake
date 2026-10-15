package dosoracle

// ExpSide is one side of the experience award at 02dc:0908-0abf. Before is
// the strength saved ahead of the battle, After the live strength now, Flags6
// the live flag word at +6.
type ExpSide struct{ Exp, Before, After, Flags6 int }

const (
	expCap      = 6
	noAwardFlag = 0x20 // bit of the victim's live word at +6
)

// Award transcribes the awards: a side whose After is zero is destroyed and
// the opponent gains 1; a side that lost strength (as a byte) gives the
// opponent 1 unless its flag word has bit 0x20; experience is capped at 6.
// kills lists the destroyed sides (0 attacker, 1 defender).
func Award(a, d ExpSide) (attacker, defender int, kills []int) {
	var gain [2]int
	for i, me := range [2]ExpSide{a, d} {
		if me.After == 0 {
			gain[1-i]++
			kills = append(kills, i)
		}
		if (me.Before-me.After)&0xFF != 0 && me.Flags6&noAwardFlag == 0 {
			gain[1-i]++
		}
	}
	return min(expCap, a.Exp+gain[0]), min(expCap, d.Exp+gain[1]), kills
}
