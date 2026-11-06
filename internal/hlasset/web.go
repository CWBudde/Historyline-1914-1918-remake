package hlasset

import "image"

// EncodeLite applies the web pack's JPEG/WebP quality to an image. The caller
// keeps the lossless original when this is larger, and keeps materials lossless.
func EncodeLite(im image.Image) (ext string, data []byte, err error) {
	return liteEncode(im)
}
