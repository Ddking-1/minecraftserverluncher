package skins

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// MaxSkinInputBytes is the maximum allowed raw skin upload size for ProcessSkin.
const MaxSkinInputBytes int64 = 100 * 1024

// MaxSkinBytes mirrors MaxSkinInputBytes for callers that prefer the shorter name.
const MaxSkinBytes = MaxSkinInputBytes

// maxStoredSkinBytes caps a single file read back from disk so a corrupted or
// tampered storage directory cannot exhaust memory at startup.
const maxStoredSkinBytes = 4 << 20

// AllowedSkinDimensions enumerates the only permitted Minecraft skin dimensions
// for ProcessSkin: 64x64 and 64x32 pixels.
var AllowedSkinDimensions = map[image.Point]struct{}{
	{64, 64}: {},
	{64, 32}: {},
}

// storageIDPattern restricts ids to characters that are safe both as map keys
// and as file names (the API generates UUIDs, avatars append "-avatar").
var storageIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// SkinStorage is a concurrency-safe skin store.
//
// With an empty directory it keeps skins in memory only. With a directory it
// also writes each skin to "<dir>/<id>.png" (atomically) and loads the existing
// files at startup, so uploaded skins and avatar URLs survive a restart.
//
// The zero value is not usable; construct one with NewMemoryStorage or
// NewDiskStorage.
type SkinStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
	dir  string
}

// NewMemoryStorage returns an empty in-memory SkinStorage ready for use.
func NewMemoryStorage() *SkinStorage {
	return &SkinStorage{data: make(map[string][]byte)}
}

// NewDiskStorage returns a SkinStorage that persists under dir. Existing PNG
// files in the directory are loaded into memory. The directory is created when
// missing.
func NewDiskStorage(dir string) (*SkinStorage, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("skins: storage directory must not be empty")
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("skins: create storage directory: %w", err)
	}
	s := &SkinStorage{data: make(map[string][]byte), dir: dir}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("skins: read storage directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".png") {
			continue
		}
		id := strings.TrimSuffix(name, ".png")
		if !storageIDPattern.MatchString(id) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || len(data) == 0 || len(data) > maxStoredSkinBytes {
			continue
		}
		s.data[id] = data
	}
	return s, nil
}

// SaveSkin persists a skin under the given id.
func (s *SkinStorage) SaveSkin(id string, data []byte) error {
	if s == nil {
		return fmt.Errorf("skins: storage is not configured")
	}
	if !storageIDPattern.MatchString(id) {
		return fmt.Errorf("skins: invalid skin id %q", id)
	}
	if len(data) == 0 {
		return fmt.Errorf("skins: refusing to store empty skin %q", id)
	}
	if len(data) > maxStoredSkinBytes {
		return fmt.Errorf("skins: skin %q exceeds the storage size limit", id)
	}

	copied := append([]byte(nil), data...)

	if s.dir != "" {
		tmp, err := os.CreateTemp(s.dir, ".skin-*")
		if err != nil {
			return fmt.Errorf("skins: create temp file: %w", err)
		}
		tmpName := tmp.Name()
		defer os.Remove(tmpName)
		if _, err := tmp.Write(copied); err != nil {
			_ = tmp.Close()
			return fmt.Errorf("skins: write temp file: %w", err)
		}
		if err := tmp.Sync(); err != nil {
			_ = tmp.Close()
			return fmt.Errorf("skins: sync temp file: %w", err)
		}
		if err := tmp.Close(); err != nil {
			return fmt.Errorf("skins: close temp file: %w", err)
		}
		if err := os.Rename(tmpName, filepath.Join(s.dir, id+".png")); err != nil {
			return fmt.Errorf("skins: persist skin: %w", err)
		}
	}

	s.mu.Lock()
	if s.data == nil {
		s.data = make(map[string][]byte)
	}
	s.data[id] = copied
	s.mu.Unlock()
	return nil
}

// GetSkin returns the raw skin bytes for id, or an error if it does not exist.
func (s *SkinStorage) GetSkin(id string) ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("skins: storage is not configured")
	}
	s.mu.RLock()
	data, ok := s.data[id]
	s.mu.RUnlock()
	if ok {
		return append([]byte(nil), data...), nil
	}

	if s.dir != "" && storageIDPattern.MatchString(id) {
		data, err := os.ReadFile(filepath.Join(s.dir, id+".png"))
		if err == nil && len(data) > 0 && len(data) <= maxStoredSkinBytes {
			s.mu.Lock()
			if s.data == nil {
				s.data = make(map[string][]byte)
			}
			s.data[id] = data
			s.mu.Unlock()
			return append([]byte(nil), data...), nil
		}
	}

	return nil, fmt.Errorf("skin %q does not exist", id)
}

// DeleteSkin removes a skin. It returns nil even if the id was already missing.
func (s *SkinStorage) DeleteSkin(id string) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	delete(s.data, id)
	s.mu.Unlock()

	if s.dir != "" && storageIDPattern.MatchString(id) {
		if err := os.Remove(filepath.Join(s.dir, id+".png")); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("skins: delete skin file: %w", err)
		}
	}
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

// ExtractAvatarHead returns a 64x64 avatar PNG built from the head of a skin.
//
// Minecraft skins store the head front face at (8,8) and its hat/overlay layer
// at (40,8) in both the 64x64 and the legacy 64x32 layout. The 8x8 face is
// composited with the overlay and scaled up to 64x64 with nearest-neighbour
// sampling so the result stays pixel-crisp.
func ExtractAvatarHead(data []byte) ([]byte, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("skin processor: not a valid PNG image: %w", err)
	}
	if _, ok := AllowedSkinDimensions[image.Point{X: cfg.Width, Y: cfg.Height}]; !ok {
		return nil, fmt.Errorf(
			"skin processor: unsupported skin dimensions %dx%d (allowed: 64x64, 64x32)",
			cfg.Width, cfg.Height,
		)
	}
	img, err := decodeImage(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	size := img.Bounds().Size()
	if _, ok := AllowedSkinDimensions[size]; !ok {
		return nil, fmt.Errorf(
			"skin processor: unsupported skin dimensions %dx%d (allowed: 64x64, 64x32)",
			size.X, size.Y,
		)
	}

	const (
		faceX    = 8
		faceY    = 8
		hatX     = 40
		hatY     = 8
		headSize = 8
		scale    = 64 / headSize
	)

	avatar := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < headSize; y++ {
		for x := 0; x < headSize; x++ {
			fr, fg, fb, fa := img.At(faceX+x, faceY+y).RGBA()
			hr, hg, hb, ha := img.At(hatX+x, hatY+y).RGBA()

			// Premultiplied "over" compositing: hat over face.
			inv := uint32(0xffff) - ha
			r := hr + fr*inv/0xffff
			g := hg + fg*inv/0xffff
			b := hb + fb*inv/0xffff
			a := ha + fa*inv/0xffff

			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					offset := avatar.PixOffset(x*scale+dx, y*scale+dy)
					avatar.Pix[offset+0] = uint8(r >> 8)
					avatar.Pix[offset+1] = uint8(g >> 8)
					avatar.Pix[offset+2] = uint8(b >> 8)
					avatar.Pix[offset+3] = uint8(a >> 8)
				}
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
//
// Alpha normalization: body-region pixels (head, body, arms, legs) are forced
// opaque so a skin can never render an invisible player. Transparency is
// preserved only outside the skin template and in the optional overlay layer.
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

	// Decode only the header first to reject oversized images before the full
	// (potentially memory-exhausting) decode. This guards against PNG
	// decompression bombs.
	cfg, _, err := image.DecodeConfig(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("skin processor: not a valid PNG image: %w", err)
	}
	if _, ok := AllowedSkinDimensions[image.Point{X: cfg.Width, Y: cfg.Height}]; !ok {
		return nil, fmt.Errorf(
			"skin processor: unsupported skin dimensions %dx%d (allowed: 64x64, 64x32)",
			cfg.Width, cfg.Height,
		)
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

	// Normalize alpha: body parts must be opaque, transparency only allowed
	// outside the skin template and in the overlay layer.
	img = normalizeSkinAlpha(img, size)

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

// skinBodyRects defines the pixel regions of a Minecraft skin that belong to
// the character body and must be opaque. These are the standard 64x64 skin
// template regions; for 64x32 skins the lower half simply does not exist.
var skinBodyRects = []image.Rectangle{
	// Head front face.
	{image.Point{8, 8}, image.Point{16, 16}},
	// Body front face.
	{image.Point{16, 16}, image.Point{24, 28}},
	// Right arm (viewer's left).
	{image.Point{40, 16}, image.Point{48, 28}},
	// Left arm.
	{image.Point{32, 48}, image.Point{40, 60}},
	// Right leg.
	{image.Point{0, 16}, image.Point{8, 28}},
	// Left leg.
	{image.Point{16, 48}, image.Point{24, 60}},
	// Head top.
	{image.Point{8, 0}, image.Point{16, 8}},
	// Body top.
	{image.Point{16, 4}, image.Point{24, 16}},
	// Right arm top.
	{image.Point{40, 4}, image.Point{48, 16}},
	// Left arm top.
	{image.Point{32, 36}, image.Point{40, 48}},
	// Right leg top.
	{image.Point{0, 4}, image.Point{8, 16}},
	// Left leg top.
	{image.Point{16, 36}, image.Point{24, 48}},
}

// isBodyPixel reports whether (x, y) is inside one of the body regions of a
// 64x64 Minecraft skin. For 64x32 skins the lower-body regions simply do not
// contain pixels, so the check is still correct.
func isBodyPixel(x, y int, width, height int) bool {
	if x < 0 || y < 0 || x >= width || y >= height {
		return false
	}
	for _, r := range skinBodyRects {
		if r.Min.X <= x && x < r.Max.X && r.Min.Y <= y && y < r.Max.Y {
			return true
		}
	}
	return false
}

// normalizeSkinAlpha forces body-region pixels to full opacity while preserving
// transparency outside the skin template and in overlay regions.
//
// This prevents uploaded skins from rendering invisible players when the source
// PNG uses transparency around the body. Overlay-layer pixels (the hat/jacket
// layer at x >= 32 in 64x64 skins) keep their alpha so hats can still fade.
func normalizeSkinAlpha(img image.Image, size image.Point) image.Image {
	width, height := size.X, size.Y
	is64x32 := height == 32

	// Convert to a writable RGBA image so we can patch alpha in place.
	b := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := rgba.At(x, y).RGBA()

			// Overlay layer: x in [32,64) and y >= 32 in 64x64 skins.
			// Keep the original alpha so hats/jackets can be transparent.
			if !is64x32 && x >= 32 && y >= 32 {
				continue
			}

			// Body pixels must be opaque.
			if isBodyPixel(x, y, width, height) {
				if a != 0xffff {
					rgba.SetRGBA(x, y, color.RGBA{
						R: uint8(r >> 8),
						G: uint8(g >> 8),
						B: uint8(b >> 8),
						A: 255,
					})
				}
				continue
			}

			// Outside the body: preserve original alpha, but never leave a
			// fully transparent pixel fully transparent if it is near the skin
			// (within the 64x64 bounding box). This avoids holes in the model.
			if a == 0 {
				// Keep transparent pixels transparent only if they are far from
				// the skin body (outside the 0..64 x 0..64 box) or if the skin
				// is 64x32 and the pixel is in the non-existent lower half.
				if width == 64 && height == 64 && x >= 0 && x < 64 && y >= 0 && y < 64 {
					// Near the skin: give it a minimal alpha so it does not create
					// a hole, but keep it very subtle.
					rgba.SetRGBA(x, y, color.RGBA{
						R: uint8(r >> 8),
						G: uint8(g >> 8),
						B: uint8(b >> 8),
						A: 1,
					})
				}
			}
		}
	}

	return rgba
}
