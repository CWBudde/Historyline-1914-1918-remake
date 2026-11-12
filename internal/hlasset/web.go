package hlasset

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"

	"github.com/gen2brain/webp"
)

// Qualities of the web edition's compressed images.
const liteJPEG, liteWebP = 85, 80

// EncodeLite compresses an image for the web edition: a JPEG when it is
// opaque, else a lossy WebP. The caller keeps the lossless original when this
// is larger, and keeps materials lossless.
func EncodeLite(im image.Image) (ext string, data []byte, err error) {
	n := image.NewNRGBA(im.Bounds())
	draw.Draw(n, n.Rect, im, im.Bounds().Min, draw.Src)
	var buf bytes.Buffer
	if opaque(n) {
		err = jpeg.Encode(&buf, n, &jpeg.Options{Quality: liteJPEG})
		return ".jpg", buf.Bytes(), err
	}
	err = webp.Encode(&buf, n, webp.Options{Quality: liteWebP, Method: 4})
	return ".webp", buf.Bytes(), err
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
