// Command dosexport reproduces the static investigation of the DOS HL.EXE: the
// load image, objdump listings, relocation and far-call indexes, strings, and the
// decoded unit and terrain tables. Output goes to the ignored build/ directory;
// original bytes never enter Git.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosexe"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosoracle"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
)

// codeEnd is the first module data segment in this exact executable.
const codeEnd = 0x32E20

// window is a manually traced range: segment, first instruction, exclusive end.
// They are not an automatic partition into functions.
type window struct {
	name            string
	seg, start, end int
}

var windows = []window{
	{"combat", 0x2466, 0x000E, 0x0ABC},
	{"combat_wrapper", 0x2453, 0x0009, 0x013C},
	{"experience", 0x02DC, 0x0908, 0x0ABF},
	{"pincer_and_blocking", 0x02DC, 0x262A, 0x29C9},
	{"score", 0x02DC, 0x29C9, 0x2B1D},
	{"repair", 0x0660, 0x3108, 0x3268},
	{"unit_constructor", 0x15A4, 0x0009, 0x01AB},
	{"unit_information", 0x272C, 0x0650, 0x0A0F},
	{"hex_distance", 0x0E47, 0x2414, 0x24DB},
	{"long_arithmetic", 0x0000, 0x03A1, 0x04CB},
}

var datFiles = []string{"UNIT.DAT", "WW1GNDS.DAT", "WW1GNDW.DAT", "NEWUNIT.DAT", "MNUDATA.DAT"}

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "dosexport:", err)
	os.Exit(1)
}

func writeCSV(path string, header []string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	w.UseCRLF = true
	_ = w.Write(header)
	_ = w.WriteAll(rows)
	w.Flush()
	if err := w.Error(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func itoa(v int) string { return strconv.Itoa(v) }

func disassemble(objdump, binary, output string, start int, end int) error {
	args := []string{"-D", "-b", "binary", "-m", "i8086", "-M", "intel", fmt.Sprintf("--start-address=%d", start)}
	if end >= 0 {
		args = append(args, fmt.Sprintf("--stop-address=%d", end))
	}
	cmd := exec.Command(objdump, append(args, binary)...)
	cmd.Env = append(os.Environ(), "LC_ALL=C") // keep the listing header locale-independent
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("objdump %s: %w", binary, err)
	}
	return os.WriteFile(output, out, 0o644)
}

func main() {
	source := flag.String("source", "original/extracted/dos-full", "installation directory")
	out := flag.String("out", "build/original-rules/dos-full", "fresh output directory")
	flag.Parse()
	objdump, err := exec.LookPath("objdump")
	if err != nil {
		fatal(fmt.Errorf("GNU objdump with i8086 support is required"))
	}
	if _, err := os.Stat(*out); err == nil {
		fatal(fmt.Errorf("output already exists; choose a fresh directory"))
	}
	src, _ := filepath.Abs(*source)
	dst, _ := filepath.Abs(*out)
	if dst == src || strings.HasPrefix(dst, src+string(filepath.Separator)) {
		fatal(fmt.Errorf("output must be outside the original installation"))
	}
	exe, err := os.ReadFile(filepath.Join(src, "HL.EXE"))
	if err != nil {
		fatal(err)
	}
	if sum(exe) != dosoracle.HLSHA256 {
		fatal(fmt.Errorf("HL.EXE differs from the investigated build; named offsets need retracing"))
	}
	header, image, relocs, err := dosexe.ParseMZ(exe)
	if err != nil {
		fatal(err)
	}
	decoded := map[string][]byte{}
	for _, n := range datFiles {
		raw, err := os.ReadFile(filepath.Join(src, n))
		if err != nil {
			fatal(err)
		}
		if decoded[n], err = tpwm.Unpack(raw, binary.LittleEndian); err != nil {
			fatal(fmt.Errorf("%s: %w", n, err))
		}
	}
	units, err := dosexe.Units(decoded["UNIT.DAT"])
	if err != nil {
		fatal(err)
	}
	ground := map[string][]dosexe.Terrain{}
	for _, n := range []string{"WW1GNDS.DAT", "WW1GNDW.DAT"} {
		if ground[n], err = dosexe.TerrainTable(decoded[n]); err != nil {
			fatal(err)
		}
	}
	for _, d := range []string{"segments", "windows"} {
		if err := os.MkdirAll(filepath.Join(dst, d), 0o755); err != nil {
			fatal(err)
		}
	}
	imagePath := filepath.Join(dst, "HL.image.bin")
	check := func(err error) {
		if err != nil {
			fatal(err)
		}
	}
	check(os.WriteFile(imagePath, image, 0o644))
	check(disassemble(objdump, imagePath, filepath.Join(dst, "HL.image.linear.asm"), 0, -1))

	segSet := map[int]bool{}
	for _, r := range relocs {
		if r.Segment*16 < codeEnd {
			segSet[r.Segment] = true
		}
	}
	var segments []int
	for s := range segSet {
		segments = append(segments, s)
	}
	sort.Ints(segments)
	for i, seg := range segments {
		// Borland starts a new address segment before the preceding function's
		// final paragraph ends. Keep the 15-byte overlap; never truncate its tail.
		next := codeEnd
		if i+1 < len(segments) {
			next = segments[i+1] * 16
		}
		end := min(next+15, codeEnd)
		bin := filepath.Join(dst, "segments", fmt.Sprintf("%04x.bin", seg))
		check(os.WriteFile(bin, image[seg*16:end], 0o644))
		check(disassemble(objdump, bin, strings.TrimSuffix(bin, ".bin")+".asm", 0, -1))
	}

	type winInfo struct {
		Name         string `json:"name"`
		Segment      string `json:"segment"`
		Start        string `json:"start"`
		EndExclusive string `json:"end_exclusive"`
		FileOffset   int    `json:"file_offset"`
	}
	var infos []winInfo
	for _, w := range windows {
		bin := filepath.Join(dst, "windows", w.name+".bin")
		check(os.WriteFile(bin, image[w.seg*16:w.seg*16+w.end], 0o644))
		check(disassemble(objdump, bin, strings.TrimSuffix(bin, ".bin")+".asm", w.start, w.end))
		infos = append(infos, winInfo{w.name, fmt.Sprintf("%04x", w.seg), fmt.Sprintf("%04x", w.start),
			fmt.Sprintf("%04x", w.end), header.HeaderBytes + w.seg*16 + w.start})
	}

	var rows [][]string
	for _, r := range relocs {
		rows = append(rows, []string{itoa(r.Segment), itoa(r.Offset), itoa(r.ImageOffset), itoa(r.SegmentValue)})
	}
	check(writeCSV(filepath.Join(dst, "relocations.csv"),
		[]string{"segment", "offset", "image_offset", "segment_value"}, rows))

	// Only relocation-backed immediate far calls/jumps. Not a complete call graph:
	// near calls and indirect callbacks are deliberately excluded.
	rows = nil
	for _, r := range relocs {
		pos := r.ImageOffset
		if pos >= 3 && pos < codeEnd && (image[pos-3] == 0x9A || image[pos-3] == 0xEA) {
			kind := "jump"
			if image[pos-3] == 0x9A {
				kind = "call"
			}
			rows = append(rows, []string{fmt.Sprintf("%04x", r.Segment), fmt.Sprintf("%04x", r.Offset-3),
				itoa(pos - 3), fmt.Sprintf("%04x", r.SegmentValue),
				fmt.Sprintf("%04x", binary.LittleEndian.Uint16(image[pos-2:])), kind})
		}
	}
	farCalls := len(rows)
	check(writeCSV(filepath.Join(dst, "far_calls.csv"),
		[]string{"source_segment", "source_offset", "image_offset", "target_segment", "target_offset", "kind"}, rows))

	rows = nil
	for _, u := range units {
		rows = append(rows, []string{itoa(u.Index), itoa(u.Offset), u.Category, u.Name,
			itoa(u.MovementBudgetRaw), itoa(u.DisplayedSpeed), itoa(u.Armour), itoa(u.MaxGroup),
			itoa(u.AlternateStrengthRaw), itoa(u.MovementMaskRaw), itoa(u.TargetFlagsRaw),
			itoa(u.AirRangeEncoded), itoa(u.SurfaceRangeEncoded), itoa(u.AirPower), itoa(u.GroundPower),
			itoa(u.WaterPower), itoa(u.Flags13Raw), itoa(u.Flags15Raw), itoa(u.Flags17Raw),
			itoa(u.CostRaw), itoa(u.CapacityRaw), itoa(u.WeightRaw), hex.EncodeToString(u.Raw)})
	}
	check(writeCSV(filepath.Join(dst, "units.csv"), []string{"index", "offset", "category", "name",
		"movement_budget_raw", "displayed_speed", "armour", "max_group", "alternate_strength_raw",
		"movement_mask_raw", "target_flags_raw", "air_range_encoded", "surface_range_encoded",
		"air_power", "ground_power", "water_power", "flags_13_raw", "flags_15_raw", "flags_17_raw",
		"cost_raw", "capacity_raw", "weight_raw", "raw_hex"}, rows))
	for _, n := range datFiles {
		check(os.WriteFile(filepath.Join(dst, n+".bin"), decoded[n], 0o644))
	}
	for n, t := range ground {
		rows = nil
		for _, r := range t {
			rows = append(rows, []string{itoa(r.Index), itoa(r.Offset), itoa(r.FlagsRaw), itoa(r.MovementMaskRaw),
				itoa(r.MovementCostRaw), itoa(r.Byte4Raw), itoa(r.CombatClass), hex.EncodeToString(r.Raw)})
		}
		check(writeCSV(filepath.Join(dst, n+".csv"), []string{"index", "offset", "flags_raw",
			"movement_mask_raw", "movement_cost_raw", "byte_4_raw", "combat_class", "raw_hex"}, rows))
	}

	rows = nil
	for _, m := range regexp.MustCompile(`[\x20-\x7e]{4,}`).FindAllIndex(exe, -1) {
		if m[0] >= header.HeaderBytes {
			rows = append(rows, []string{itoa(m[0]), itoa(m[0] - header.HeaderBytes), string(exe[m[0]:m[1]])})
		}
	}
	check(writeCSV(filepath.Join(dst, "strings.csv"), []string{"file_offset", "image_offset", "text"}, rows))

	version, err := exec.Command(objdump, "--version").Output()
	check(err)
	type srcInfo struct {
		StoredBytes    int        `json:"stored_bytes"`
		SHA256         string     `json:"sha256"`
		UnpackedBytes  *int       `json:"unpacked_bytes,omitempty"`
		UnpackedSHA256 string     `json:"unpacked_sha256,omitempty"`
		MZ             *dosexe.MZ `json:"mz,omitempty"`
	}
	sources := map[string]srcInfo{}
	for _, n := range append([]string{"HL.EXE", "HL14-18.EXE"}, datFiles...) {
		raw, err := os.ReadFile(filepath.Join(src, n))
		check(err)
		e := srcInfo{StoredBytes: len(raw), SHA256: sum(raw)}
		if d, ok := decoded[n]; ok {
			l := len(d)
			e.UnpackedBytes, e.UnpackedSHA256 = &l, sum(d)
		}
		if strings.HasSuffix(n, ".EXE") {
			h, _, _, err := dosexe.ParseMZ(raw)
			check(err)
			e.MZ = &h
		}
		sources[n] = e
	}
	manifest := struct {
		Source       string             `json:"source"`
		HLSHA256     string             `json:"hl_sha256"`
		MZ           dosexe.MZ          `json:"mz"`
		Objdump      string             `json:"objdump_version"`
		SegmentCount int                `json:"segment_count"`
		FarCalls     int                `json:"far_call_or_jump_count"`
		Windows      []winInfo          `json:"windows"`
		Sources      map[string]srcInfo `json:"sources"`
	}{src, sum(exe), header, strings.SplitN(string(version), "\n", 2)[0], len(segments), farCalls, infos, sources}
	b, err := json.MarshalIndent(manifest, "", "  ")
	check(err)
	check(os.WriteFile(filepath.Join(dst, "manifest.json"), append(b, '\n'), 0o644))
	fmt.Printf("Wrote %s: %d image bytes, %d segment listings, %d traced windows, %d units, %d terrain records per season\n",
		dst, len(image), len(segments), len(windows), len(units), dosexe.TerrainCount)
}
