package server

import (
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeOnceStopsAfterRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- serve(listener, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}), true)
	}()

	response, err := http.Get("http://" + listener.Addr().String())
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNoContent)
	}

	select {
	case err := <-serverDone:
		if err != nil {
			t.Fatalf("serve returned an error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop after serving a request")
	}
}
