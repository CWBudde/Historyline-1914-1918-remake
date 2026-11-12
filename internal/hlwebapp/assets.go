package hlwebapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hlasset"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/pixfont/bitmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/CWBudde/Historyline-1914-1918-remake/levels"
)

type AssetIndex struct {
	Schema      int                            `json:"schema"`
	Rules       map[string]FileRef             `json:"rules"`
	Levels      map[string]LevelAssets         `json:"levels"`
	Packs       map[string]map[int]QualityPair `json:"packs"`
	PixelFonts  map[int]PixelFont              `json:"pixelFonts"`
	Fonts       map[string]FileRef             `json:"fonts"`
	FontLicense FileRef                        `json:"fontLicense"`
}
type LevelAssets struct {
	Files  map[string]FileRef        `json:"files"`
	Layers map[scenario.Mode]FileRef `json:"layers"`
	// BackdropFile is the painting's tile pyramid as a file of its own: inline,
	// every month's pyramid made up nearly all of index.json.
	BackdropFile *FileRef `json:"backdropFile,omitempty"`
}
type PixelFont struct {
	Image   FileRef `json:"image"`
	Metrics FileRef `json:"metrics"`
}

// Assets exports the browser's catalog from existing authored assets. The index
// is published last; interrupted exports leave the previous index usable.
func Assets(ctx context.Context, args []string, stdout, stderr io.Writer) (err error) {
	f := flag.NewFlagSet("hl webapp assets", flag.ContinueOnError)
	f.SetOutput(stderr)
	out := f.String("out", "build/webapp/data", "output directory")
	packs := f.String("packs", "", "reuse an existing validated native pack catalog (default: update build/packs)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("hl webapp assets: expected flags only")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if *packs == "" {
		*packs = "build/packs"
		if err := hlasset.Run([]string{"packs", "-out", *packs}, stdout, stderr); err != nil {
			return err
		}
	}
	b := blobWriter{root: *out}
	index := AssetIndex{Schema: 1, Rules: map[string]FileRef{}, Levels: map[string]LevelAssets{}, Packs: map[string]map[int]QualityPair{}, PixelFonts: map[int]PixelFont{}, Fonts: map[string]FileRef{}}
	var terrain map[string]bool
	for _, name := range []string{"units", "terrain", "screens"} {
		if err := ctx.Err(); err != nil {
			return err
		}
		catalog, ids, err := b.catalog(ctx, filepath.Join(*packs, name), name)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		index.Packs[name] = catalog
		if name == "terrain" {
			terrain = ids
		}
		if _, err := fmt.Fprintf(stdout, "exported %s: %d IDs\n", name, len(ids)); err != nil {
			return err
		}
	}
	// The native interface pack is separate from its three map/menu catalogs.
	tmp, err := os.MkdirTemp("", "historyline-webapp-interface-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmp)) }()
	for _, d := range []int{1, 2, 4} {
		if err := hlasset.Run([]string{"build", "-density", strconv.Itoa(d), "-out", filepath.Join(tmp, fmt.Sprintf("%dx", d))}, stdout, stderr); err != nil {
			return err
		}
	}
	if index.Packs["interface"], _, err = b.catalog(ctx, tmp, "interface"); err != nil {
		return err
	}
	for _, path := range []string{"internal/data/terrain.json", "internal/data/units.json", "internal/data/weapons.json", "internal/data/notes.json", "internal/dosrules/full.json", "internal/dosrules/demo.json", "internal/dosrules/extension.json", "internal/dosbind/records.json", "internal/origview/looks.json", "assets/interface/animations.json", "assets/interface/bindings.json"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		ref, err := b.put("json", raw)
		if err != nil {
			return err
		}
		index.Rules[filepath.Base(path)] = ref
	}
	if err := writePWA(*out); err != nil {
		return err
	}
	set, err := data.Load()
	if err != nil {
		return err
	}
	type sprite struct {
		Family string `json:"family"`
		Long   string `json:"long,omitempty"`
	}
	sprites := map[string]sprite{}
	for id := range set.Units {
		family, ok := unitsprite.Family(id)
		if !ok {
			return fmt.Errorf("no sprite family: %s", id)
		}
		long, _ := unitsprite.Long(family)
		sprites[id] = sprite{family, long}
	}
	if index.Rules["sprites.json"], err = b.json(sprites); err != nil {
		return err
	}
	water, err := newWaterBaker(filepath.Join(*packs, "terrain"), b)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, water.close()) }()
	months, err := scenario.Months(levels.Levels)
	if err != nil {
		return err
	}
	for _, month := range months {
		if err := ctx.Err(); err != nil {
			return err
		}
		entry := LevelAssets{Files: map[string]FileRef{}, Layers: map[scenario.Mode]FileRef{}}
		files, err := fs.Glob(levels.Levels, month+"/*.json")
		if err != nil {
			return err
		}
		for _, path := range files {
			raw, err := levels.Levels.ReadFile(path)
			if err != nil {
				return err
			}
			ref, err := b.put("json", raw)
			if err != nil {
				return err
			}
			entry.Files[filepath.Base(path)] = ref
		}
		sc, err := scenario.Read(levels.Levels, month)
		if err != nil {
			return err
		}
		for _, mode := range scenario.Modes {
			lv, err := sc.Level(mode)
			if err != nil {
				return err
			}
			layers, err := b.layers(ctx, lv, func(id string) bool { return terrain[id] })
			if err != nil {
				return err
			}
			if err := water.bake(ctx, lv, &layers); err != nil {
				return err
			}
			mixed, err := water.bakeSeasons(ctx, lv, &layers)
			if err != nil {
				return err
			}
			if len(mixed.Layers) > 0 {
				layers.SeasonBakes = &mixed
			}
			if entry.Layers[mode], err = b.json(layers); err != nil {
				return err
			}
		}
		raw, err := levels.Levels.ReadFile(month + "/backdrop.jpg")
		if err == nil {
			im, _, err := image.Decode(bytes.NewReader(raw))
			if err != nil {
				return err
			}
			ref, err := b.backdrop(ctx, im)
			if err != nil {
				return err
			}
			entry.BackdropFile = &ref
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		index.Levels[month] = entry
		if _, err := fmt.Fprintf(stdout, "exported %s: %d setups\n", month, len(entry.Layers)); err != nil {
			return err
		}
	}
	for _, height := range []int{5, 7} {
		im, metrics := bitmap.Atlas(height)
		pixel := PixelFont{}
		if pixel.Image, err = b.png(im); err != nil {
			return err
		}
		if pixel.Metrics, err = b.json(metrics); err != nil {
			return err
		}
		index.PixelFonts[height] = pixel
	}
	for _, name := range []string{"NotoSerifDisplay-Bold", "NotoSans-Regular", "NotoSans-SemiBold", "NotoSerif-Regular", "NotoSerif-Italic"} {
		raw, err := os.ReadFile(filepath.Join("cmd/historyline/fonts", name+".ttf"))
		if err != nil {
			return err
		}
		font, err := woff2(raw)
		if err != nil {
			return err
		}
		if index.Fonts[name], err = b.put("woff2", font); err != nil {
			return err
		}
	}
	license, err := os.ReadFile("cmd/historyline/fonts/LICENSE")
	if err != nil {
		return err
	}
	if index.FontLicense, err = b.put("txt", license); err != nil {
		return err
	}
	raw, err := json.Marshal(index)
	if err != nil {
		return err
	}
	ref, err := b.put("json", raw)
	if err != nil {
		return err
	}
	// Rename a complete copy, rather than exposing a partially written index.
	indexPath := filepath.Join(*out, "index.json")
	next, err := os.CreateTemp(*out, ".index-")
	if err != nil {
		return err
	}
	defer func() {
		if e := os.Remove(next.Name()); e != nil && !errors.Is(e, fs.ErrNotExist) {
			err = errors.Join(err, e)
		}
	}()
	if _, err = next.Write(raw); err != nil {
		return errors.Join(err, next.Close())
	}
	if err := next.Close(); err != nil {
		return err
	}
	if err := os.Rename(next.Name(), indexPath); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "wrote %d levels to %s (index %s)\n", len(index.Levels), *out, ref.SHA256)
	return err
}
