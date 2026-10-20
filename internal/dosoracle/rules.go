package dosoracle

import "github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"

// The movement, route, container and visibility rules are internal/dosrules;
// the runners take and return its record-level types.
type (
	MoveType    = dosrules.MoveType
	MoveTerrain = dosrules.MoveTerrain
	MoveTables  = dosrules.MoveTables
	Piece       = dosrules.Piece
	Movement    = dosrules.Movement
	MoveOrder   = dosrules.MoveOrder
	StepRun     = dosrules.StepRun
	Hold        = dosrules.Hold
	Action      = dosrules.Action
	ExitOrder   = dosrules.ExitOrder
	ExitRun     = dosrules.ExitRun
	Sighted     = dosrules.Sighted
	PhaseChange = dosrules.PhaseChange
	PhaseRun    = dosrules.PhaseRun
)

func putWord(b []byte, off, v int) { b[off], b[off+1] = byte(v), byte(v>>8) }
