package server

import "testing"

func TestFileHeaderPolicy_UsesTextForGoModAndInlineForVideo(t *testing.T) {
	if got, _ := fileHeaderPolicy("go.mod"); got != "text/plain; charset=utf-8" {
		t.Fatalf("go.mod mime = %q, want text/plain; charset=utf-8", got)
	}
	if _, disposition := fileHeaderPolicy("clip.mp4"); disposition != "inline; filename=\"clip.mp4\"" {
		t.Fatalf("mp4 disposition = %q, want inline; filename=\"clip.mp4\"", disposition)
	}
}
