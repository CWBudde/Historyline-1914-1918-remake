// Command hlbackdrop keeps the painted backdrops' lossless sources and the
// JPEGs the game ships in step. A backdrop's source is the PNG under
// levels/backdrops in its scenario's directory, named like the backdrop
// (levels/backdrops/monthly/1915-08.png for levels/monthly/1915-08.jpg); every
// edit happens on the source, and the shipped JPEG is the source encoded with
// image/jpeg at the JPEG's own quality, read from its quantisation table (90
// for 1914-08 through 1915-04, 95 later; 95 for a new backdrop).
//
//	go run ./cmd/hlbackdrop jpeg [-q 95] monthly/1915-08 ...   # write the JPEGs from their sources
//	go run ./cmd/hlbackdrop check                               # every source encodes to its committed JPEG
//
// check fails on a backdrop without a source, a source whose JPEG differs and
// a source no scenario names.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // the sources
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// defaultQuality is the quality of a backdrop that has no JPEG yet.
const defaultQuality = 95

// workers is how many backdrops check decodes at once: a large one takes
// about 100 MB.
const workers = 3

// backdrop is one scenario's painting: its source and the JPEG it ships.
type backdrop struct {
	name     string // the scenario, "monthly/1915-08"
	src, out string // paths
}

func main() {
	log.SetFlags(0)
	levels := flag.String("levels", "levels", "the levels directory")
	q := flag.Int("q", 0, "jpeg: the quality (default: the JPEG's own, 95 for a new one)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: hlbackdrop [-levels dir] jpeg [-q n] monthly/1915-08 ... | check")
	}
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	switch args[0] {
	case "jpeg":
		fl := flag.NewFlagSet("jpeg", flag.ExitOnError)
		fl.IntVar(q, "q", *q, "the quality")
		_ = fl.Parse(args[1:])
		if fl.NArg() == 0 {
			flag.Usage()
			os.Exit(2)
		}
		for _, name := range fl.Args() {
			b, err := find(*levels, name)
			if err != nil {
				log.Fatal(err)
			}
			n, used, err := writeJPEG(b, *q)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s: %.2f MB at quality %d\n", b.out, float64(n)/1e6, used)
		}
	case "check":
		if errs := check(*levels); len(errs) > 0 {
			for _, err := range errs {
				log.Print(err)
			}
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(2)
	}
}

// find returns the backdrop of the scenario name ("monthly/1915-08").
func find(levels, name string) (backdrop, error) {
	name = strings.TrimSuffix(filepath.ToSlash(name), ".json")
	b, ok, err := backdropOf(levels, name)
	if err != nil {
		return backdrop{}, err
	}
	if !ok {
		return backdrop{}, fmt.Errorf("%s: the scenario names no backdrop", name)
	}
	return b, nil
}

// backdropOf reads the backdrop the scenario file levels/name.json names.
func backdropOf(levels, name string) (backdrop, bool, error) {
	data, err := os.ReadFile(filepath.Join(levels, filepath.FromSlash(name)+".json"))
	if err != nil {
		return backdrop{}, false, err
	}
	var f struct {
		Backdrop string `json:"backdrop"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		return backdrop{}, false, fmt.Errorf("%s: %w", name, err)
	}
	if f.Backdrop == "" {
		return backdrop{}, false, nil
	}
	if ext := strings.ToLower(filepath.Ext(f.Backdrop)); ext != ".jpg" && ext != ".jpeg" {
		return backdrop{}, false, fmt.Errorf("%s: backdrop %s is not a JPEG", name, f.Backdrop)
	}
	dir := filepath.Dir(filepath.FromSlash(name))
	stem := strings.TrimSuffix(f.Backdrop, filepath.Ext(f.Backdrop))
	return backdrop{
		name: name,
		src:  filepath.Join(levels, "backdrops", dir, stem+".png"),
		out:  filepath.Join(levels, dir, f.Backdrop),
	}, true, nil
}

// all lists the backdrops of every scenario file in the levels' directories
// (monthly, skirmish), in name order.
func all(levels string) ([]backdrop, error) {
	var out []backdrop
	for _, dir := range []string{"monthly", "skirmish"} {
		files, err := filepath.Glob(filepath.Join(levels, dir, "*.json"))
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			b, ok, err := backdropOf(levels, dir+"/"+strings.TrimSuffix(filepath.Base(f), ".json"))
			if err != nil {
				return nil, err
			}
			if ok {
				out = append(out, b)
			}
		}
	}
	return out, nil
}

// encode returns the backdrop's source encoded at quality q.
func encode(b backdrop, q int) ([]byte, error) {
	f, err := os.Open(b.src)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	im, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", b.src, err)
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, im, &jpeg.Options{Quality: q}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeJPEG encodes the backdrop's source to its JPEG at quality q, or at the
// existing JPEG's quality when q is 0, and returns the JPEG's size and the
// quality.
func writeJPEG(b backdrop, q int) (int, int, error) {
	if q == 0 {
		var err error
		if q, err = currentQuality(b.out); err != nil {
			return 0, 0, err
		}
	}
	data, err := encode(b, q)
	if err != nil {
		return 0, 0, err
	}
	return len(data), q, os.WriteFile(b.out, data, 0o644)
}

// currentQuality is the quality of the JPEG at path, defaultQuality when
// there is none.
func currentQuality(path string) (int, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return defaultQuality, nil
	}
	if err != nil {
		return 0, err
	}
	q, err := quality(data)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", path, err)
	}
	return q, nil
}

// check returns why the backdrops under levels and their sources disagree.
func check(levels string) []error {
	bs, err := all(levels)
	if err != nil {
		return []error{err}
	}
	errs := make([]error, len(bs))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for i := range jobs {
				errs[i] = checkOne(bs[i])
			}
		})
	}
	for i := range bs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	errs = slices.DeleteFunc(errs, func(err error) bool { return err == nil })

	named := map[string]bool{}
	for _, b := range bs {
		named[filepath.Clean(b.src)] = true
	}
	_ = filepath.WalkDir(filepath.Join(levels, "backdrops"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".png") && !named[filepath.Clean(p)] {
			errs = append(errs, fmt.Errorf("%s: no scenario names this backdrop", p))
		}
		return nil
	})
	return errs
}

func checkOne(b backdrop) error {
	want, err := os.ReadFile(b.out)
	if err != nil {
		return err
	}
	q, err := quality(want)
	if err != nil {
		return fmt.Errorf("%s: %w", b.out, err)
	}
	if _, err := os.Stat(b.src); err != nil {
		return fmt.Errorf("%s: no source (want %s)", b.out, b.src)
	}
	got, err := encode(b, q)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("%s is not %s at quality %d: run go run ./cmd/hlbackdrop jpeg %s", b.out, b.src, q, b.name)
	}
	return nil
}

// stdLuma is the IJG luminance quantisation table in zigzag order, as
// image/jpeg scales it for a quality setting.
var stdLuma = [64]int{
	16, 11, 12, 14, 12, 10, 16, 14, 13, 14, 18, 17, 16, 19, 24, 40,
	26, 24, 22, 22, 24, 49, 35, 37, 29, 40, 58, 51, 61, 60, 57, 51,
	56, 55, 64, 72, 92, 78, 64, 68, 87, 69, 55, 56, 80, 109, 81, 87,
	95, 98, 103, 104, 103, 62, 77, 113, 121, 112, 100, 120, 92, 101, 103, 99,
}

// quality is the image/jpeg quality setting whose luminance table the JPEG
// data carries; a table no setting gives is an error.
func quality(data []byte) (int, error) {
	var table []byte
	for i := 2; i+4 <= len(data) && data[i] == 0xFF; {
		marker, n := data[i+1], int(binary.BigEndian.Uint16(data[i+2:]))
		if marker == 0xDA || i+2+n > len(data) { // start of scan
			break
		}
		if marker == 0xDB && n >= 67 { // DQT
			table = data[i+4 : i+4+65]
			break
		}
		i += 2 + n
	}
	if table == nil {
		return 0, errors.New("no quantisation table")
	}
	if table[0]>>4 != 0 {
		return 0, errors.New("16-bit quantisation table")
	}
	for q := 1; q <= 100; q++ {
		s := 200 - 2*q
		if q < 50 {
			s = 5000 / q
		}
		match := true
		for k, v := range stdLuma {
			if int(table[1+k]) != min(max((v*s+50)/100, 1), 255) {
				match = false
				break
			}
		}
		if match {
			return q, nil
		}
	}
	return 0, errors.New("not an image/jpeg quantisation table")
}
