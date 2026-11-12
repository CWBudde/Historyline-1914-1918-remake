package hlwebapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
)

// pwaIconSource is the runtime picture the app icons are scaled from.
const pwaIconSource = "assets/portraits/a7v/portrait-4x.png"

// pwaBackground is the page colour, the manifest's background and theme.
var pwaBackground = color.RGBA{0x16, 0x26, 0x30, 0xff}

type pwaIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
}

type pwaManifest struct {
	Name            string    `json:"name"`
	ShortName       string    `json:"short_name"`
	Display         string    `json:"display"`
	Orientation     string    `json:"orientation"`
	StartURL        string    `json:"start_url"`
	Scope           string    `json:"scope"`
	BackgroundColor string    `json:"background_color"`
	ThemeColor      string    `json:"theme_color"`
	Icons           []pwaIcon `json:"icons"`
}

// writePWA writes the web app manifest and its icons into out. All URLs are
// relative: the site is served from a sub-path.
func writePWA(out string) error {
	file, err := os.Open(pwaIconSource)
	if err != nil {
		return err
	}
	defer file.Close()
	src, err := png.Decode(file)
	if err != nil {
		return fmt.Errorf("%s: %w", pwaIconSource, err)
	}
	b := src.Bounds()
	side := min(b.Dx(), b.Dy())
	// Square crop, left of centre: the picture's subject sits there.
	x0 := b.Min.X + (b.Dx()-side)/4
	square := image.Rect(x0, b.Min.Y, x0+side, b.Min.Y+side)

	type spec struct {
		name  string
		size  int
		inset bool
	}
	for _, s := range []spec{
		{"icon-192.png", 192, false},
		{"icon-512.png", 512, false},
		{"icon-maskable-512.png", 512, true},
		{"apple-touch-icon.png", 180, false},
	} {
		dst := image.NewRGBA(image.Rect(0, 0, s.size, s.size))
		art := dst.Bounds()
		if s.inset {
			// Keep the art inside the maskable safe zone (inner 80%).
			draw.Draw(dst, art, image.NewUniform(pwaBackground), image.Point{}, draw.Src)
			pad := s.size / 10
			art = art.Inset(pad)
		}
		xdraw.CatmullRom.Scale(dst, art, src, square, xdraw.Src, nil)
		var buf bytes.Buffer
		if err := png.Encode(&buf, dst); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, s.name), buf.Bytes(), 0o644); err != nil {
			return err
		}
	}

	manifest := pwaManifest{
		Name:            "Historyline 1914-1918",
		ShortName:       "Historyline",
		Display:         "standalone",
		Orientation:     "any",
		StartURL:        "./",
		Scope:           "./",
		BackgroundColor: "#162630",
		ThemeColor:      "#162630",
		Icons: []pwaIcon{
			{Src: "icon-192.png", Sizes: "192x192", Type: "image/png", Purpose: "any"},
			{Src: "icon-512.png", Sizes: "512x512", Type: "image/png", Purpose: "any"},
			{Src: "icon-maskable-512.png", Sizes: "512x512", Type: "image/png", Purpose: "maskable"},
		},
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "manifest.webmanifest"), append(raw, '\n'), 0o644)
}
