package skins

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"sync"
)

// MaxSkinInputBytes is the maximum allowed raw skin upload size for ProcessSkin.
const MaxSkinInputBytes int64 = 100 * 1024

// MaxSkinBytes mirrors MaxSkinInputBytes for callers that prefer the shorter name.
const MaxSkinBytes = MaxSkinInputBytes

// AllowedSkinDimensions enumerates the only permitted Minecraft skin dimensions
// for ProcessSkin: 64x64 and 64x32 pixels.
var AllowedSkinDimensions = map[image.Point]struct{}{
	{64, 64}: {},
	{64, 32}: {},
}

// SkinStorage is the in-memory skin store used by the API handlers.
type SkinStorage map[string][]byte

// NewMemoryStorage returns an empty SkinStorage ready for use.
func NewMemoryStorage() SkinStorage {
	return make(SkinStorage)
}

// SaveSkin persists a skin under the given id.
func (s SkinStorage) SaveSkin(id string, data []byte) error {
	s[id] = append([]byte(nil), data...)
	return nil
}

// GetSkin returns the raw skin bytes for id, or an error if it does not exist.
func (s SkinStorage) GetSkin(id string) ([]byte, error) {
	data, ok := s[id]
	if !ok {
		return nil, fmt.Errorf("skin %q does not exist", id)
	}
	return append([]byte(nil), data...), nil
}

// DeleteSkin removes a skin. It returns nil even if the id was already missing.
func (s SkinStorage) DeleteSkin(id string) error {
	delete(s, id)
	return nil
}

// ThreadSafeSkinHistory tracks recently uploaded skins and supports rotation,
// deletion, and quick-switch active_skin selection.
type ThreadSafeSkinHistory struct {
	mu        sync.Mutex
	history   []string
	active    string
	byAccount map[string][]string
}

// NewThreadSafeSkinHistory returns a ready-to-use thread-safe skin history.
func NewThreadSafeSkinHistory() *ThreadSafeSkinHistory {
	return &ThreadSafeSkinHistory{
		byAccount: make(map[string][]string),
	}
}

// AddSkin records a newly uploaded skin and returns the id of any evicted skin.
func (h *ThreadSafeSkinHistory) AddSkin(id string, recent bool) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if recent {
		h.history = append(h.history, id)
		if len(h.history) > 10 {
			evicted := h.history[0]
			h.history = h.history[1:]
			return evicted, nil
		}
	}
	return "", nil
}

// RemoveSkin removes a skin from history and clears it as active if needed.
func (h *ThreadSafeSkinHistory) RemoveSkin(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.active == id {
		h.active = ""
	}
	for i, existing := range h.history {
		if existing == id {
			copy(h.history[i:], h.history[i+1:])
			h.history = h.history[:len(h.history)-1]
			break
		}
	}
}

// SetActive marks id as the active skin. It returns an error if id is unknown.
func (h *ThreadSafeSkinHistory) SetActive(id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, existing := range h.history {
		if existing == id {
			h.active = id
			return nil
		}
	}
	return fmt.Errorf("skin %q is not in history", id)
}

// ExtractAvatarHead returns a 64x64 head avatar PNG from a full skin PNG.
func ExtractAvatarHead(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("skin processor: not a valid PNG image: %w", err)
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	avatarW := 64
	avatarH := 64
	if avatarW > width {
		avatarW = width
	}
	if avatarH > height {
		avatarH = height
	}
	avatar := image.NewRGBA(image.Rect(0, 0, avatarW, avatarH))
	for y := 0; y < avatarH; y++ {
		for x := 0; x < avatarW; x++ {
			srcX := x * width / avatarW
			srcY := y * height / avatarH
						srcY0 := srcY + bounds.Min.Y
			srcX0 := srcX + bounds.Min.X
		srcIdx := srcY0*bounds.Stride + srcX0*4
		dstIdx := y*avatar.(*image.RGBA).Stride + x*4
			if rgba, ok := img.(*image.RGBA); ok {
				avatar.Pix[dstIdx] = rgba.Pix[srcIdx]
				avatar.Pix[dstIdx+1] = rgba.Pix[srcIdx+1]
				avatar.Pix[dstIdx+2] = rgba.Pix[srcIdx+2]
				avatar.Pix[dstIdx+3] = 0xff
			} else {
				r, g, b, _ := img.At(srcX0, srcY0).RGBA()
				avatar.Pix[dstIdx] = uint8(r >> 8)
				avatar.Pix[dstIdx+1] = uint8(g >> 8)
				avatar.Pix[dstIdx+2] = uint8(b >> 8)
				avatar.Pix[dstIdx+3] = 0xff
			}
		}
	}
	out := new(bytes.Buffer)
	if err := png.Encode(out, avatar); err != nil {
		return nil, fmt.Errorf("skin processor: encoding error: %w", err)
	}
	return out.Bytes(), nil
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
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("skin processor: not a valid PNG image: %w", err)
	}
	return img, nil
}
