package server

import (
	"html"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type TreeNode struct {
	Name     string
	Preview  []string
	Children []*TreeNode
}

type dirEntry struct {
	Name string
	Dir  bool
}

func BuildTree(root string, maxDepth int) []*TreeNode {
	if maxDepth <= 0 {
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	children := make([]*TreeNode, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		node := &TreeNode{Name: entry.Name()}
		node.Preview = PreviewEntries(filepath.Join(root, entry.Name()), 3)
		if maxDepth > 1 {
			node.Children = BuildTree(filepath.Join(root, entry.Name()), maxDepth-1)
		}
		children = append(children, node)
	}
	sort.Slice(children, func(i, j int) bool { return children[i].Name < children[j].Name })
	return children
}

func PreviewEntries(dirPath string, max int) []string {
	entries, err := os.ReadDir(dirPath)
	if err != nil || len(entries) == 0 {
		return nil
	}

	out := make([]string, 0, max)
	for _, entry := range entries {
		label := entry.Name()
		if entry.IsDir() {
			label += "/"
		}
		out = append(out, label)
		if len(out) >= max {
			break
		}
	}
	sort.Strings(out)
	return out
}

func parentFolderPath(requestPath string) string {
	clean := strings.TrimSuffix(requestPath, "/")
	if clean == "" || clean == "/" {
		return "/"
	}
	parent := path.Dir(clean)
	if parent == "." {
		return "/"
	}
	if !strings.HasPrefix(parent, "/") {
		parent = "/" + parent
	}
	return parent
}

func readDirEntries(dirPath string) []dirEntry {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil
	}

	result := make([]dirEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, dirEntry{Name: entry.Name(), Dir: entry.IsDir()})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Dir != result[j].Dir {
			return result[i].Dir
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func treeHTML(nodes []*TreeNode, basePath string, depth int) string {
	if len(nodes) == 0 {
		return "<div>No folders</div>"
	}

	var b strings.Builder
	b.WriteString("<ul class=\"tree\">")
	for _, node := range nodes {
		link := path.Join(basePath, node.Name) + "/?depth=" + strconv.Itoa(depth)
		b.WriteString("<li>")
		b.WriteString("<a href=\"")
		b.WriteString(html.EscapeString(link))
		b.WriteString("\">📁 ")
		b.WriteString(html.EscapeString(node.Name))
		b.WriteString("</a>")
		if len(node.Preview) > 0 {
			b.WriteString("<div class=\"tree-preview\">")
			for i, item := range node.Preview {
				if i > 0 {
					b.WriteString(" · ")
				}
				b.WriteString(html.EscapeString(item))
			}
			b.WriteString("</div>")
		}
		if len(node.Children) > 0 {
			b.WriteString(treeHTML(node.Children, path.Join(basePath, node.Name), depth))
		}
		b.WriteString("</li>")
	}
	b.WriteString("</ul>")
	return b.String()
}

func RenderDirectoryPage(w http.ResponseWriter, root, currentDir string, depth int, requestPath string, uploadEnabled bool) {
	entries := readDirEntries(currentDir)
	parts := strings.Split(strings.Trim(requestPath, "/"), "/")
	var breadcrumb []string
	if requestPath == "/" {
		breadcrumb = []string{"/"}
	} else {
		breadcrumb = []string{"/"}
		curr := ""
		for _, part := range parts {
			if part == "" {
				continue
			}
			curr = path.Join(curr, part)
			breadcrumb = append(breadcrumb, "/"+curr)
		}
	}

	var listing strings.Builder
	listing.WriteString("<div class=\"topbar\"><div class=\"title\"><h1>Index of ")
	listing.WriteString(html.EscapeString(root))
	listing.WriteString("</h1></div><div class=\"depth-control\"><label for=\"depth\">Depth</label><input id=\"depth\" type=\"number\" min=\"1\" max=\"8\" value=\"")
	listing.WriteString(strconv.Itoa(depth))
	listing.WriteString("\" /></div></div>")
	downloadTarget := "/zip?dir=" + url.QueryEscape(requestPath)
	listing.WriteString("<div class=\"toolbar\"><a class=\"download-button\" href=\"")
	listing.WriteString(html.EscapeString(downloadTarget))
	listing.WriteString("\"><i data-lucide=\"download\"></i> Download Zip</a></div>")
	listing.WriteString("<div class=\"breadcrumbs\">")
	for i, crumb := range breadcrumb {
		if i == 0 {
			listing.WriteString("<a href=\"/?depth=")
			listing.WriteString(strconv.Itoa(depth))
			listing.WriteString("\">/</a>")
			continue
		}
		listing.WriteString(" <span>/</span> ")
		listing.WriteString("<a href=\"")
		listing.WriteString(html.EscapeString(crumb + "?depth=" + strconv.Itoa(depth)))
		listing.WriteString("\">")
		listing.WriteString(html.EscapeString(path.Base(crumb)))
		listing.WriteString("</a>")
	}
	listing.WriteString("</div>")
	listing.WriteString("<div class=\"layout\"><div class=\"content\"><ul class=\"entries\">")
	if requestPath != "/" {
		parent := parentFolderPath(requestPath)
		listing.WriteString("<li class=\"entry parent\"><a href=\"")
		listing.WriteString(html.EscapeString(parent + "?depth=" + strconv.Itoa(depth)))
		listing.WriteString("\">⬆️ ..</a></li>")
	}
	for _, entry := range entries {
		itemPath := path.Join(requestPath, entry.Name)
		if entry.Dir {
			itemPath = strings.TrimSuffix(itemPath, "/") + "/"
		} else {
			itemPath = strings.TrimSuffix(itemPath, "/")
		}
		listing.WriteString("<li class=\"entry\"><a href=\"")
		listing.WriteString(html.EscapeString(itemPath + "?depth=" + strconv.Itoa(depth)))
		listing.WriteString("\">")
		if entry.Dir {
			listing.WriteString("📁 ")
		} else {
			listing.WriteString("📄 ")
		}
		listing.WriteString(html.EscapeString(entry.Name))
		listing.WriteString("</a></li>")
	}
	listing.WriteString("</ul>")
	if uploadEnabled {
		uploadTarget := "/upload?dir=" + url.QueryEscape(requestPath) + "&depth=" + strconv.Itoa(depth)
		listing.WriteString("<form class=\"upload\" action=\"")
		listing.WriteString(html.EscapeString(uploadTarget))
		listing.WriteString("\" method=\"post\" enctype=\"multipart/form-data\"><input type=\"hidden\" name=\"dir\" value=\"")
		listing.WriteString(html.EscapeString(requestPath))
		listing.WriteString("\" /><input type=\"file\" name=\"files\" multiple /><button type=\"submit\">Upload</button></form>")
	}
	listing.WriteString("</div><aside class=\"sidebar\"><h2>Folders</h2>")
	listing.WriteString(treeHTML(BuildTree(root, depth), "/", depth))
	listing.WriteString("</aside></div>")

	page := `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Index of ` + html.EscapeString(root) + `</title>
<style>
  body { margin: 0; font-family: Segoe UI, sans-serif; background: white; color: #111; line-height: 1.5; }
  a { color: #111; text-decoration: none; }
  a:hover { text-decoration: underline; }
  .wrap { max-width: 1400px; margin: 0 auto; padding: 24px; }
  .topbar { display: flex; justify-content: space-between; align-items: center; gap: 16px; margin-bottom: 12px; }
  h1 { margin: 0; font-size: 24px; font-weight: 600; }
  .depth-control { display: flex; align-items: center; gap: 8px; }
  input { width: 60px; padding: 4px 6px; border: 1px solid #bbb; }
  .breadcrumbs { margin: 12px 0 16px; }
  .toolbar { display: flex; justify-content: flex-end; margin: 0 0 12px; }
  .download-button { display: inline-flex; align-items: center; gap: 8px; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 8px; background: #f5f5f5; color: #111; font-weight: 600; }
  .download-button:hover { background: #eee; text-decoration: none; }
  .download-button svg { width: 16px; height: 16px; }
  .layout { display: grid; grid-template-columns: minmax(0, 1fr) 300px; gap: 20px; }
  .content, .sidebar { padding: 0; }
  .entries { list-style: none; margin: 0; padding: 0; }
  .entry { margin: 4px 0; }
  .upload { margin-top: 16px; display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
  .upload input[type="file"] { width: min(100%, 260px); max-width: 100%; }
  button { padding: 5px 10px; }
  .tree { list-style: none; margin: 0; padding-left: 14px; }
  .tree li { margin: 6px 0; }
  .tree-preview { color: #555; font-size: 12px; margin-left: 18px; }
  h2 { margin: 0 0 10px; font-size: 18px; font-weight: 600; }
  @media (max-width: 900px) { .layout { grid-template-columns: 1fr; } }
</style>
<script src="https://unpkg.com/lucide@latest"></script>
<script>
  window.addEventListener('DOMContentLoaded', function() {
    if (window.lucide) {
      lucide.createIcons();
    }
    const depthInput = document.getElementById('depth');
    if (depthInput) {
      depthInput.addEventListener('change', function() {
        const url = new URL(window.location.href);
        url.searchParams.set('depth', this.value);
        window.location.href = url.toString();
      });
    }
  });
</script>
</head>
<body>
<div class="wrap">` + listing.String() + `</div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(page))
}
