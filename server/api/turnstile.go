package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"net/url"
	"time"
)

// verifyTurnstile checks a Cloudflare Turnstile token against the siteverify
// endpoint. Returns nil on success (or when no secret is configured, i.e. dev
// mode). On failure returns an error describing the problem.
func (s *Server) verifyTurnstile(token, remoteIP string) error {
	if s == nil || s.cfg == nil || s.cfg.TurnstileSecret == "" {
		return nil // dev mode: CAPTCHA not enforced
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("missing turnstile token")
	}
	form := url.Values{}
	form.Set("secret", s.cfg.TurnstileSecret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", form)
	if err != nil {
		return fmt.Errorf("turnstile request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return fmt.Errorf("turnstile read failed: %w", err)
	}
	var result struct {
		Success bool     `json:"success"`
		Errors  []string `json:"error-codes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("turnstile decode failed: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("turnstile verification failed: %v", result.Errors)
	}
	return nil
}
