package server

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func resolveUploadDirectory(root, requestPath string) (string, bool) {
	if requestPath == "" || requestPath == "/" {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return "", false
		}
		return absRoot, true
	}
	clean := path.Clean("/" + requestPath)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	candidate := filepath.Join(absRoot, filepath.FromSlash(strings.TrimPrefix(clean, "/")))
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(absRoot, absCandidate)
	if err != nil {
		return "", false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", false
	}
	info, err := os.Stat(absCandidate)
	if err == nil && info.IsDir() {
		return absCandidate, true
	}
	if err == nil {
		return absCandidate, true
	}
	return "", false
}

func UploadHandler(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		targetDir := r.FormValue("dir")
		if targetDir == "" {
			targetDir = r.URL.Query().Get("dir")
		}
		if targetDir == "" {
			targetDir = "/"
		}

		dirPath, ok := resolveUploadDirectory(root, targetDir)
		if !ok {
			http.Error(w, "invalid upload directory", http.StatusBadRequest)
			return
		}
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		files := r.MultipartForm.File["files"]
		for _, fh := range files {
			cleanName := filepath.Clean(fh.Filename)
			if cleanName == "." || cleanName == ".." || filepath.IsAbs(cleanName) || strings.Contains(cleanName, "..") {
				http.Error(w, "invalid filename", http.StatusBadRequest)
				return
			}

			in, err := fh.Open()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer in.Close()

			outPath := filepath.Join(dirPath, filepath.Base(fh.Filename))
			out, err := os.Create(outPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer out.Close()
			if _, err := out.ReadFrom(in); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		depth := r.FormValue("depth")
		if depth == "" {
			depth = r.URL.Query().Get("depth")
		}
		if depth == "" {
			depth = "2"
		}
		redirectTarget := targetDir
		if redirectTarget == "" || redirectTarget == "/" {
			redirectTarget = "/"
		}
		if _, err := strconv.Atoi(depth); err != nil {
			depth = "2"
		}
		http.Redirect(w, r, redirectTarget+"?depth="+depth, http.StatusSeeOther)
	})
}
