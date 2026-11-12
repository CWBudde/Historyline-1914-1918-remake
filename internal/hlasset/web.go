package hlasset

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"

	"github.com/gen2brain/webp"
)

// Qualities of the web edition's compressed images. LiteWebP is the first
// (and offline) art of hex-masked terrain, replaced online by the lossless
// original; FieldWebP is final art, a crater field's bake has no lossless
// twin.
const (
	liteJPEG = 85
	// LiteWebP is the quality of lite art with an alpha channel.
	LiteWebP = 80
	// FieldWebP is the quality of baked art that ships without a lossless twin.
	FieldWebP = 85
)

// LosslessWebP is the encoder setting of the web edition's full art: lossless, at a
// mid effort that costs about a third of PNG's bytes less. Exact keeps the colour of
// fully transparent pixels, which the compositor can read through its filters.
var LosslessWebP = webp.Options{Lossless: true, Quality: 75, Method: 4, Exact: true}

// EncodeLossless stores im as a lossless WebP that decodes to the same pixels
// as the PNG of im, including the colour under alpha 0.
func EncodeLossless(im image.Image) ([]byte, error) {
	n, ok := im.(*image.NRGBA)
	if !ok || n.Rect.Min != (image.Point{}) {
		n = image.NewNRGBA(image.Rect(0, 0, im.Bounds().Dx(), im.Bounds().Dy()))
		draw.Draw(n, n.Rect, im, im.Bounds().Min, draw.Src)
	}
	var buf bytes.Buffer
	if err := webp.Encode(&buf, n, LosslessWebP); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// bleedPasses is how far the colour of a visible pixel is spread into the
// transparent pixels around it: far enough to cover the blocks the encoder
// transforms together.
const bleedPasses = 8

// EncodeLite compresses an image for the web edition: a JPEG when it is
// opaque, else a lossy WebP at LiteWebP. The caller keeps the lossless
// original when this is larger, and keeps materials lossless.
func EncodeLite(im image.Image) (ext string, data []byte, err error) {
	return EncodeAt(im, LiteWebP)
}

// EncodeField is EncodeLite at FieldWebP, for art that has no lossless twin.
func EncodeField(im image.Image) (ext string, data []byte, err error) {
	return EncodeAt(im, FieldWebP)
}

// EncodeAt is EncodeLite with the WebP quality of an image with transparency.
// The colour of its transparent pixels is first spread from the visible ones
// so the encoder does not blend the hex mask's black into the border; a
// hard-edged mask would otherwise show a dark fringe.
func EncodeAt(im image.Image, webpQuality int) (ext string, data []byte, err error) {
	n := image.NewNRGBA(im.Bounds())
	draw.Draw(n, n.Rect, im, im.Bounds().Min, draw.Src)
	var buf bytes.Buffer
	if opaque(n) {
		err = jpeg.Encode(&buf, n, &jpeg.Options{Quality: liteJPEG})
		return ".jpg", buf.Bytes(), err
	}
	err = webp.Encode(&buf, Bleed(n, bleedPasses), webp.Options{Quality: webpQuality, Method: 4})
	return ".webp", buf.Bytes(), err
}

// Bleed returns a copy of n (a zero-based image, as image.NewNRGBA makes) whose fully transparent pixels take the average
// colour of their visible neighbours, repeated for up to passes steps. The
// visible pixels and all alpha values are unchanged.
func Bleed(n *image.NRGBA, passes int) *image.NRGBA {
	out := image.NewNRGBA(n.Rect)
	copy(out.Pix, n.Pix)
	w, h := n.Rect.Dx(), n.Rect.Dy()
	known := make([]bool, w*h)
	for i := range known {
		known[i] = out.Pix[4*i+3] != 0
	}
	next := make([]bool, len(known))
	for range passes {
		copy(next, known)
		changed := false
		for y := range h {
			for x := range w {
				if known[y*w+x] {
					continue
				}
				var r, g, b, c int
				add := func(xx, yy int) {
					if xx < 0 || yy < 0 || xx >= w || yy >= h || !known[yy*w+xx] {
						return
					}
					i := 4 * (yy*w + xx)
					r, g, b, c = r+int(out.Pix[i]), g+int(out.Pix[i+1]), b+int(out.Pix[i+2]), c+1
				}
				add(x-1, y)
				add(x+1, y)
				add(x, y-1)
				add(x, y+1)
				if c == 0 {
					continue
				}
				i := 4 * (y*w + x)
				out.Pix[i], out.Pix[i+1], out.Pix[i+2] = uint8(r/c), uint8(g/c), uint8(b/c)
				next[y*w+x] = true
				changed = true
			}
		}
		known, next = next, known
		if !changed {
			break
		}
	}
	return out
}

func opaque(n *image.NRGBA) bool {
	for y := range n.Rect.Dy() {
		row := n.Pix[y*n.Stride : y*n.Stride+4*n.Rect.Dx()]
		for i := 3; i < len(row); i += 4 {
			if row[i] != 255 {
				return false
			}
		}
	}
	return true
}
