// Command art-review creates an offline sprite comparison at reference size.
//
// Usage:
//
//	go run ./tools/art-review -reference tile.png -candidate raw.png -out review
//
// The candidate must be at least 4x reference size and match its aspect ratio
// within 1%. Outputs are reference-masked 4x/2x working PNGs, a reference-size
// downsample, a reference/candidate/difference sheet (native row, nearest 4x
// row) and advisory JSON. Seven-hex patches are produced only for a 96x80
// reference, using the known repository layout. No world spacing is inferred
// for other reference dimensions. Set -reference-platform amiga-native for
// a native Amiga reference; the default provenance is repo-provisional.
// -resampling area (default) preserves the original area/box chain; bilinear
// and lanczos use alpha-aware antialiased filters and mask each raw level.
// Neighborhood recipes default to their box contract; source-backed entries
// can select Lanczos in the recipe. They still reject the -resampling flag.
// Inputs are never modified;
// existing output files are rejected. This is a contact-sheet utility, not
// a fidelity acceptance test or a claim that tile edges are seamless.
// -connected builds new-material road/stream/path/rail/trench overlays from explicit geometry and
// profiles, with reciprocal seven-cell patches at both runtime line parities.
// It requires a fresh output directory and excludes all other review modes.
// -shore builds the bounded summer shoreline family and grass/water bases,
// with pinned new materials and reciprocal patches at both runtime parities.
// -fog builds the bounded opaque unseen candidate and seasonal frontier patches.
// -ground builds five context-aware ground materials and fixed seasonal patches.
// -forest builds bounded whole-crown seasonal forest contexts and fixed patches.
// -hills builds bounded summer height-derived relief contexts and fixed patches.
// -unit builds the canonical runtime unit families and actual PNG terrain fixtures.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
)

const (
	referenceWidth  = 96
	referenceHeight = 80
	aspectTolerance = 0.01
	maxInputPixels  = 64 * 1024 * 1024
)

type dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type diagnostics struct {
	Resampling          string     `json:"resampling"`
	ResamplingAlgorithm string     `json:"resampling_algorithm"`
	ResamplingChain     []string   `json:"resampling_chain"`
	ReferencePlatform   string     `json:"reference_platform"`
	ReferencePath       string     `json:"reference_path"`
	CandidatePath       string     `json:"candidate_path"`
	ReferenceSize       dimensions `json:"reference_dimensions"`
	CandidateSize       dimensions `json:"candidate_dimensions"`
	Working4xSize       dimensions `json:"working_4x_dimensions"`
	Working2xSize       dimensions `json:"working_2x_dimensions"`
	SourceScaleX        float64    `json:"source_scale_x"`
	SourceScaleY        float64    `json:"source_scale_y"`
	AspectDeviation     float64    `json:"relative_aspect_deviation"`
	AspectTolerance     float64    `json:"aspect_tolerance"`
	RawCoverage         coverage   `json:"unmasked_native_coverage"`
	FinalCoverage       coverage   `json:"masked_native_coverage"`
	PatchCoverage       *coverage  `json:"seven_hex_coverage,omitempty"`
	ComparisonOrder     []string   `json:"comparison_panel_order"`
	DisplayScale        int        `json:"nearest_display_scale"`
	RGBMetric           string     `json:"rgb_metric"`
	AlphaMetric         string     `json:"alpha_metric"`
	DifferencePanel     string     `json:"difference_panel"`
	FootprintPolicy     string     `json:"footprint_policy"`
	PatchPlacement      string     `json:"patch_placement,omitempty"`
	Advisory            string     `json:"advisory"`
}

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "art-review:", err)
		os.Exit(1)
	}
}

func run(args []string, log io.Writer) error {
	fs := flag.NewFlagSet("art-review", flag.ContinueOnError)
	fs.SetOutput(log)
	referencePath := fs.String("reference", "", "reference PNG; seven-hex layout available only for 96x80")
	referencePlatform := fs.String("reference-platform", "repo-provisional", "reference provenance label (e.g. repo-provisional or amiga-native); supplied by caller")
	neighborhoodPath := fs.String("neighborhood", "", "JSON recipe for original Amiga seven-cell neighborhood review")
	mixedPath := fs.String("mixed", "", "bounded new-material bridge/crossing manifest; requires fresh -out")
	connectedPath := fs.String("connected", "", "new-material connected road/stream/path/rail/trench JSON manifest; requires fresh -out")
	compoundPath := fs.String("compound", "", "bounded compound scene manifest; requires fresh -out")
	fogPath := fs.String("fog", "", "bounded opaque unseen candidate manifest; requires fresh -out")
	groundPath := fs.String("ground", "", "bounded context-aware ground material manifest; requires fresh -out")
	forestPath := fs.String("forest", "", "bounded whole-crown seasonal forest manifest; requires fresh -out")
	layeredForestPath := fs.String("forest-layered", "", "layered summer forest floor, crowns, shadows and meadow props; requires fresh -out")
	hillsPath := fs.String("hills", "", "bounded summer height-derived hills manifest; requires fresh -out")
	shorePath := fs.String("shore", "", "bounded new-material summer shoreline manifest; requires fresh -out")
	unitPath := fs.String("unit", "", "six-facing canonical runtime unit family manifest; requires fresh -out")
	portraitPath := fs.String("portrait", "", "opaque canonical runtime unit family portrait manifest; requires fresh -out")
	packPath := fs.String("pack", "", "candidate-only JSON pack manifest; fixed Lanczos3 rectangle exports, requires fresh -out")
	candidatePath := fs.String("candidate", "", "candidate PNG, at least 4x reference dimensions and aspect within 1%")
	resampling := fs.String("resampling", "area", "single-sprite reduction: area, bilinear (antialiased tent), or lanczos (antialiased Lanczos3); excluded by -neighborhood")
	out := fs.String("out", "", "output directory (existing artifact names are rejected)")
	root := fs.String("root", "", "directory manifest source paths are relative to (default: the repository root above the manifest)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	sourceRoot = *root
	portraitProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "portrait" {
			portraitProvided = true
		}
	})
	if portraitProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "portrait" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *portraitPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-portrait requires -out and excludes all other flags and positional arguments")
		}
		return runPortrait(*portraitPath, *out, log)
	}
	unitProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "unit" {
			unitProvided = true
		}
	})
	if unitProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "unit" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *unitPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-unit requires -out and excludes all other flags and positional arguments")
		}
		return runUnit(*unitPath, *out, log)
	}
	hillsProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "hills" {
			hillsProvided = true
		}
	})
	if hillsProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "hills" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *hillsPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-hills requires -out and excludes all other flags and positional arguments")
		}
		return runHills(*hillsPath, *out, log)
	}
	if *layeredForestPath != "" {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "forest-layered" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *out == "" || fs.NArg() != 0 {
			return errors.New("-forest-layered requires -out and excludes other flags and positional arguments")
		}
		return runLayeredForest(*layeredForestPath, *out, log)
	}
	forestProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "forest" {
			forestProvided = true
		}
	})
	if forestProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "forest" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *forestPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-forest requires -out and excludes all other flags and positional arguments")
		}
		return runForest(*forestPath, *out, log)
	}
	groundProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "ground" {
			groundProvided = true
		}
	})
	if groundProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "ground" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *groundPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-ground requires -out and excludes all other flags and positional arguments")
		}
		return runGround(*groundPath, *out, log)
	}
	fogProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "fog" {
			fogProvided = true
		}
	})
	if fogProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "fog" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *fogPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-fog requires -out and excludes all other flags and positional arguments")
		}
		return runFog(*fogPath, *out, log)
	}
	shoreProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "shore" {
			shoreProvided = true
		}
	})
	if shoreProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "shore" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *shorePath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-shore requires -out and excludes all other flags and positional arguments")
		}
		return runShore(*shorePath, *out, log)
	}
	compoundProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "compound" {
			compoundProvided = true
		}
	})
	if compoundProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "compound" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *compoundPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-compound requires -out and excludes all other flags and positional arguments")
		}
		return runCompound(*compoundPath, *out, log)
	}
	mixedProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "mixed" {
			mixedProvided = true
		}
	})
	if mixedProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "mixed" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *mixedPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-mixed requires -out and excludes all other flags and positional arguments")
		}
		return runMixed(*mixedPath, *out, log)
	}
	connectedProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "connected" {
			connectedProvided = true
		}
	})
	if connectedProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "connected" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *connectedPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-connected requires -out and excludes all other flags and positional arguments")
		}
		return runConnected(*connectedPath, *out, log)
	}
	packProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "pack" {
			packProvided = true
		}
	})
	if packProvided {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "pack" && f.Name != "out" && f.Name != "root" {
				conflict = true
			}
		})
		if conflict || *packPath == "" || *out == "" || fs.NArg() != 0 {
			return errors.New("-pack requires -out and excludes reference/candidate/neighborhood/resampling flags and positional arguments")
		}
		return runPack(*packPath, *out, log)
	}
	if *neighborhoodPath != "" {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == "reference" || f.Name == "candidate" || f.Name == "reference-platform" || f.Name == "resampling" {
				conflict = true
			}
		})
		if conflict || *out == "" || fs.NArg() != 0 {
			return errors.New("-neighborhood requires -out and excludes reference/candidate/resampling flags and positional arguments")
		}
		return runNeighborhood(*neighborhoodPath, *out, log)
	}
	if *resampling != "area" && *resampling != "bilinear" && *resampling != "lanczos" {
		return errors.New("-resampling must be area, bilinear, or lanczos")
	}
	if *referencePath == "" || *candidatePath == "" || *out == "" || *referencePlatform == "" || fs.NArg() != 0 {
		return errors.New("required: -reference PNG -candidate PNG -out directory; no positional arguments")
	}
	ref, err := loadPNG(*referencePath)
	if err != nil {
		return fmt.Errorf("reference: %w", err)
	}
	rw, rh := ref.Bounds().Dx(), ref.Bounds().Dy()
	if int64(rw)*int64(rh) > maxInputPixels/60 {
		return errors.New("reference is too large for the comparison sheet (60x reference pixel count must fit 64 megapixels)")
	}
	if measure(ref, ref).ReferencePixels == 0 {
		return errors.New("reference has no visible alpha footprint")
	}
	src, err := loadPNG(*candidatePath)
	if err != nil {
		return fmt.Errorf("candidate: %w", err)
	}
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	deviation := math.Abs((float64(sw)/float64(sh))/(float64(rw)/float64(rh)) - 1)
	if deviation > aspectTolerance {
		return fmt.Errorf("candidate aspect deviation %.4f exceeds %.2f tolerance; refusing to stretch", deviation, aspectTolerance)
	}
	if sw < 4*rw || sh < 4*rh {
		return fmt.Errorf("candidate must be at least %dx%d (4x reference); upscaling is not supported", 4*rw, 4*rh)
	}
	resize := resizeArea
	switch *resampling {
	case "bilinear":
		resize = resizeBilinear
	case "lanczos":
		resize = resizeLanczos
	}
	raw4x, err := resize(src, 4*rw, 4*rh)
	if err != nil {
		return err
	}
	var raw2x, rawNative, working2x, native *image.NRGBA
	working4x := applyFootprint(raw4x, ref)
	algorithm := "Integer rational-coordinate area overlaps; alpha-weighted byte RGB, straight RGB output, round half up after each reduction."
	chain := []string{"source -> raw4x: area", "raw4x -> raw2x -> rawNative: box factor 2 at each step", "raw4x -> working4x: apply reference footprint", "working4x -> working2x -> native: box factor 2 then apply reference footprint at each step"}
	if *resampling != "area" {
		raw2x, err = resize(raw4x, 2*rw, 2*rh)
		if err != nil {
			return err
		}
		rawNative, err = resize(raw4x, rw, rh)
		if err != nil {
			return err
		}
		working2x = applyFootprint(raw2x, ref)
		native = applyFootprint(rawNative, ref)
		algorithm = "Separable antialiased tent, radius source/destination per axis; pixel centers (d+0.5)*source/destination-0.5. Clip support to image and renormalize. Integer-coordinate weights, row-major float64 sums; alpha-weighted byte RGB (no linear-light conversion), unpremultiply and round half up once per reduction; zero RGB at zero output alpha."
		chain = []string{"source -> raw4x: bilinear", "raw4x -> raw2x: bilinear", "raw4x -> rawNative: bilinear", "raw4x/raw2x/rawNative -> outputs: apply reference footprint independently"}
		if *resampling == "lanczos" {
			algorithm = "Separable antialiased Lanczos3, kernel sinc(t)*sinc(t/3), radius 3 widened by source/destination per axis; pixel centers (d+0.5)*source/destination-0.5. Clip support to image and renormalize. Integer-coordinate distances, Go math.Sin kernel, row-major float64 sums without intermediate quantization; premultiplied byte RGB (no linear-light conversion). For positive raw alpha, clamp premultiplied channels to [0,255*raw alpha] before unpremultiplying; independently clamp alpha to [0,255], preserving constant straight color through alpha overshoot. Round half up once per reduction; zero RGB at zero output alpha."
			chain = []string{"source -> raw4x: lanczos", "raw4x -> raw2x: lanczos", "raw4x -> rawNative: lanczos", "raw4x/raw2x/rawNative -> outputs: apply reference footprint independently"}
		}
	} else {
		raw2x, err = boxDownsample(raw4x, 2)
		if err != nil {
			return err
		}
		rawNative, err = boxDownsample(raw2x, 2)
		if err != nil {
			return err
		}
		working2x, err = boxDownsample(working4x, 2)
		if err != nil {
			return err
		}
		working2x = applyFootprint(working2x, ref)
		native, err = boxDownsample(working2x, 2)
		if err != nil {
			return err
		}
		native = applyFootprint(native, ref)
	}
	report := diagnostics{
		Resampling: *resampling, ResamplingAlgorithm: algorithm, ResamplingChain: chain,
		ReferencePlatform: *referencePlatform, ReferencePath: *referencePath, CandidatePath: *candidatePath,
		ReferenceSize: dimensions{rw, rh}, CandidateSize: dimensions{sw, sh},
		Working4xSize: dimensions{4 * rw, 4 * rh}, Working2xSize: dimensions{2 * rw, 2 * rh},
		SourceScaleX: float64(sw) / float64(rw), SourceScaleY: float64(sh) / float64(rh),
		AspectDeviation: deviation, AspectTolerance: aspectTolerance,
		RawCoverage: measure(ref, rawNative), FinalCoverage: measure(ref, native),
		ComparisonOrder: []string{"reference", "candidate", "difference"}, DisplayScale: 4,
		RGBMetric:       "Mean absolute straight RGB difference / 255 over shared visible pixels, weighted by min(reference alpha, candidate alpha); null if no shared coverage. Smaller means closer colors, not higher artistic fidelity.",
		AlphaMetric:     "Mean absolute alpha difference / 255 over the whole image. Missing/extra counts use alpha > 0. Patch gaps count only inside reference patch coverage, excluding the exposed outer boundary.",
		DifferencePanel: "Absolute premultiplied RGB difference; each channel is at least the absolute alpha difference. Opaque where either input is visible, transparent elsewhere.",
		FootprintPolicy: "Reference alpha is repeated at integer working scales; candidate alpha is capped to it. Candidate holes remain unfilled. Unmasked diagnostics expose discarded coverage.",
		Advisory:        "Advisory measurements only; not an acceptance score or proof of seamless tile edges. Visually review reference-size and nearest magnified sheets. Reference provenance is supplied by the caller; the utility does not authenticate it.",
	}
	type artifact struct {
		name string
		img  image.Image
	}
	outputs := []artifact{
		{"candidate-4x.png", working4x}, {"candidate-2x.png", working2x}, {"downsample.png", native},
		{"comparison.png", comparison(ref, native)},
	}
	if rw == referenceWidth && rh == referenceHeight {
		referencePatch, candidatePatch := sevenHex(ref), sevenHex(native)
		patchCoverage := measure(referencePatch, candidatePatch)
		report.PatchCoverage = &patchCoverage
		report.PatchPlacement = "x=col*142+(line%2)*71, y=line*40; center (line=2,col=1), six neighbors; blit in ascending line then column; cropped bounding box 238x240. This is repository geometry, even when a caller provides another provenance label."
		outputs = append(outputs, artifact{"seven-hex-reference.png", referencePatch}, artifact{"seven-hex-candidate.png", candidatePatch})
	}
	outputNames := []string{"diagnostics.json"}
	for _, output := range outputs {
		outputNames = append(outputNames, output.name)
	}
	for _, name := range outputNames {
		if _, err := os.Lstat(filepath.Join(*out, name)); !errors.Is(err, os.ErrNotExist) {
			if err == nil {
				return fmt.Errorf("output already exists: %s", name)
			}
			return err
		}
	}
	if err := os.MkdirAll(*out, 0755); err != nil {
		return err
	}
	for _, output := range outputs {
		if err := savePNG(filepath.Join(*out, output.name), output.img); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := saveExclusive(filepath.Join(*out, "diagnostics.json"), func(w io.Writer) error { _, err := w.Write(append(data, '\n')); return err }); err != nil {
		return err
	}
	_, err = fmt.Fprintf(log, "Wrote %d PNGs and diagnostics.json to %s (advisory only; reference platform %s).\n", len(outputs), *out, *referencePlatform)
	return err
}

func loadPNG(path string) (*image.NRGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	config, err := png.DecodeConfig(f)
	if err != nil {
		return nil, err
	}
	if config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > maxInputPixels {
		return nil, errors.New("PNG dimensions are invalid or exceed 64 megapixels")
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	src, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	img := image.NewNRGBA(image.Rect(0, 0, config.Width, config.Height))
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			img.SetNRGBA(x, y, color.NRGBAModel.Convert(src.At(x+src.Bounds().Min.X, y+src.Bounds().Min.Y)).(color.NRGBA))
		}
	}
	return img, nil
}

func savePNG(path string, img image.Image) error {
	return saveExclusive(path, func(w io.Writer) error { return png.Encode(w, img) })
}

func saveExclusive(path string, write func(io.Writer) error) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	writeErr := write(f)
	return errors.Join(writeErr, f.Close())
}
