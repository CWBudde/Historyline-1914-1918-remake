// Command native-reference exports private visual references from local Amiga
// Historyline data. It does not distribute, modify, or interpret resource art.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "native-reference:", err)
		os.Exit(1)
	}
}

type artifact struct {
	Library, Name, Coverage string
	Offset, PackedBytes     int
	Graphic                 graphic
	SourceHash              string
}

func run(args []string, log io.Writer) error {
	fs := flag.NewFlagSet("native-reference", flag.ContinueOnError)
	fs.SetOutput(log)
	data := fs.String("data", "", "local Amiga data folder containing LIB/, 00.PAL and HEXMSK*.DAT")
	out := fs.String("out", "build/native-reference", "fresh output subdirectory inside this repository's ignored build/ or original/")
	paletteName := fs.String("palette", "00", "native palette: 00, 01, or 02")
	libraryNames := fs.String("libraries", "PARTS,PARTW,UNIT,BIGUNIT", "comma-separated native graphics libraries")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *data == "" || fs.NArg() != 0 {
		return errors.New("required: -data local-Amiga-data-folder; no positional arguments")
	}
	if *paletteName != "00" && *paletteName != "01" && *paletteName != "02" {
		return errors.New("palette must be 00, 01, or 02")
	}
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	output, err := privateOutput(root, *out)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(output); err == nil {
		return errors.New("output already exists; select a fresh private subdirectory")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	input, err := filepath.Abs(*data)
	if err != nil {
		return err
	}
	input, err = filepath.EvalSymlinks(input)
	if err != nil {
		return err
	}
	if within(input, output) || within(output, input) {
		return errors.New("input and output directories must not overlap")
	}
	sourceHashes := make(map[string]string)
	read := func(name string) ([]byte, error) {
		b, err := readBounded(filepath.Join(input, filepath.FromSlash(name)))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		sourceHashes[name] = fmt.Sprintf("%x", sha256.Sum256(b))
		return b, nil
	}
	palBytes, err := read(*paletteName + ".PAL")
	if err != nil {
		return err
	}
	pal, err := parsePalette(palBytes)
	if err != nil {
		return err
	}
	maskBytes, err := read("HEXMSK00.DAT")
	if err != nil {
		return err
	}
	inverseBytes, err := read("HEXMSK20.DAT")
	if err != nil {
		return err
	}
	mask, err := validateMasks(maskBytes, inverseBytes)
	if err != nil {
		return err
	}
	supported := map[string]bool{"PARTS": true, "PARTW": true, "UNIT": true, "BIGUNIT": true, "EXP": true, "UMAPDASA": true, "UMAPDAS2": true, "UMAPDAWA": true, "UMAPDAW2": true}
	seen := make(map[string]bool)
	var artifacts []artifact
	totalPixels := 0
	for _, library := range strings.Split(*libraryNames, ",") {
		if !supported[library] || seen[library] {
			return fmt.Errorf("unsupported or repeated graphics library %q", library)
		}
		seen[library] = true
		name := "LIB/" + library + ".LIB"
		packed, err := read(name)
		if err != nil {
			return err
		}
		resources, err := parseLibrary(packed)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		var coverage []byte
		coverageName := "record"
		if library == "PARTS" || library == "PARTW" {
			coverage = mask
			coverageName = "HEXMSK00"
		}
		for _, resource := range resources {
			g, err := decodeGraphic(resource.Data, pal, coverage)
			if err != nil {
				return fmt.Errorf("%s/%s: %w", library, resource.Name, err)
			}
			totalPixels += g.Width * g.Height
			if totalPixels > 16<<20 {
				return errors.New("selected graphics exceed 16 megapixel export limit")
			}
			actualCoverage := coverageName
			if coverage == nil {
				actualCoverage = "none"
				if g.DataOffset > 24 {
					actualCoverage = "embedded"
				}
			}
			artifacts = append(artifacts, artifact{Library: library, Name: resource.Name, Coverage: actualCoverage, Offset: resource.Offset, PackedBytes: len(resource.Data), Graphic: g, SourceHash: sourceHashes[name]})
		}
	}
	// All inputs are validated before creating any outputs.
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(output, 0755); err != nil {
		return err
	}
	rows := [][]string{{"library", "resource", "archive_offset", "stored_bytes", "unpacked_bytes", "width", "height", "row_bytes", "plane_mask", "pixel_data_offset", "coverage", "source_sha256", "png_sha256"}}
	created := make(map[string]bool)
	for _, a := range artifacts {
		dir := filepath.Join(output, a.Library)
		if !created[a.Library] {
			if err := os.Mkdir(dir, 0755); err != nil {
				return err
			}
			created[a.Library] = true
		}
		encoded, err := encodePNG(a.Graphic.Image)
		if err != nil {
			return err
		}
		if err := writeNew(filepath.Join(dir, a.Name+".png"), encoded); err != nil {
			return err
		}
		g := a.Graphic
		rows = append(rows, []string{a.Library, a.Name, strconv.Itoa(a.Offset), strconv.Itoa(a.PackedBytes), strconv.Itoa(g.RawBytes), strconv.Itoa(g.Width), strconv.Itoa(g.Height), strconv.Itoa(g.RowBytes), fmt.Sprintf("0x%04x", g.PlaneMask), strconv.Itoa(g.DataOffset), a.Coverage, a.SourceHash, fmt.Sprintf("%x", sha256.Sum256(encoded))})
	}
	maskImage := image.NewNRGBA(image.Rect(0, 0, 24, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			if mask[y*4+x/8]&(128>>uint(x%8)) != 0 {
				maskImage.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}
	encoded, err := encodePNG(maskImage)
	if err != nil {
		return err
	}
	if err := writeNew(filepath.Join(output, "HEXMSK00.png"), encoded); err != nil {
		return err
	}
	paletteImage := image.NewNRGBA(image.Rect(0, 0, 256, 128))
	for i, c := range pal {
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				paletteImage.SetNRGBA((i%8)*32+x, (i/8)*32+y, c)
			}
		}
	}
	encoded, err = encodePNG(paletteImage)
	if err != nil {
		return err
	}
	if err := writeNew(filepath.Join(output, "palette-"+*paletteName+".png"), encoded); err != nil {
		return err
	}
	var csvBuffer bytes.Buffer
	writer := csv.NewWriter(&csvBuffer)
	if err := writer.WriteAll(rows); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(output, "metadata.csv"), csvBuffer.Bytes()); err != nil {
		return err
	}
	var names []string
	for name := range sourceHashes {
		names = append(names, name)
	}
	sort.Strings(names)
	var manifest strings.Builder
	for _, name := range names {
		fmt.Fprintf(&manifest, "%s  %s\n", sourceHashes[name], name)
	}
	if err := writeNew(filepath.Join(output, "source-sha256.txt"), []byte(manifest.String())); err != nil {
		return err
	}
	_, err = fmt.Fprintf(log, "exported %d native resource PNGs and palette/mask references to %s\n", len(artifacts), output)
	return err
}

func readBounded(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	b, readErr := io.ReadAll(io.LimitReader(f, maxRawBytes+1))
	closeErr := f.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(b) > maxRawBytes {
		return nil, errors.New("input exceeds 32 MiB limit")
	}
	return b, nil
}
func encodePNG(im image.Image) ([]byte, error) {
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func writeNew(path string, b []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(b)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func repositoryRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		_, gitErr := os.Stat(filepath.Join(dir, ".git"))
		_, modErr := os.Stat(filepath.Join(dir, "go.mod"))
		if gitErr == nil && modErr == nil {
			return filepath.EvalSymlinks(dir)
		}
		if filepath.Dir(dir) == dir {
			return "", errors.New("run inside the repository containing go.mod and .git")
		}
	}
}
func within(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func privateOutput(root, path string) (string, error) {
	if path == "" {
		return "", errors.New("output path is empty")
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(canonicalRoot, candidate)
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	// Resolve existing parents too: a build/ symlink must not redirect writes
	// into tracked sources or outside the repository.
	ancestor := candidate
	var suffix []string
	for {
		_, err := os.Lstat(ancestor)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", errors.New("cannot resolve output ancestor")
		}
		suffix = append(suffix, filepath.Base(ancestor))
		ancestor = parent
	}
	resolved, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", err
	}
	for i := len(suffix) - 1; i >= 0; i-- {
		resolved = filepath.Join(resolved, suffix[i])
	}
	if !within(canonicalRoot, resolved) {
		return "", errors.New("output must stay inside the repository's private directories")
	}
	rel, err := filepath.Rel(canonicalRoot, resolved)
	if err != nil {
		return "", err
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) < 2 || (parts[0] != "build" && parts[0] != "original") {
		return "", errors.New("output must be a subdirectory of ignored build/ or original/")
	}
	return resolved, nil
}

func validateMasks(input, inverse []byte) ([]byte, error) {
	mask, used, err := unpackBE(input)
	if err != nil {
		return nil, err
	}
	if used != len(input) || len(mask) != 96 {
		return nil, errors.New("HEXMSK00 must decode to one 24x24 word-padded plane (96 bytes)")
	}
	inv, used, err := unpackBE(inverse)
	if err != nil {
		return nil, err
	}
	if used != len(inverse) || len(inv) != 96 {
		return nil, errors.New("HEXMSK20 must decode to 96 bytes")
	}
	widths := []int{8, 10, 10, 12, 14, 16, 16, 18, 20, 22, 22, 24, 24, 22, 22, 20, 18, 16, 16, 14, 12, 10, 10, 8}
	for y, width := range widths {
		for x := 0; x < 32; x++ {
			set := mask[y*4+x/8]&(128>>uint(x%8)) != 0
			want := x >= (24-width)/2 && x < (24+width)/2
			if set != want {
				return nil, errors.New("HEXMSK00 is not the verified native Amiga 24x24 hex shape")
			}
		}
	}
	for i := range mask {
		if mask[i]^inv[i] != 255 {
			return nil, errors.New("HEXMSK20 must be the exact inverse of HEXMSK00")
		}
	}
	return mask, nil
}
