package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSPAHandler(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("app"), 0o644); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(spaHandler(dir))
	defer server.Close()

	for _, route := range []string{"/", "/settings"} {
		resp, err := http.Get(server.URL + route)
		if err != nil {
			t.Fatal(err)
		}
		body := make([]byte, 3)
		_, _ = resp.Body.Read(body)
		resp.Body.Close()
		if !strings.Contains(string(body), "app") {
			t.Fatalf("%s did not return SPA", route)
		}
	}
}
