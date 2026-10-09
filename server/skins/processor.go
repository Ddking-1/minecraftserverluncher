package skins

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
)

// MaxSkinInputBytes is the maximum allowed raw skin upload size for ProcessSkin.
const MaxSkinInputBytes int64 = 100 * 1024

// AllowedSkinDimensions enumerates the only permitted Minecraft skin dimensions
// for ProcessSkin: 64x64 and 64x32 pixels.
var AllowedSkinDimensions = map[image.Point]struct{}{
	{64, 64}: {},
	{64, 32}: {},
}

// ProcessSkin validates and re-encodes a Minecraft skin PNG.
//
// The input must be a valid PNG not exceeding MaxSkinInputBytes in raw size and
// decoded dimensions limited to 64x64 or 64x32. The image is fully decoded and
// re-encoded into a new PNG to strip EXIF, metadata, and any other extra chunks
// from the original file. The returned bytes are the clean re-encoded PNG.
func ProcessSkin(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes > MaxSkinInputBytes {
		maxBytes = MaxSkinInputBytes
	}
	if maxBytes <= 0 {
		return nil, fmt.Errorf("skin processor: invalid max size %d", maxBytes)
	}

	// Read the input with a hard cap so we never allocate beyond the limit.
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(io.LimitReader(r, maxBytes+1)); err != nil {
		return nil, fmt.Errorf("skin processor: read error: %w", err)
	}
	if buf.Len() > int(maxBytes) {
		return nil, fmt.Errorf(
			"skin processor: input too large (%d bytes), max %d bytes",
			buf.Len(), maxBytes,
		)
	}

	if !pngSignature(buf.Bytes()) {
		return nil, fmt.Errorf("skin processor: not a valid PNG (bad magic bytes)")
	}

	img, err := decodeImage(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("skin processor: %w", err)
	}

	size := img.Bounds().Size()
	if _, ok := AllowedSkinDimensions[size]; !ok {
		return nil, fmt.Errorf(
			"skin processor: unsupported skin dimensions %dx%d (allowed: 64x64, 64x32)",
			size.X, size.Y,
		)
	}

	// Re-encode to strip metadata, EXIF, comments, and extra chunks.
	out := new(bytes.Buffer)
	if err := png.Encode(out, img); err != nil {
		return nil, fmt.Errorf("skin processor: encoding error: %w", err)
	}

	return out.Bytes(), nil
}

// pngSignature reports whether b starts with the 8-byte PNG file signature.
func pngSignature(b []byte) bool {
	if len(b) < 8 {
		return false
	}
	return b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4E && b[3] == 0x47 &&
		b[4] == 0x0D && b[5] == 0x0A && b[6] == 0x1A && b[7] == 0x0A
}

// decodeImage decodes a PNG image using the standard library decoder. The
// image/png import is required so that image.Decode can recognize PNG data.
func decodeImage(r io.Reader) (image.Image, error) {
	img, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("skin processor: not a valid PNG image: %w", err)
	}
	return img, nil
}
