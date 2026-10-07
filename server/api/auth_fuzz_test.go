package api

import (
	"testing"

	"minecraftserverluncher/server/sessions"
)

func FuzzParseSessionToken(f *testing.F) {
	f.Add("")
	f.Add("not-base64!!")
	f.Add("a===")
	f.Add("AAAA")
	f.Add("AA==")
	f.Add("validtoken")
	f.Fuzz(func(t *testing.T, tok string) {
		ts := sessions.NewService(nil, 0, 0)
		if _, err := ts.ParsePair(tok); err == nil {
			// We only care about malformed input paths; valid base64 tokens
			// are expected to parse here, so we do not fail on them.
		}
	})
}
