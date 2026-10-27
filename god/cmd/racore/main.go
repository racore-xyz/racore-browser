package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var apiBase = "http://127.0.0.1:47831"

type buildFile struct {
	name string
	data []byte
}

var (
	quotedRootPathPattern = regexp.MustCompile(`["'\x60](/[^"'\x60\s]+)["'\x60]`)
	cssURLPattern         = regexp.MustCompile(`url\(\s*["']?(/[^"')\s]+)`)
)

// rootRelativeWarnings reports build references that start at the gateway root.
// A release is served from /ipfs/<cid>/, so "/assets/app.js" misses the release
// even though the file is inside it.
func rootRelativeWarnings(files []buildFile) []string {
	known := make(map[string]bool, len(files))
	for _, file := range files {
		known[strings.ToLower(file.name)] = true
	}
	seen := make(map[string]bool)
	warnings := make([]string, 0)
	for _, file := range files {
		switch strings.ToLower(path.Ext(file.name)) {
		case ".html", ".htm", ".css", ".js", ".mjs":
		default:
			continue
		}
		matches := quotedRootPathPattern.FindAllSubmatch(file.data, -1)
		matches = append(matches, cssURLPattern.FindAllSubmatch(file.data, -1)...)
		for _, match := range matches {
			raw := string(match[1])
			if strings.HasPrefix(raw, "//") {
				continue
			}
			target := raw
			if cut := strings.IndexAny(target, "?#"); cut >= 0 {
				target = target[:cut]
			}
			target = strings.TrimPrefix(target, "/")
			if target == "" || !known[strings.ToLower(target)] || seen[target] {
				continue
			}
			seen[target] = true
			warnings = append(warnings, fmt.Sprintf(
				"%s references %q from the site root; under /ipfs/<cid>/ it resolves to %s instead of %s. Rebuild with a relative base (for example Vite base: \"./\").",
				file.name, raw, raw, target,
			))
		}
	}
	sort.Strings(warnings)
	return warnings
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Fprintln(os.Stderr, "Usage: racore <command> [args...]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  status                   Show daemon health")
		fmt.Fprintln(os.Stderr, "  domains                  List claimed domains")
		fmt.Fprintln(os.Stderr, "  claim <domain>           Claim a domain")
		fmt.Fprintln(os.Stderr, "  publish --domain <d> --build <dir> --version <v>")
		fmt.Fprintln(os.Stderr, "                           Publish a web build")
		fmt.Fprintln(os.Stderr, "  releases <domain>        List domain releases")
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "status":
		run("GET", "/health", nil)
	case "domains":
		run("GET", "/v1/authority/domains", nil)
	case "claim":
		if len(args) < 1 {
			fatal("usage: racore claim <domain>")
		}
		domain := args[0]
		avail := getJSON(fmt.Sprintf("/v1/authority/domains/%s/available", domain))
		if m, ok := avail.(map[string]any); ok {
			if a, _ := m["available"].(bool); !a {
				fatal("domain %s is not available (scope: %v)", domain, m["scope"])
			}
		}
		body := map[string]string{"domain": domain}
		run("POST", "/v1/authority/domains", body)
	case "publish":
		domain := ""
		buildDir := ""
		version := ""

		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--domain":
				if i+1 < len(args) {
					domain = args[i+1]
					i++
				}
			case "--build":
				if i+1 < len(args) {
					buildDir = args[i+1]
					i++
				}
			case "--version":
				if i+1 < len(args) {
					version = args[i+1]
					i++
				}
			}
		}

		if domain == "" || buildDir == "" || version == "" {
			fatal("usage: racore publish --domain <d> --build <dir> --version <v>")
		}

		publish(domain, buildDir, version)
	case "releases":
		if len(args) < 1 {
			fatal("usage: racore releases <domain>")
		}
		run("GET", fmt.Sprintf("/v1/authority/domains/%s/releases", args[0]), nil)
	default:
		fatalf("unknown command: %s", cmd)
	}
}

func publish(domain, buildDir, version string) {
	if _, err := os.Stat(filepath.Join(buildDir, "index.html")); err != nil {
		fatal("build directory must contain index.html")
	}

	files := make([]buildFile, 0)
	fileIndex := make([]map[string]string, 0)
	totalSize := 0

	walkErr := filepath.WalkDir(buildDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		rel, err := filepath.Rel(buildDir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		files = append(files, buildFile{name: filepath.ToSlash(rel), data: data})
		fileIndex = append(fileIndex, map[string]string{
			"path": filepath.ToSlash(rel), "sha256": hex.EncodeToString(sum[:]),
		})
		totalSize += len(data)
		return nil
	})
	if walkErr != nil {
		fatal("read build directory: %v", walkErr)
	}
	if len(files) == 0 {
		fatal("build directory is empty")
	}
	for _, warning := range rootRelativeWarnings(files) {
		fmt.Fprintf(os.Stderr, "warning: %s\n", warning)
	}

	indexData, _ := json.Marshal(fileIndex)
	contentRoot := sha256.Sum256(indexData)

	var upload bytes.Buffer
	writer := multipart.NewWriter(&upload)
	paths := make([]string, 0, len(files))
	for _, file := range files {
		name := strings.ReplaceAll(file.name, "\\", "/")
		part, err := writer.CreateFormFile("file", path.Base(name))
		if err != nil {
			fatal("build upload: %v", err)
		}
		if _, err := part.Write(file.data); err != nil {
			fatal("build upload: %v", err)
		}
		paths = append(paths, name)
	}
	pathIndex, err := json.Marshal(paths)
	if err != nil {
		fatal("build upload: %v", err)
	}
	if err := writer.WriteField("paths", string(pathIndex)); err != nil {
		fatal("build upload: %v", err)
	}
	writer.Close()

	if !isDomainOwned(domain) {
		if !isDomainAvailable(domain) {
			fatal("domain %s is claimed by another node", domain)
		}
		run("POST", "/v1/authority/domains", map[string]string{"domain": domain})
		fmt.Printf("claimed domain %s\n", domain)
	}

	uploadResp := doRequest("POST", "/v1/ipfs/add", writer.FormDataContentType(), upload.Bytes())
	var uploadResult map[string]any
	if err := json.Unmarshal(uploadResp, &uploadResult); err != nil {
		fatal("invalid upload response: %v", err)
	}
	cid, _ := uploadResult["cid"].(string)
	if cid == "" {
		fatal("upload did not return a content identifier: %s", string(uploadResp))
	}
	fmt.Printf("uploaded %d files to IPFS: %s\n", len(files), cid)

	releaseBody := map[string]any{
		"version":     version,
		"cid":         cid,
		"contentRoot": hex.EncodeToString(contentRoot[:]),
		"entrypoint":  "index.html",
		"files":       len(fileIndex),
		"size":        totalSize,
	}
	run("POST", fmt.Sprintf("/v1/authority/domains/%s/releases", domain), releaseBody)
	fmt.Printf("published release %s for %s\n", version, domain)
}

func isDomainOwned(domain string) bool {
	resp := getJSON("/v1/authority/domains")
	domains, ok := resp.([]any)
	if !ok {
		return false
	}
	for _, d := range domains {
		if m, ok := d.(map[string]any); ok {
			if m["domain"] == domain {
				return true
			}
		}
	}
	return false
}

func isDomainAvailable(domain string) bool {
	resp := getJSON(fmt.Sprintf("/v1/authority/domains/%s/available", domain))
	if m, ok := resp.(map[string]any); ok {
		a, _ := m["available"].(bool)
		return a
	}
	return false
}

func run(method, path string, body any) {
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
	}
	resp := doRequest(method, path, "application/json", data)
	var out bytes.Buffer
	json.Indent(&out, resp, "", "  ")
	out.WriteString("\n")
	os.Stdout.Write(out.Bytes())
}

func doRequest(method, path, contentType string, body []byte) []byte {
	url := apiBase + path
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		fatal("request: %v", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fatal("http: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fatal("read: %v", err)
	}

	if resp.StatusCode >= 400 {
		fatal("error %d: %s", resp.StatusCode, string(data))
	}

	return data
}

func getJSON(path string) any {
	data := doRequest("GET", path, "", nil)
	var result any
	json.Unmarshal(data, &result)
	return result
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fatal(format, args...)
}
