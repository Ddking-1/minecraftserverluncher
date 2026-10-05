package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// postJSON sends a JSON POST and returns (status, decoded body).
func postJSON(base, path string, body any, authToken string) (int, map[string]any) {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, base+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return do(req)
}

// postRaw sends a raw-body POST (used for the skin upload) and returns (status, decoded body).
func postRaw(base, path string, body []byte, contentType, authToken string) (int, map[string]any) {
	req, _ := http.NewRequest(http.MethodPost, base+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return do(req)
}

// getAdminJSON sends a GET with the X-Admin-Token header.
func getAdminJSON(base, path, adminToken string) (int, map[string]any) {
	req, _ := http.NewRequest(http.MethodGet, base+path, nil)
	if adminToken != "" {
		req.Header.Set("X-Admin-Token", adminToken)
	}
	return do(req)
}

func getJSON(base, path string, authToken string) (int, map[string]any) {
	req, _ := http.NewRequest(http.MethodGet, base+path, nil)
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return do(req)
}

func do(req *http.Request) (int, map[string]any) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil
	}
	defer resp.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body
}

func must(b bool, msg string) {
	if !b {
		fmt.Println("FAIL:", msg)
		os.Exit(1)
	}
}

func main() {
	baseURL := "http://localhost:8089"
	adminToken := "e2e-admin"
	// A minimal valid 1x1 PNG.
	png := decodeB64("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAHQAHx8vJPeAAAAABJRU5ErkJggg==")

	// 1. Register -> 201, tokens.
	code, body := postJSON(baseURL, "/api/v1/auth/register", map[string]any{
		"username": "e2euser", "password": "supersecret", "device_id": "dev1", "device_name": "pc",
	}, "")
	must(code == http.StatusCreated, "register != 201")
	access, _ := body["access_token"].(string)
	must(access != "", "register no access token")

	// 2. Login -> 200, new access token.
	code, body = postJSON(baseURL, "/api/v1/auth/login", map[string]any{
		"username": "e2euser", "password": "supersecret", "device_id": "dev1",
	}, "")
	must(code == http.StatusOK, "login != 200")
	access, _ = body["access_token"].(string)
	must(access != "", "login no access token")

	// 3. Me -> 200, username correct.
	code, body = getJSON(baseURL, "/api/v1/account/me", access)
	must(code == http.StatusOK && body["username"] == "e2euser", "me != e2euser")

	// 4. Forgot -> 200 (no enumeration).
	code, _ = postJSON(baseURL, "/api/v1/auth/forgot", map[string]any{"identifier": "e2euser"}, "")
	must(code == http.StatusOK, "forgot != 200")

	// 5. Async skin upload -> 202 + task_id.
	code, body = postRaw(baseURL, "/api/v1/skins/upload/default", png, "image/png", access)
	must(code == http.StatusAccepted, "skin upload != 202")
	taskID, _ := body["task_id"].(string)
	must(taskID != "", "skin upload no task_id")

	// 6. Poll task until done (max 20 x 100ms).
	done := false
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		code, body = getJSON(baseURL, "/api/v1/tasks/"+taskID, access)
		if code == http.StatusOK && body["status"] == "done" {
			done = true
			break
		}
	}
	must(done, "skin task never reached done")

	// 7. Admin list (requires X-Admin-Token).
	code, body = getAdminJSON(baseURL, "/api/admin/users", adminToken)
	must(code == http.StatusOK, "admin list != 200")
	total, _ := body["total"].(float64)
	must(total >= 1, "admin total < 1")

	// 8. Logout-all -> 200.
	code, _ = postJSON(baseURL, "/api/v1/auth/logout-all", nil, access)
	must(code == http.StatusOK, "logout-all != 200")

	fmt.Println("PASS")
}

func decodeB64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
