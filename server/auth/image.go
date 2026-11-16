// Image format sniffing and Minecraft skin dimension validation.
//
// Security: the declared Content-Type is never trusted. The actual bytes are
// inspected (magic numbers) so a script or executable cannot be smuggled in
// with an image content type. Only PNG and JPEG are accepted.
package auth

import (
	"encoding/binary"
	"errors"
)

// ErrInvalidImage is returned when the bytes are not a recognized image.
var ErrInvalidImage = errors.New("invalid image data")

// ImageInfo describes a recognized image.
type ImageInfo struct {
	Format     string // "png" or "jpeg"
	ContentType string // canonical MIME type
	Width      int
	Height     int
}

// SniffImage inspects raw bytes and returns the detected format and, when
// available, the pixel dimensions. It rejects anything that is not PNG/JPEG.
func SniffImage(data []byte) (*ImageInfo, error) {
	if isPNG(data) {
		w, h, ok := pngDimensions(data)
		if !ok {
			return nil, ErrInvalidImage
		}
		return &ImageInfo{Format: "png", ContentType: "image/png", Width: w, Height: h}, nil
	}
	if isJPEG(data) {
		w, h, ok := jpegDimensions(data)
		if !ok {
			return nil, ErrInvalidImage
		}
		return &ImageInfo{Format: "jpeg", ContentType: "image/jpeg", Width: w, Height: h}, nil
	}
	return nil, ErrInvalidImage
}

func isPNG(b []byte) bool {
	return len(b) >= 8 &&
		b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4E && b[3] == 0x47 &&
		b[4] == 0x0D && b[5] == 0x0A && b[6] == 0x1A && b[7] == 0x0A
}

func isJPEG(b []byte) bool {
	return len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF
}

// pngDimensions reads width/height from the IHDR chunk (bytes 16..24).
func pngDimensions(b []byte) (int, int, bool) {
	if len(b) < 24 {
		return 0, 0, false
	}
	// The first chunk must be IHDR for a valid PNG.
	if string(b[12:16]) != "IHDR" {
		return 0, 0, false
	}
	w := int(binary.BigEndian.Uint32(b[16:20]))
	h := int(binary.BigEndian.Uint32(b[20:24]))
	if w <= 0 || h <= 0 {
		return 0, 0, false
	}
	// Defense in depth: reject absurdly large declared dimensions even though
	// ValidSkinDimensions would reject them later, so a malicious header
	// cannot drive downstream allocations.
	if w > 10000 || h > 10000 {
		return 0, 0, false
	}
	return w, h, true
}

// jpegDimensions scans markers until a SOFn frame header is found.
func jpegDimensions(b []byte) (int, int, bool) {
	i := 2
	for i+9 < len(b) {
		if b[i] != 0xFF {
			i++
			continue
		}
		marker := b[i+1]
		// Skip padding fill bytes.
		if marker == 0xFF {
			i++
			continue
		}
		// Standalone markers without a length payload.
		if marker == 0xD8 || marker == 0xD9 || (marker >= 0xD0 && marker <= 0xD7) {
			i += 2
			continue
		}
		if i+4 > len(b) {
			return 0, 0, false
		}
		segLen := int(binary.BigEndian.Uint16(b[i+2 : i+4]))
		if segLen < 2 {
			return 0, 0, false
		}
		// SOF0..SOF15, excluding DHT (0xC4), JPG (0xC8), DAC (0xCC).
		if marker >= 0xC0 && marker <= 0xCF &&
			marker != 0xC4 && marker != 0xC8 && marker != 0xCC {
			if i+9 > len(b) {
				return 0, 0, false
			}
			h := int(binary.BigEndian.Uint16(b[i+5 : i+7]))
			w := int(binary.BigEndian.Uint16(b[i+7 : i+9]))
			if w <= 0 || h <= 0 {
				return 0, 0, false
			}
			// Defense in depth: reject absurdly large declared dimensions.
			if w > 10000 || h > 10000 {
				return 0, 0, false
			}
			return w, h, true
		}
		i += 2 + segLen
	}
	return 0, 0, false
}

// ValidSkinDimensions reports whether the dimensions match a Minecraft skin:
// 64x64 (modern, with overlays) or 64x32 (legacy).
func ValidSkinDimensions(w, h int) bool {
	return (w == 64 && h == 64) || (w == 64 && h == 32)
}
