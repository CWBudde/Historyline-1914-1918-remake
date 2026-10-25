package finmap

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
)

// The tests below read the original game files from original/ above the
// working directory and skip without them. Their fixtures are numbers only.

func originalDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		o := filepath.Join(dir, "original", "extracted", "dos-full")
		if st, err := os.Stat(filepath.Join(o, "MAP")); err == nil && st.IsDir() {
			return o
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("no original files")
		}
		dir = parent
	}
}

func fullRules() Rules { return RulesFrom(dosrules.FullTables) }

func readPacked(t *testing.T, path string) ([]byte, bool) {
	t.Helper()
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false
	}
	if err != nil {
		t.Fatal(err)
	}
	u, err := tpwm.Unpack(b, binary.LittleEndian)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return u, true
}

// fullLevels decodes the 72 numbered maps of the full game.
func fullLevels(t *testing.T) ([]*Level, []Code) {
	t.Helper()
	dir := originalDir(t)
	raw, _ := readPacked(t, filepath.Join(dir, "CODES.DAT"))
	codes, err := DecodeCodes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 72 {
		t.Fatalf("%d codes", len(codes))
	}
	var out []*Level
	for i := range codes {
		stem := filepath.Join(dir, "MAP", strconv.Itoa(i/10)+strconv.Itoa(i%10))
		in := Input{Index: i, Code: &codes[i]}
		b, ok := readPacked(t, stem+".FIN")
		if !ok {
			t.Fatalf("map %d missing", i)
		}
		if in.Map, err = DecodeMap(b); err != nil {
			t.Fatalf("map %d: %v", i, err)
		}
		if b, ok := readPacked(t, stem+".SHP"); ok {
			if in.Shop, err = DecodeShop(b); err != nil {
				t.Fatalf("map %d: %v", i, err)
			}
		}
		if b, ok := readPacked(t, stem+".COM"); ok {
			if in.Computer, err = DecodeComputer(b); err != nil {
				t.Fatalf("map %d: %v", i, err)
			}
		}
		l, err := Build(in, fullRules())
		if err != nil {
			t.Fatalf("map %d: %v", i, err)
		}
		out = append(out, l)
	}
	return out, codes
}

// The single-player campaigns are maps 0-23 (Central Powers) and 24-47
// (Allies), the two-player maps 48-71; only single-player maps have a
// computer file, and on two-player maps only HQs carry a computer value.
func TestOriginalMissions(t *testing.T) {
	ls, codes := fullLevels(t)
	for i, l := range ls {
		m, c := l.Mission, codes[i]
		wantPlayers, wantHuman := 1, []int{Central}
		switch {
		case i >= 48:
			wantPlayers, wantHuman = 2, []int{Central, Allied}
		case i >= 24:
			wantHuman = []int{Allied}
		}
		if m.Players != wantPlayers || !slices.Equal(m.HumanSides, wantHuman) || l.Computer != (i < 48) {
			t.Errorf("map %d: players %d human %v computer file %v", i, m.Players, m.HumanSides, l.Computer)
		}
		last := i == 23 || i == 47 || i == 71
		if (m.Next == nil) != last || (m.Next != nil && *m.Next != i+1) || m.CampaignEnd != (i == 23 || i == 47) {
			t.Errorf("map %d: next %v campaign end %v", i, m.Next, m.CampaignEnd)
		}
		if c.Raw5 != 2 || c.Raw9 != 0 {
			t.Errorf("map %d: raw bytes %d %d", i, c.Raw5, c.Raw9)
		}
		for _, k := range l.Containers {
			if i >= 48 && k.Kind != "hq" && k.ComputerValue != 0 {
				t.Errorf("map %d: two-player %s computer value %d", i, k.Kind, k.ComputerValue)
			}
		}
	}
	// The winter maps: two in every six of each campaign.
	var winter []int
	for i, l := range ls {
		if l.Mission.Winter {
			winter = append(winter, i)
		}
	}
	want := []int{1, 2, 8, 9, 13, 14, 19, 20, 25, 26, 32, 33, 37, 38, 43, 44, 49, 50, 55, 56, 61, 62, 67, 68}
	if !slices.Equal(winter, want) {
		t.Errorf("winter maps %v", winter)
	}
}

// The appendix of the manual lists the passwords of the two-player maps in
// map order.
func TestOriginalManualPasswords(t *testing.T) {
	ls, _ := fullLevels(t)
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "manual", "instructions", "appendix.typ"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	i := strings.Index(s, "<two-player-maps>")
	if i < 0 {
		t.Fatal("no <two-player-maps> table in the manual")
	}
	s = s[i:]
	if i := strings.Index(s, "map("); i >= 0 {
		s = s[:i]
	}
	got := regexp.MustCompile(`"([A-Z]{5})"`).FindAllStringSubmatch(s, -1)
	if len(got) != 24 {
		t.Fatalf("%d passwords in the manual", len(got))
	}
	for i, m := range got {
		if ls[48+i].Mission.Code != m[1] {
			t.Errorf("map %d: password differs from the manual's entry %d", 48+i, i)
		}
	}
}

// Every map has two HQs, one per side; buildings of a kind take their table
// entries by entrance; each transporter record names a transporter of the
// record's side, except where pinned.
func TestOriginalContainers(t *testing.T) {
	ls, _ := fullLevels(t)
	var noHolder, sideDiffers []string
	for i, l := range ls {
		hq := map[int]int{}
		for _, b := range l.Buildings {
			if b.Kind == "hq" {
				hq[b.Owner]++
			}
		}
		if len(hq) != 2 || hq[Central] != 1 || hq[Allied] != 1 {
			t.Errorf("map %d: HQs %v", i, hq)
		}
		owner := map[int]int{}
		for _, b := range l.Buildings {
			owner[b.Container] = b.Owner
		}
		for _, u := range l.Units {
			if u.Container != nil {
				owner[*u.Container] = u.Side
			}
		}
		for _, c := range l.Containers {
			id := strconv.Itoa(i) + ":" + strconv.Itoa(c.Container)
			o, ok := owner[c.Container]
			switch {
			case c.Holder == "none":
				noHolder = append(noHolder, id)
			case ok && o != c.Side && (c.Kind != "hq" || c.Index < 2):
				sideDiffers = append(sideDiffers, id)
			}
		}
	}
	sort.Strings(noHolder)
	// Map 48 has a record for a fifth transporter it does not have.
	if !slices.Equal(noHolder, []string{"48:4"}) {
		t.Errorf("records without a holder %v", noHolder)
	}
	// Records whose units are created for another side than the
	// building's owner or the transporter's side.
	if !slices.Equal(sideDiffers, []string{"6:72", "11:3", "14:84", "16:82", "16:83", "32:72", "45:73", "45:74", "47:72", "47:73", "48:1", "51:72"}) {
		t.Errorf("records of another side %v", sideDiffers)
	}
}

// The full game's maps: terrain, units and buildings within the tables, an
// HQ per side, the campaigns' chains, winter and players.
func TestOriginalMapShapes(t *testing.T) {
	terrains, types := len(dosrules.FullTables.Move.Terrains), len(dosrules.FullTables.Move.Types)
	// Winter maps by position in their series of 24: both campaigns, then
	// the two-player series.
	winter := [][]int{{1, 2, 8, 9, 13, 14, 19, 20}, {1, 2, 8, 9, 13, 14, 19, 20}, {1, 2, 7, 8, 13, 14, 19, 20}}
	ls, _ := fullLevels(t)
	for i, l := range ls {
		in := func(c [2]int) bool { return c[0] >= 0 && c[0] < l.Height && c[1] >= 0 && c[1] < l.Width }
		if len(l.Terrain) != l.Height {
			t.Fatalf("map %02d: %d rows of %d", i, len(l.Terrain), l.Height)
		}
		for _, row := range l.Terrain {
			if len(row) != l.Width || slices.Max(row) >= terrains || slices.Min(row) < 0 {
				t.Fatalf("map %02d: row %v", i, row)
			}
		}
		for _, u := range l.Units {
			if !in(u.Cell) || u.Type < 0 || u.Type >= types || u.Side < Central || u.Side > Neutral {
				t.Errorf("map %02d: unit %+v", i, u)
			}
		}
		hq := map[int]int{}
		for _, b := range l.Buildings {
			if !in(b.Cell) || b.Container < HQBase || b.Container >= Containers {
				t.Errorf("map %02d: building %+v", i, b)
			}
			if b.Kind == HQ.String() {
				hq[b.Owner]++
			}
		}
		if len(hq) != 2 || hq[Central] != 1 || hq[Allied] != 1 {
			t.Errorf("map %02d: HQs %v", i, hq)
		}
		for _, c := range l.Containers {
			for _, u := range c.Units {
				if u.Type < 0 || u.Type >= types {
					t.Errorf("map %02d: container %d holds type %d", i, c.Container, u.Type)
				}
			}
		}

		m := l.Mission
		if m == nil {
			t.Fatalf("map %02d: no mission", i)
		}
		last := i%24 == 23
		if (m.Next == nil) != last || (m.Next != nil && *m.Next != i+1) || m.CampaignEnd != (last && i < 48) {
			t.Errorf("map %02d: next %v campaign end %v", i, m.Next, m.CampaignEnd)
		}
		if m.Winter != slices.Contains(winter[i/24], i%24) {
			t.Errorf("map %02d: winter %v", i, m.Winter)
		}
		if (m.Players == 1) != (i < 48) || l.Computer != (i < 48) {
			t.Errorf("map %02d: players %d computer %v", i, m.Players, l.Computer)
		}
	}
}
