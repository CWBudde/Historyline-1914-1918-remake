// Package hlbackdrop keeps the painted backdrops' lossless sources and the
// JPEGs the game ships in step. A level folder's backdrop is backdrop.jpg,
// its source backdrop.png beside it (levels/1915-08/backdrop.png for
// levels/1915-08/backdrop.jpg); every edit happens on the source, and the
// shipped JPEG is the source encoded with image/jpeg at the JPEG's own
// quality, read from its quantisation table (90 for 1914-08 through 1915-04,
// 95 later; 95 for a new backdrop).
//
//	go run ./cmd/hl levels backdrop jpeg [-q 95] 1915-08 ...   # write the JPEGs from their sources
//	go run ./cmd/hl levels backdrop check                       # every source encodes to its committed JPEG
//
// check fails on a backdrop without a source or without its JPEG, a source
// whose JPEG differs and a source outside a level folder.
package hlbackdrop

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // the sources
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// defaultQuality is the quality of a backdrop that has no JPEG yet.
const defaultQuality = 95

// workers is how many backdrops check decodes at once: a large one takes
// about 100 MB.
const workers = 3

// backdrop is one level folder's painting: its source and the JPEG it ships.
type backdrop struct {
	name     string // the level folder, "1915-08"
	src, out string // paths
}

// find returns the backdrop of the level folder name ("1915-08").
func find(levels, name string) (backdrop, error) {
	name = strings.Trim(filepath.ToSlash(filepath.Clean(name)), "/")
	if _, err := os.Stat(filepath.Join(levels, filepath.FromSlash(name), scenario.MapFile)); err != nil {
		return backdrop{}, fmt.Errorf("%s: no level folder: %w", name, err)
	}
	return backdropOf(levels, name), nil
}

// backdropOf is the backdrop of the level folder levels/name.
func backdropOf(levels, name string) backdrop {
	dir := filepath.Join(levels, filepath.FromSlash(name))
	return backdrop{
		name: name,
		src:  filepath.Join(dir, scenario.BackdropSource),
		out:  filepath.Join(dir, scenario.BackdropFile),
	}
}

// all lists the backdrops of the level folders that have a JPEG or a
// source, in folder order.
func all(levels string) ([]backdrop, error) {
	dirs, err := scenario.Months(os.DirFS(levels))
	if err != nil {
		return nil, err
	}
	var out []backdrop
	for _, dir := range dirs {
		b := backdropOf(levels, dir)
		_, errOut := os.Stat(b.out)
		_, errSrc := os.Stat(b.src)
		if errOut == nil || errSrc == nil {
			out = append(out, b)
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
	_ = filepath.WalkDir(levels, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == scenario.BackdropSource && !named[filepath.Clean(p)] {
			errs = append(errs, fmt.Errorf("%s: not in a level folder", p))
		}
		return nil
	})
	return errs
}

func checkOne(b backdrop) error {
	want, err := os.ReadFile(b.out)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: no JPEG for the source (hlbackdrop jpeg %s)", b.src, b.name)
		}
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
		return fmt.Errorf("%s is not %s at quality %d: run go run ./cmd/hl levels backdrop jpeg %s", b.out, b.src, q, b.name)
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
