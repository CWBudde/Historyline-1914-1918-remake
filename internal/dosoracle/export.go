package dosoracle

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
)

// HLSHA256 identifies the investigated HL.EXE build.
const HLSHA256 = "fe4fa44699ef36df0f1152786557e790651f44b836f9f6bffd3ca5dcc38a9b92"

const (
	baseSeg   = 0x1000 // load segment of the image
	scratch   = 0x6000 // segments from here up hold the synthetic structures
	retSeg    = 0x5FF0 // return address planted for the routine's retf
	recordLen = 69
)

var (
	stackLimit = [2]int{0x3CD8, 0x11FA}
	// errStackCheck reports that the Borland stack-overflow helper fired.
	errStackCheck = errors.New("stack overflow check fired")
)

// Export is the local load image written by tools/original-rules/disassemble.
// The original bytes stay under the ignored build/ directory.
type Export struct {
	Image  []byte
	Relocs []int
	Units  []byte
}

// LoadExport reads an export directory and refuses one from another build.
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
	if manifest.SHA != HLSHA256 {
		return nil, errors.New("export is not from the identified HL.EXE build")
	}
	e := &Export{}
	if e.Image, err = os.ReadFile(filepath.Join(dir, "HL.image.bin")); err != nil {
		return nil, err
	}
	if e.Units, err = os.ReadFile(filepath.Join(dir, "UNIT.DAT.bin")); err != nil {
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
// classes from 3ac5:0008.
func (e *Export) TerrainTerms() (t [9][2]int) {
	base := 0x3AC5<<4 + 8
	for i := range t {
		t[i] = [2]int{int(int8(e.Image[base+2*i])), int(int8(e.Image[base+2*i+1]))}
	}
	return t
}

// machine loads the image into a fresh interpreter.
func (e *Export) machine() *dos8086.Machine {
	m := dos8086.New()
	m.Load(e.Image, e.Relocs, baseSeg)
	return m
}

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
