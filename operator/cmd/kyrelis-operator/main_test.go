package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLocalSecurity(t *testing.T) {
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	a := &app{config: Config{RPC: "http://127.0.0.1:26657", REST: "http://127.0.0.1:1317"}, token: hex.EncodeToString(tokenBytes), configPath: t.TempDir() + "/config.json"}
	h := a.handler()
	req := httptest.NewRequest("GET", "http://evil.example/api/session", nil)
	req.Host = "evil.example"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("DNS rebinding accepted: %d", rec.Code)
	}
	req = httptest.NewRequest("GET", "http://"+listenAt+"/api/status", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("unauthenticated status: %d", rec.Code)
	}
	req = httptest.NewRequest("POST", "http://"+listenAt+"/api/config", strings.NewReader(`{"rpc":"http://example.org:26657","rest":"http://127.0.0.1:1317"}`))
	req.Header.Set("X-Kyrelis-Token", a.token)
	req.Header.Set("Origin", "http://"+listenAt)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("remote URL permitted: %d", rec.Code)
	}
}

func TestHostAPIAndInvalidBroadcast(t *testing.T) {
	a := &app{config: Config{RPC: "http://127.0.0.1:26657", REST: "http://127.0.0.1:1317"}, token: "local-secret", worker: &runner{}}
	for _, tc := range []struct {
		method, path, body string
		expected           int
	}{
		{"GET", "/api/host", "", 200},
		{"POST", "/api/broadcast", `{"tx_bytes":"not_base64"}`, 400},
		{"POST", "/api/broadcast", `{"tx_bytes":"bad","unexpected":true}`, 400},
	} {
		req := httptest.NewRequest(tc.method, "http://"+listenAt+tc.path, strings.NewReader(tc.body))
		req.Header.Set("X-Kyrelis-Token", a.token)
		req.Header.Set("Origin", "http://"+listenAt)
		rr := httptest.NewRecorder()
		a.handler().ServeHTTP(rr, req)
		if rr.Code != tc.expected {
			t.Errorf("%s returned %d, expected %d: %s", tc.path, rr.Code, tc.expected, rr.Body.String())
		}
	}
}
