package skins

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"sync"
	"testing"
)

// encodeSkin builds a valid 64x64 skin PNG. face fills the head front face
// (8,8)-(16,16); overlay fills the hat layer (40,8)-(48,16).
func encodeSkin(t *testing.T, height int, face, overlay color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, height))
	for y := 8; y < 16; y++ {
		for x := 8; x < 16; x++ {
			img.SetRGBA(x, y, face)
			if height >= 16 {
				img.SetRGBA(x+32, y, overlay)
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return buf.Bytes()
}

func TestExtractAvatarHeadUsesFacePixels(t *testing.T) {
	skin := encodeSkin(t, 64, color.RGBA{R: 200, G: 20, B: 20, A: 255}, color.RGBA{A: 0})
	avatar, err := ExtractAvatarHead(skin)
	if err != nil {
		t.Fatalf("ExtractAvatarHead: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(avatar))
	if err != nil {
		t.Fatalf("decode avatar: %v", err)
	}
	if got := img.Bounds().Size(); got.X != 64 || got.Y != 64 {
		t.Fatalf("avatar size = %dx%d, want 64x64", got.X, got.Y)
	}
	r, g, b, a := img.At(32, 32).RGBA()
	if r>>8 != 200 || g>>8 != 20 || b>>8 != 20 || a>>8 != 255 {
		t.Fatalf("avatar centre = (%d,%d,%d,%d), want (200,20,20,255)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestExtractAvatarHeadCompositesOverlay(t *testing.T) {
	skin := encodeSkin(t, 64, color.RGBA{R: 200, A: 255}, color.RGBA{G: 200, A: 255})
	avatar, err := ExtractAvatarHead(skin)
	if err != nil {
		t.Fatalf("ExtractAvatarHead: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(avatar))
	if err != nil {
		t.Fatalf("decode avatar: %v", err)
	}
	r, g, b, a := img.At(10, 10).RGBA()
	if r>>8 != 0 || g>>8 != 200 || b>>8 != 0 || a>>8 != 255 {
		t.Fatalf("overlay pixel = (%d,%d,%d,%d), want (0,200,0,255)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestExtractAvatarHeadLegacy64x32(t *testing.T) {
	skin := encodeSkin(t, 32, color.RGBA{B: 255, A: 255}, color.RGBA{A: 0})
	if _, err := ExtractAvatarHead(skin); err != nil {
		t.Fatalf("64x32 skin should be accepted: %v", err)
	}
}

func TestExtractAvatarHeadRejectsInvalidDimensions(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractAvatarHead(buf.Bytes()); err == nil {
		t.Fatal("32x32 skin must be rejected")
	}
}

func TestSkinStorageLifecycle(t *testing.T) {
	store := NewMemoryStorage()
	if err := store.SaveSkin("abc-123", []byte("payload")); err != nil {
		t.Fatalf("SaveSkin: %v", err)
	}
	got, err := store.GetSkin("abc-123")
	if err != nil || string(got) != "payload" {
		t.Fatalf("GetSkin = %q, %v", got, err)
	}
	if err := store.DeleteSkin("abc-123"); err != nil {
		t.Fatalf("DeleteSkin: %v", err)
	}
	if _, err := store.GetSkin("abc-123"); err == nil {
		t.Fatal("deleted skin should not be found")
	}
}

func TestSkinStorageRejectsUnsafeIDs(t *testing.T) {
	store := NewMemoryStorage()
	for _, id := range []string{"", "../escape", "a/b", "id with space"} {
		if err := store.SaveSkin(id, []byte("x")); err == nil {
			t.Fatalf("SaveSkin(%q) should be rejected", id)
		}
	}
}

func TestSkinStorageConcurrentAccess(t *testing.T) {
	store := NewMemoryStorage()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := "skin-" + string(rune('a'+n%26))
			if err := store.SaveSkin(id, []byte("data")); err != nil {
				t.Errorf("SaveSkin: %v", err)
			}
			_, _ = store.GetSkin(id)
		}(i)
	}
	wg.Wait()
}

func TestDiskStoragePersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	first, err := NewDiskStorage(dir)
	if err != nil {
		t.Fatalf("NewDiskStorage: %v", err)
	}
	if err := first.SaveSkin("persisted-id", []byte("skin-bytes")); err != nil {
		t.Fatalf("SaveSkin: %v", err)
	}

	second, err := NewDiskStorage(dir)
	if err != nil {
		t.Fatalf("reopen NewDiskStorage: %v", err)
	}
	got, err := second.GetSkin("persisted-id")
	if err != nil || string(got) != "skin-bytes" {
		t.Fatalf("GetSkin after reopen = %q, %v", got, err)
	}
	if err := second.DeleteSkin("persisted-id"); err != nil {
		t.Fatalf("DeleteSkin: %v", err)
	}

	third, err := NewDiskStorage(dir)
	if err != nil {
		t.Fatalf("reopen after delete: %v", err)
	}
	if _, err := third.GetSkin("persisted-id"); err == nil {
		t.Fatal("deleted skin must not survive a reopen")
	}
}
