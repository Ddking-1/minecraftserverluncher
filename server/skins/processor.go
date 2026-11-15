package skins

import (
	"bytes"
	"fmt"
	"image"
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
