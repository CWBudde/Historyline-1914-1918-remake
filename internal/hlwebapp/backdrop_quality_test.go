package hlwebapp

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/gen2brain/webp"
)

// paintedSample is a stand-in for a painted map: colour fields that drift
// slowly, a mid-frequency grain, fine speckle and a few hard strokes.
func paintedSample(w, h int) *image.NRGBA {
	rng := rand.New(rand.NewSource(3))
	grid := func(cell int) [][3]float64 {
		g := make([][3]float64, (w/cell+2)*(h/cell+2))
		for i := range g {
			base := rng.Float64()
			g[i] = [3]float64{base + 0.3*rng.Float64(), base + 0.3*rng.Float64(), base + 0.3*rng.Float64()}
		}
		return g
	}
	octaves := []struct {
		cell int
		amp  float64
		g    [][3]float64
	}{{64, 0.30, grid(64)}, {12, 0.16, grid(12)}, {4, 0.09, grid(4)}}
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			var c [3]float64
			for _, o := range octaves {
				stride := w/o.cell + 2
				fx, fy := float64(x)/float64(o.cell), float64(y)/float64(o.cell)
				x0, y0 := int(fx), int(fy)
				tx, ty := fx-float64(x0), fy-float64(y0)
				for k := range c {
					top := o.g[y0*stride+x0][k]*(1-tx) + o.g[y0*stride+x0+1][k]*tx
					bottom := o.g[(y0+1)*stride+x0][k]*(1-tx) + o.g[(y0+1)*stride+x0+1][k]*tx
					c[k] += o.amp * (top*(1-ty) + bottom*ty)
				}
			}
			// The grain is mostly in the lightness, a little in each colour.
			light, k := 0.07*(rng.Float64()-0.5), 0.05
			im.SetNRGBA(x, y, color.NRGBA{
				clamp8(60 + 190*(c[0]+light+k*(rng.Float64()-0.5))),
				clamp8(60 + 190*(c[1]+light+k*(rng.Float64()-0.5))),
				clamp8(60 + 190*(c[2]+light+k*(rng.Float64()-0.5))), 255,
			})
		}
	}
	for range 40 {
		x, y, dx, dy := rng.Intn(w), rng.Intn(h), rng.Intn(5)-2, rng.Intn(5)-2
		col := color.NRGBA{uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256)), 255}
		for i := range 60 {
			im.SetNRGBA(x+i*dx/2, y+i*dy/2, col)
		}
	}
	return im
}

func clamp8(v float64) uint8 { return uint8(math.Max(0, math.Min(255, v))) }

// decodeWebP decodes a lossy opaque WebP the way a browser's libwebp does: the
// planes are limited-range BT.601 and the chroma is upsampled with the 9-3-3-1
// weights. The decoder in the webp package returns the planes as they are, and
// image.YCbCr would read them as full-range.
func decodeWebP(t *testing.T, data []byte) *image.NRGBA {
	t.Helper()
	d, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	p, ok := d.(*image.NYCbCrA)
	if !ok || p.SubsampleRatio != image.YCbCrSubsampleRatio420 {
		t.Fatalf("decoded %T", d)
	}
	w, h := p.Rect.Dx(), p.Rect.Dy()
	cw, ch := (w+1)/2, (h+1)/2
	chroma := func(plane []uint8, x, y int) int {
		return int(plane[min(max(y, 0), ch-1)*p.CStride+min(max(x, 0), cw-1)])
	}
	up := func(plane []uint8, x, y int) int {
		nx, ny := x/2+(x&1)*2-1, y/2+(y&1)*2-1
		return (9*chroma(plane, x/2, y/2) + 3*chroma(plane, nx, y/2) + 3*chroma(plane, x/2, ny) + chroma(plane, nx, ny) + 8) >> 4
	}
	clip := func(v int) uint8 { return uint8(min(max(v>>6, 0), 255)) }
	out := image.NewNRGBA(p.Rect)
	for y := range h {
		for x := range w {
			yy, u, v := int(p.Y[y*p.YStride+x])*19077>>8, up(p.Cb, x, y), up(p.Cr, x, y)
			out.SetNRGBA(x, y, color.NRGBA{clip(yy + v*26149>>8 - 14234), clip(yy - u*6419>>8 - v*13320>>8 + 8708), clip(yy + u*33050>>8 - 17685), 255})
		}
	}
	return out
}

func psnr(a, b *image.NRGBA) float64 {
	var sum float64
	for y := a.Rect.Min.Y; y < a.Rect.Max.Y; y++ {
		for x := a.Rect.Min.X; x < a.Rect.Max.X; x++ {
			i, j := a.PixOffset(x, y), b.PixOffset(x, y)
			for k := range 3 {
				d := float64(a.Pix[i+k]) - float64(b.Pix[j+k])
				sum += d * d
			}
		}
	}
	return 10 * math.Log10(255*255/(sum/float64(3*a.Rect.Dx()*a.Rect.Dy())))
}

// TestBackdropTilesKeepTheSource pins the full-resolution level's fidelity: the
// tiles decoded and joined must stay close to the lossless source, and well
// above what the tiles used to be, the shipped JPEG (quality 95) encoded a
// second time at 85.
func TestBackdropTilesKeepTheSource(t *testing.T) {
	src := paintedSample(1100, 700)
	b := blobWriter{root: t.TempDir()}
	ref, err := b.backdrop(context.Background(), src)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(b.root, ref.Path))
	if err != nil {
		t.Fatal(err)
	}
	var p Pyramid
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	joined := image.NewNRGBA(src.Rect)
	for _, tile := range p.Mips[0].Tiles {
		data, err := os.ReadFile(filepath.Join(b.root, tile.File.Path))
		if err != nil {
			t.Fatal(err)
		}
		got := decodeWebP(t, data)
		for y := range tile.Height {
			for x := range tile.Width {
				joined.SetNRGBA(tile.X+x, tile.Y+y, got.NRGBAAt(tile.X-tile.SourceX+x, tile.Y-tile.SourceY+y))
			}
		}
	}
	got := psnr(src, joined)

	var shipped, again bytes.Buffer
	if err := jpeg.Encode(&shipped, src, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	once, err := jpeg.Decode(&shipped)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(&again, once, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatal(err)
	}
	twice, err := jpeg.Decode(&again)
	if err != nil {
		t.Fatal(err)
	}
	old := image.NewNRGBA(src.Rect)
	for y := range 700 {
		for x := range 1100 {
			old.Set(x, y, twice.At(x, y))
		}
	}
	before := psnr(src, old)
	t.Logf("full-resolution tiles: %.2f dB, shipped JPEG encoded again: %.2f dB", got, before)
	if got < minBackdropPSNR {
		t.Errorf("tiles reach %.2f dB, want at least %.1f", got, minBackdropPSNR)
	}
	if got < before+1.5 {
		t.Errorf("tiles reach %.2f dB, want 1.5 dB over the JPEG's second generation at %.2f", got, before)
	}
}

const minBackdropPSNR = 35.0
