package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUploadDirectory_UsesCurrentFolderAndRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "alpha", "beta")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	got, ok := resolveUploadDirectory(root, "/alpha/beta")
	if !ok {
		t.Fatal("resolveUploadDirectory should allow nested target directory")
	}
	if got != nested {
		t.Fatalf("resolved directory = %q, want %q", got, nested)
	}

	if _, ok := resolveUploadDirectory(root, "/alpha/../etc"); ok {
		t.Fatal("resolveUploadDirectory should reject traversal outside the root")
	}

	rootDir, ok := resolveUploadDirectory(root, "/")
	if !ok {
		t.Fatal("resolveUploadDirectory should accept the root directory")
	}
	if rootDir != root {
		t.Fatalf("root directory = %q, want %q", rootDir, root)
	}
}
