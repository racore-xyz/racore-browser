package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

var apiBase = "http://127.0.0.1:47831"

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
	index, err := os.Stat(filepath.Join(buildDir, "index.html"))
	if err != nil {
		fatal("build directory must contain index.html")
	}
	_ = index

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	fileIndex := make([]map[string]string, 0)
	totalSize := 0

	filepath.WalkDir(buildDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(buildDir, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fw, _ := zw.Create(rel)
		fw.Write(data)

		h := sha256.Sum256(data)
		fileIndex = append(fileIndex, map[string]string{
			"path": rel, "sha256": hex.EncodeToString(h[:]),
		})
		totalSize += len(data)
		return nil
	})
	zw.Close()

	indexData, _ := json.Marshal(fileIndex)
	contentRoot := hex.EncodeToString(sha256.New().Sum(indexData))[:64]

	if !isDomainOwned(domain) {
		if !isDomainAvailable(domain) {
			fatal("domain %s is claimed by another node", domain)
		}
		run("POST", "/v1/authority/domains", map[string]string{"domain": domain})
		fmt.Printf("claimed domain %s\n", domain)
	}

	uploadResp := doRequest("POST", "/v1/ipfs/add", "application/octet-stream", buf.Bytes())
	var uploadResult map[string]any
	json.Unmarshal(uploadResp, &uploadResult)

	cid, _ := uploadResult["cid"].(string)
	fmt.Printf("uploaded to IPFS: %s\n", cid)

	releaseBody := map[string]any{
		"version":     version,
		"cid":         cid,
		"contentRoot": contentRoot,
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
