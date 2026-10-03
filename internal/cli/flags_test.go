package cli

import "testing"

func TestParseArgs_Defaults(t *testing.T) {
	cfg, err := ParseArgs([]string{"."})
	if err != nil {
		t.Fatalf("ParseArgs returned unexpected error: %v", err)
	}
	if cfg.Root != "." {
		t.Fatalf("Root = %q, want %q", cfg.Root, ".")
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Depth != 2 {
		t.Fatalf("Depth = %d, want 2", cfg.Depth)
	}
}

func TestParseArgs_Flags(t *testing.T) {
	cfg, err := ParseArgs([]string{"./dist", "--port", "3000", "--host", "127.0.0.1", "--open", "--upload", "--depth", "4", "--expires", "30m"})
	if err != nil {
		t.Fatalf("ParseArgs returned unexpected error: %v", err)
	}
	if cfg.Root != "./dist" {
		t.Fatalf("Root = %q, want %q", cfg.Root, "./dist")
	}
	if cfg.Port != 3000 {
		t.Fatalf("Port = %d, want 3000", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "127.0.0.1")
	}
	if !cfg.Open {
		t.Fatal("Open should be true")
	}
	if !cfg.Upload {
		t.Fatal("Upload should be true")
	}
	if cfg.Depth != 4 {
		t.Fatalf("Depth = %d, want 4", cfg.Depth)
	}
	if cfg.Expires != "30m" {
		t.Fatalf("Expires = %q, want %q", cfg.Expires, "30m")
	}
}
