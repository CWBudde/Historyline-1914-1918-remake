// Package hlart keeps the art catalogue: art/<kind>/<family>/prompt.md with its
// masters, the runtime files in assets/index.json and the generated art/index.json
// that the art book (docs/artbook) reads.
//
//	hl art index               write art/index.json (families without prompt.md are skipped)
//	hl art check               fail on any prompt, master, runtime or index problem
//	hl art thumbs -out DIR     write downscaled master thumbnails for the art book
//	hl art pulse -out DIR      render the PULSE level with the runtime art for the art book
//	hl art map -level FILE[:MODE] -out DIR  render another level (a scenario file's setup MODE) and a native detail crop
package hlart

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Run maintains the art catalogue or renders its thumbnails and maps.
func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: hl art index|check|thumbs|pulse|map [-root DIR]")
	}
	fl := flag.NewFlagSet("hl art "+args[0], flag.ContinueOnError)
	fl.SetOutput(stderr)
	root := fl.String("root", ".", "repository root")
	out := fl.String("out", "", "thumbs, pulse, map: output directory")
	levelPath := fl.String("level", "", "map: level JSON path relative to repository root, or a scenario file with :cp, :ec or :2p")
	limit := fl.Int("max", 640, "thumbs: longest side in pixels")
	if err := fl.Parse(args[1:]); err != nil {
		return err
	}
	if fl.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fl.Args())
	}
	switch args[0] {
	case "index":
		idx, problems, err := build(*root)
		if err != nil {
			return err
		}
		for _, p := range problems {
			say(stderr, "warning: %s\n", p)
		}
		raw, err := encode(idx)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(*root, "art", "index.json"), raw, 0o644); err != nil {
			return err
		}
		say(stdout, "art/index.json: %d families\n", len(idx.Assets))
		return nil
	case "check":
		problems, err := check(*root)
		if err != nil {
			return err
		}
		for _, p := range problems {
			say(stdout, "%s\n", p)
		}
		if len(problems) > 0 {
			return fmt.Errorf("%d problems", len(problems))
		}
		say(stdout, "art: ok\n")
		return nil
	case "thumbs":
		if *out == "" {
			return fmt.Errorf("thumbs needs -out")
		}
		idx, _, err := build(*root)
		if err != nil {
			return err
		}
		n, err := thumbs(*root, *out, idx, *limit)
		if err != nil {
			return err
		}
		say(stdout, "%s: %d thumbnails written\n", *out, n)
		return nil
	case "pulse":
		if *out == "" {
			return fmt.Errorf("pulse needs -out")
		}
		if err := pulse(*root, *out); err != nil {
			return err
		}
		say(stdout, "%s: PULSE map rendered\n", *out)
		return nil
	case "map":
		if *out == "" || *levelPath == "" {
			return fmt.Errorf("map needs -level and -out")
		}
		if err := renderLevel(*root, *out, *levelPath, "map", true); err != nil {
			return err
		}
		say(stdout, "%s: %s rendered\n", *out, *levelPath)
		return nil
	}
	return fmt.Errorf("unknown command %q, want index, check, thumbs, pulse or map", args[0])
}

func say(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}
