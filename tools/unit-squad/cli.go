package unitsquad

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

// Run composes a unit atlas with the requested number and scale of soldiers.
func Run(args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("hl export unit-squad", flag.ContinueOnError)
	f.SetOutput(stderr)
	src := f.String("src", "", "source atlas with three men per cell")
	n := f.Int("n", 6, "men per cell, 1..6; 0 scales the source cells as they are")
	out := f.String("out", "", "atlas to write")
	scale := f.Float64("scale", .52, "size of a man relative to the source")
	lone := f.Bool("gunner", false, "first drop the loader of a two-man team (use with -n 0)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if *src == "" || *out == "" || f.NArg() != 0 {
		return fmt.Errorf("requires -src FILE -out FILE and no positional arguments")
	}
	file, err := os.Open(*src)
	if err != nil {
		return err
	}
	im, err := png.Decode(file)
	_ = file.Close()
	if err != nil {
		return err
	}
	nrgba := image.NewNRGBA(im.Bounds())
	draw.Draw(nrgba, nrgba.Bounds(), im, im.Bounds().Min, draw.Src)
	if *lone {
		if nrgba, err = gunners(nrgba); err != nil {
			return err
		}
	}
	res, err := atlas(nrgba, *n, *scale)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	o, err := os.Create(*out)
	if err != nil {
		return err
	}
	if err := png.Encode(o, res); err != nil {
		_ = o.Close()
		return err
	}
	return o.Close()
}
