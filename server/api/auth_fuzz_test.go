package api

import (
	"testing"

	"minecraftserverluncher/server/auth"
)

func FuzzHashToken(f *testing.F) {
	f.Add("")
	f.Add("not-base64!!")
	f.Add("a===")
	f.Add("AAAA")
	f.Add("AA==")
	f.Add("validtoken")
	f.Fuzz(func(t *testing.T, tok string) {
		// HashToken should never panic on any input.
		result := auth.HashToken(tok)
		if result == "" {
			t.Fatal("HashToken returned empty string")
		}
	})
}
