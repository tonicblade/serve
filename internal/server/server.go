package server

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	mime.AddExtensionType(".mod", "text/plain; charset=utf-8")
	mime.AddExtensionType(".go", "text/plain; charset=utf-8")
	mime.AddExtensionType(".md", "text/markdown; charset=utf-8")
	mime.AddExtensionType(".json", "application/json; charset=utf-8")
}

func Start(root string, host string, port int, depth int, quiet bool, open bool, once bool) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	handler := NewHandler(root, depth, true)
	if !quiet {
		fmt.Printf("Serving %s\n\n", root)
		fmt.Printf("  Local:   http://localhost:%d\n", port)
		for _, ip := range netAddrs(host) {
			fmt.Printf("  Network: http://%s:%d\n", ip, port)
		}
		fmt.Printf("  Depth:   %d\n", depth)
		fmt.Println()
		fmt.Println("Press Ctrl+C to stop")
	}
	if open {
		go func() {
			time.Sleep(250 * time.Millisecond)
			_ = openBrowser(fmt.Sprintf("http://localhost:%d", port))
		}()
	}
	return serve(listener, handler, once)
}

func serve(listener net.Listener, handler http.Handler, once bool) error {
	server := &http.Server{Handler: handler}
	if !once {
		return server.Serve(listener)
	}

	shutdownResult := make(chan error, 1)
	var requestReceived atomic.Bool
	var shutdownOnce sync.Once
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requestReceived.CompareAndSwap(false, true) {
			http.Error(w, "server is shutting down", http.StatusServiceUnavailable)
			return
		}

		handler.ServeHTTP(w, r)
		shutdownOnce.Do(func() {
			go func() {
				shutdownResult <- server.Shutdown(context.Background())
			}()
		})
	})

	err := server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return <-shutdownResult
	}
	return err
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", "", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}

func netAddrs(host string) []string {
	if host != "0.0.0.0" && host != "::" && host != "" {
		return nil
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, 4)
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, raw := range addrs {
			ipNet, ok := raw.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if _, ok := seen[ip.String()]; ok {
				continue
			}
			seen[ip.String()] = struct{}{}
			out = append(out, ip.String())
		}
	}
	if len(out) == 0 {
		out = append(out, "127.0.0.1")
	}
	return out
}

func NewHandler(root string, defaultDepth int, uploadEnabled bool) http.Handler {
	root = filepath.Clean(root)
	fs := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uploadEnabled && r.URL.Path == "/upload" {
			UploadHandler(root).ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/zip" {
			DownloadZipHandler(root).ServeHTTP(w, r)
			return
		}

		reqPath := r.URL.Path
		if reqPath == "" {
			reqPath = "/"
		}
		if reqPath != "/" && !strings.HasSuffix(reqPath, "/") {
			candidate := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(reqPath, "/")))
			info, err := os.Stat(candidate)
			if err == nil && info.IsDir() {
				http.Redirect(w, r, reqPath+"/?depth="+strconv.Itoa(defaultDepth), http.StatusMovedPermanently)
				return
			}
		}

		resolved := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(pathClean(reqPath), "/")))
		info, err := os.Stat(resolved)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if info.IsDir() {
			depth := defaultDepth
			if q := r.URL.Query().Get("depth"); q != "" {
				if v, err := strconv.Atoi(q); err == nil && v > 0 {
					depth = v
				}
			}
			RenderDirectoryPage(w, root, resolved, depth, reqPath, uploadEnabled)
			return
		}

		contentType, disposition := fileHeaderPolicy(info.Name())
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		if disposition != "" {
			w.Header().Set("Content-Disposition", disposition)
		}
		fs.ServeHTTP(w, r)
	})
}

func pathClean(p string) string {
	if p == "" || p == "/" {
		return "/"
	}
	return strings.TrimSuffix(path.Clean("/"+p), "/")
}

func isVideoFile(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range []string{".mp4", ".webm", ".ogg", ".mov", ".m4v", ".avi", ".mkv"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func fileHeaderPolicy(name string) (string, string) {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".mp4"):
		return "video/mp4", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".webm"):
		return "video/webm", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".ogg"):
		return "video/ogg", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".mov"):
		return "video/quicktime", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".m4v"):
		return "video/x-m4v", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".avi"):
		return "video/x-msvideo", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".mkv"):
		return "video/x-matroska", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".mod"):
		return "text/plain; charset=utf-8", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".go"):
		return "text/plain; charset=utf-8", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".json"):
		return "application/json; charset=utf-8", "inline; filename=\"" + name + "\""
	case strings.HasSuffix(lower, ".txt") || strings.HasSuffix(lower, ".md"):
		return "text/plain; charset=utf-8", "inline; filename=\"" + name + "\""
	default:
		return "", ""
	}
}

func DownloadZipHandler(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		targetDir := r.URL.Query().Get("dir")
		if targetDir == "" {
			targetDir = "/"
		}

		dirPath, ok := resolveUploadDirectory(root, targetDir)
		if !ok {
			http.Error(w, "invalid directory", http.StatusBadRequest)
			return
		}
		info, err := os.Stat(dirPath)
		if err != nil || !info.IsDir() {
			http.Error(w, "directory not found", http.StatusNotFound)
			return
		}

		archiveName := filepath.Base(dirPath)
		if archiveName == "." || archiveName == "" || archiveName == string(filepath.Separator) {
			archiveName = filepath.Base(root)
		}
		if archiveName == "" || archiveName == "." {
			archiveName = "download"
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+archiveName+".zip\"")

		if err := writeZipArchive(w, dirPath, archiveName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func writeZipArchive(w io.Writer, dirPath, archiveName string) error {
	zw := zip.NewWriter(w)
	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		entryName := filepath.ToSlash(filepath.Join(archiveName, rel))
		if info.IsDir() {
			_, err = zw.Create(entryName + "/")
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = entryName
		header.Method = zip.Deflate

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		return err
	}); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

func WaitForShutdown(d time.Duration) {
	<-time.After(d)
}
