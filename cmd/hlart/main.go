// Command hlart keeps the art catalogue: art/<kind>/<family>/prompt.md with its
// masters, the runtime files in assets/index.json and the generated art/index.json
// that the art book (docs/artbook) reads.
//
//	hlart index               write art/index.json (families without prompt.md are skipped)
//	hlart check               fail on any prompt, master, runtime or index problem
//	hlart thumbs -out DIR     write downscaled master thumbnails for the art book
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "hlart:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: hlart index|check|thumbs [-root DIR]")
	}
	fl := flag.NewFlagSet("hlart "+args[0], flag.ContinueOnError)
	fl.SetOutput(stderr)
	root := fl.String("root", ".", "repository root")
	out := fl.String("out", "", "thumbs: output directory")
	limit := fl.Int("max", 640, "thumbs: longest side in pixels")
	if err := fl.Parse(args[1:]); err != nil {
		return err
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
	}
	return fmt.Errorf("unknown command %q, want index, check or thumbs", args[0])
}

func say(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}
