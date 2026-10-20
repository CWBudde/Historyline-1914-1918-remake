package dosoracle

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
)

const (
	baseSeg   = 0x1000 // load segment of the image
	scratch   = 0x6000 // segments from here up hold the synthetic structures
	retSeg    = 0x5FF0 // return address planted for the routine's retf
	recordLen = 69
)

// errStackCheck reports that the Borland stack-overflow helper fired.
var errStackCheck = errors.New("stack overflow check fired")

// Export is the local load image written by cmd/dosexport.
// The original bytes stay under the ignored build/ directory.
type Export struct {
	B      *Build
	Image  []byte
	Relocs []int
	Units  []byte

	// Terrains holds the 6-byte terrain records (the summer table, WW1GNDS.DAT)
	// and Winter the winter table (WW1GNDW.DAT, nil where the export has none).
	Terrains, Winter []byte

	// watch, when set, receives every machine a runner creates, before the
	// runner sets it up (the field census of consumables_test.go).
	watch func(*dos8086.Machine)
}

// LoadExport reads an export directory and refuses one from an unidentified
// build.
func LoadExport(dir string) (*Export, error) {
	var manifest struct {
		SHA string `json:"hl_sha256"`
	}
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		return nil, err
	}
	b2 := BuildBySHA(manifest.SHA)
	if b2 == nil {
		return nil, errors.New("export is not from an identified HL.EXE build")
	}
	e := &Export{B: b2}
	if e.Image, err = os.ReadFile(filepath.Join(dir, "HL.image.bin")); err != nil {
		return nil, err
	}
	if e.Units, err = os.ReadFile(filepath.Join(dir, "UNIT.DAT.bin")); err != nil {
		return nil, err
	}
	if e.Terrains, err = os.ReadFile(filepath.Join(dir, "WW1GNDS.DAT.bin")); err != nil {
		return nil, err
	}
	if e.Winter, err = os.ReadFile(filepath.Join(dir, "WW1GNDW.DAT.bin")); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	f, err := os.Open(filepath.Join(dir, "relocations.csv"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	col := -1
	for i, h := range rows[0] {
		if h == "image_offset" {
			col = i
		}
	}
	if col < 0 {
		return nil, errors.New("relocations.csv has no image_offset column")
	}
	for _, r := range rows[1:] {
		v, err := strconv.Atoi(r[col])
		if err != nil {
			return nil, err
		}
		e.Relocs = append(e.Relocs, v)
	}
	return e, nil
}

// TypeRecord returns a copy of the 69-byte record of a unit type.
func (e *Export) TypeRecord(i int) []byte {
	return append([]byte(nil), e.Units[i*recordLen:(i+1)*recordLen]...)
}

// TerrainTerms returns the (attack, defence) signed terms of the nine terrain
// classes (3ac5:0008 in the full build).
func (e *Export) TerrainTerms() (t [9][2]int) {
	base := e.B.Terrain[0]<<4 + e.B.Terrain[1]
	for i := range t {
		t[i] = [2]int{int(int8(e.Image[base+2*i])), int(int8(e.Image[base+2*i+1]))}
	}
	return t
}

// machine loads the image into a fresh interpreter with the stack check
// disarmed and hooked.
func (e *Export) machine() *dos8086.Machine {
	m := dos8086.New()
	m.Load(e.Image, e.Relocs, baseSeg)
	m.Hooks[e.lin(e.B.StackCheck)] = errHook(errStackCheck)
	m.WW(baseSeg+e.B.StackLimit[0], e.B.StackLimit[1], 0xFFFF)
	if e.watch != nil {
		e.watch(m)
	}
	return m
}

// lin is the linear address of an image segment:offset pair once loaded.
func (e *Export) lin(a [2]int) int { return dos8086.Lin(baseSeg+a[0], a[1]) }

// arena hands out paragraph-aligned segments for synthetic structures.
type arena struct {
	m   *dos8086.Machine
	seg int
}

func newArena(m *dos8086.Machine) *arena { return &arena{m, scratch + 0x200} }

func (a *arena) alloc(data []byte) int {
	s := a.seg
	copy(a.m.Mem[s*16:], data)
	a.seg += (len(data)+15)/16 + 1
	return s
}

// farPtr stores a far pointer (offset 0, segment s) at ctx[off:].
func farPtr(ctx []byte, off, s int) {
	ctx[off], ctx[off+1], ctx[off+2], ctx[off+3] = 0, 0, byte(s), byte(s>>8)
}

func errHook(err error) dos8086.Hook {
	return func(*dos8086.Machine) error { return err }
}

func noHook(*dos8086.Machine) error { return nil }

func wrapRun(what string, err error) error {
	if err == nil || errors.Is(err, dos8086.ErrDivideError) {
		return err
	}
	return fmt.Errorf("%s: %w", what, err)
}

// winter is the export with the winter terrain table in place of the summer
// one; nil where the export has none.
func (e *Export) winter() *Export {
	if e.Winter == nil {
		return nil
	}
	w := *e
	w.Terrains = e.Winter
	return &w
}
