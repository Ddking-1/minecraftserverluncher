package api

import (
	"testing"

	"minecraftserverluncher/server/auth"
)

// FuzzSniffPNG feeds arbitrary byte slices to auth.SniffImage (png branch)
// and verifies that the function never panics, even on malformed/corrupt
// PNG-like headers. This is the skin image-sniffing counterpart to the
// auth-token fuzzing in auth_fuzz_test.go.
func FuzzSniffPNG(f *testing.F) {
	// Seeds cover empty input, truncated headers, valid magic but non-IHDR,
	// valid IHDR with non-standard dimensions, and a real 1x1 minimal PNG.
	f.Add([]byte{})
	f.Add([]byte{0x89, 0x50, 0x4E, 0x47})
	f.Add([]byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x40, // width = 64
		0x00, 0x00, 0x00, 0x40, // height = 64
	})
	f.Add([]byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x40,
		0x00, 0x00, 0x00, 0x20, // height = 32 (acceptable skin)
	})
	f.Add([]byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0xFF, 0xFF, 0xFF, 0xFF, // non-sensical dimensions
		0xFF, 0xFF, 0xFF, 0xFF,
	})
	f.Add([]byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, // width = 1
		0x00, 0x00, 0x00, 0x01, // height = 1
	})

	f.Fuzz(func(t *testing.T, data []byte) {
		// SniffImage must never panic on arbitrary input. It should return
		// either a valid image info or ErrInvalidImage (or a read error).
		info, err := auth.SniffImage(data)

		if err == nil && info != nil {
			// When parsing "succeeds", the returned dimensions still need to
			// obey the same invariants the skin handler enforces downstream.
			if info.Width <= 0 || info.Height <= 0 {
				t.Fatalf("SniffImage returned non-positive dimensions: %dx%d", info.Width, info.Height)
			}
			if info.Format != "png" && info.Format != "jpeg" {
				t.Fatalf("SniffImage returned unexpected format: %q", info.Format)
			}
		}

		// Additional skin-specific validation: if the image is a plausible PNG
		// with plausible dimensions, make sure ValidSkinDimensions behaves
		// deterministically and never panics on these values.
		if info != nil && info.Format == "png" {
			ok := auth.ValidSkinDimensions(info.Width, info.Height)
			_ = ok
		}
	})
}
