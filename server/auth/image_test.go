package auth

import (
	"encoding/base64"
	"testing"
)

const (
	png64x64B64 = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAAnSURBVHhe7cEBDQAAAMKg909tDjegAAAAAAAAAAAAAAAAAAAAgHcDQEAAAVkKewwAAAAASUVORK5CYII="
	png32x32B64 = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAAaSURBVFhH7cEBAQAAAIIg/69uSEABAAAA7wYQIAABwWAsnwAAAABJRU5ErkJggg=="
)

func decode(t *testing.T, s string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return b
}

func TestSniffPNG(t *testing.T) {
	info, err := SniffImage(decode(t, png64x64B64))
	if err != nil {
		t.Fatalf("SniffImage: %v", err)
	}
	if info.Format != "png" || info.ContentType != "image/png" {
		t.Fatalf("unexpected info: %+v", info)
	}
	if info.Width != 64 || info.Height != 64 {
		t.Fatalf("dimensions = %dx%d, want 64x64", info.Width, info.Height)
	}
}

func TestSniffRejectsNonImage(t *testing.T) {
	for _, bad := range [][]byte{
		[]byte("<?php echo 1; ?>"),
		[]byte("MZ\x90\x00"), // Windows executable header
		{},
		[]byte{0x89, 0x50}, // truncated PNG signature
	} {
		if _, err := SniffImage(bad); err == nil {
			t.Fatalf("expected rejection for %q", bad)
		}
	}
}

func TestValidSkinDimensions(t *testing.T) {
	valid := [][2]int{{64, 64}, {64, 32}}
	for _, d := range valid {
		if !ValidSkinDimensions(d[0], d[1]) {
			t.Fatalf("%v should be valid", d)
		}
	}
	invalid := [][2]int{{32, 32}, {128, 128}, {64, 128}, {0, 0}}
	for _, d := range invalid {
		if ValidSkinDimensions(d[0], d[1]) {
			t.Fatalf("%v should be invalid", d)
		}
	}
}
