package server

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadZipHandler_DownloadsCurrentDirectory(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "alpha", "beta")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "hello.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/zip?dir=/alpha/beta", nil)
	w := httptest.NewRecorder()
	NewHandler(root, 2, true).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "application/zip") {
		t.Fatalf("Content-Type = %q, want zip", got)
	}
	if got := w.Header().Get("Content-Disposition"); !strings.Contains(got, "beta.zip") {
		t.Fatalf("Content-Disposition = %q, want beta.zip", got)
	}

	zr, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	if len(zr.File) != 1 {
		t.Fatalf("zip entries = %d, want 1", len(zr.File))
	}
	if got := zr.File[0].Name; got != "beta/hello.txt" {
		t.Fatalf("zip entry = %q, want %q", got, "beta/hello.txt")
	}
	f, err := zr.File[0].Open()
	if err != nil {
		t.Fatalf("Open zip entry: %v", err)
	}
	defer f.Close()
	contents, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("ReadAll zip entry: %v", err)
	}
	if string(contents) != "hello" {
		t.Fatalf("zip contents = %q, want %q", string(contents), "hello")
	}
}
