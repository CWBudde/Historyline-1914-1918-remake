package hlcli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const typstFonts = "docs/manual/fonts"

// sub runs another hl command and turns a non-zero exit code into an error.
func sub(ctx context.Context, out, errs io.Writer, args ...string) error {
	if code := Run(ctx, args, strings.NewReader(""), out, errs); code != 0 {
		return fmt.Errorf("hl %s failed (exit %d)", strings.Join(args, " "), code)
	}
	return nil
}

// ensureBuildDir creates build/ and the build/go.mod that keeps stray Go files there out of ./...
func ensureBuildDir() error {
	if err := os.MkdirAll("build", 0o755); err != nil {
		return err
	}
	if _, err := os.Stat("build/go.mod"); err == nil {
		return nil
	}
	raw, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("run this repository-only command from the repository root: %w", err)
	}
	for line := range strings.SplitSeq(string(raw), "\n") {
		if v, ok := strings.CutPrefix(line, "go "); ok {
			return os.WriteFile("build/go.mod", []byte("module scratch\n\ngo "+strings.TrimSpace(v)+"\n"), 0o644)
		}
	}
	return errors.New("go.mod has no go directive")
}

func typst(ctx context.Context, out, errs io.Writer, args ...string) error {
	base := []string{"compile", "--root", ".", "--font-path", typstFonts}
	return runProcess(ctx, "", "typst", append(base, args...), nil, out, errs)
}

// artbook renders the art book from art/index.json into build/artbook/Historyline_Art.pdf.
func artbook(ctx context.Context, args []string, _ io.Reader, out, errs io.Writer) error {
	f := flag.NewFlagSet("hl artbook", flag.ContinueOnError)
	f.SetOutput(errs)
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	if err := ensureBuildDir(); err != nil {
		return err
	}
	if err := os.RemoveAll("build/artbook/shore"); err != nil {
		return err
	}
	steps := [][]string{
		{"review", "art", "-shore", "art/pipeline/shore-terrain-v2.json", "-out", "build/artbook/shore"},
		{"art", "pulse", "-out", "build/artbook/pulse"},
		{"review", "reference", "-look-out", "build/artbook/look"},
		{"art", "map", "-level", "levels/1914-10:cp", "-out", "build/artbook/winter"},
		{"art", "map", "-level", "levels/1914-12:cp", "-out", "build/artbook/winter3"},
		{"art", "map", "-level", "levels/1914-08:cp", "-out", "build/artbook/campaign/summer"},
		{"art", "map", "-level", "levels/1916-10:cp", "-out", "build/artbook/campaign/winter"},
		{"art", "map", "-level", "levels/1916-06:ec", "-out", "build/artbook/campaign/sea"},
		{"art", "map", "-level", "levels/1917-01:2p", "-out", "build/artbook/campaign/harbour"},
		{"review", "reference", "-compare", "build/artbook/compare", "-compare-format", "png"},
	}
	for _, s := range steps {
		if err := sub(ctx, out, errs, s...); err != nil {
			return err
		}
	}
	compare, style := "", "false"
	if _, err := os.Stat("build/artbook/compare/compare.json"); err == nil {
		compare = "/build/artbook/compare"
	}
	if _, err := os.Stat("art/style.md"); err == nil {
		style = "true"
	}
	const pdf = "build/artbook/Historyline_Art.pdf"
	err := typst(ctx, out, errs,
		"--input", "pulse=/build/artbook/pulse", "--input", "look=/build/artbook/look",
		"--input", "winter=/build/artbook/winter", "--input", "winter3=/build/artbook/winter3",
		"--input", "campaign=/build/artbook/campaign", "--input", "shore=/build/artbook/shore",
		"--input", "compare="+compare, "--input", "style="+style,
		"docs/artbook/artbook.typ", pdf)
	if err == nil {
		_, err = fmt.Fprintln(out, pdf)
	}
	return err
}

// manual builds the reconstructed manual, with the original's pictures when the local PDF is present.
func manual(ctx context.Context, args []string, _ io.Reader, out, errs io.Writer) error {
	f := flag.NewFlagSet("hl manual", flag.ContinueOnError)
	f.SetOutput(errs)
	imagesOnly := f.Bool("images", false, "only extract the original manual's pictures into build/manual/img")
	watch := f.Bool("watch", false, "rebuild on every change (extract the pictures once first with -images)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", f.Args())
	}
	if err := ensureBuildDir(); err != nil {
		return err
	}
	const (
		original = "original/extracted/manual/Historyline_1914-1918.pdf"
		pdf      = "build/manual/Historyline_1914-1918_Manual.pdf"
		source   = "docs/manual/manual.typ"
	)
	if err := os.MkdirAll("build/manual/img", 0o755); err != nil {
		return err
	}
	extract := func() error {
		return runProcess(ctx, "", "pdfimages", []string{"-all", "-p", original, "build/manual/img/img"}, nil, out, errs)
	}
	if *imagesOnly {
		return extract()
	}
	if *watch {
		return runProcess(ctx, "", "typst", []string{"watch", "--root", ".", "--font-path", typstFonts, "--input", "images=true", source, pdf}, nil, out, errs)
	}
	images := "false"
	if _, err := os.Stat(original); err == nil {
		if err := extract(); err != nil {
			return err
		}
		images = "true"
	}
	if err := typst(ctx, out, errs, "--input", "images="+images, source, pdf); err != nil {
		return err
	}
	_, err := fmt.Fprintf(out, "%s (images=%s)\n", pdf, images)
	return err
}
