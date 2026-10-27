package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Bridge struct {
	apiURL     string
	gatewayURL string
	client     *http.Client
}

type File struct {
	Name string
	Data []byte
}

func New(apiURL, gatewayURL string) *Bridge {
	return &Bridge{
		apiURL:     apiURL,
		gatewayURL: gatewayURL,
		client:     &http.Client{Timeout: 2 * time.Minute},
	}
}

func (b *Bridge) Status(ctx context.Context) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", b.apiURL+"/api/v0/id", nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return map[string]any{"online": false}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]any{"online": false}, nil
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return map[string]any{"online": false}, nil
	}
	return map[string]any{
		"online":       true,
		"peerId":       raw["ID"],
		"agentVersion": raw["AgentVersion"],
		"addresses":    raw["Addresses"],
		"protocols":    raw["Protocols"],
	}, nil
}

func (b *Bridge) AddBytes(ctx context.Context, name string, content []byte) (map[string]any, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", name)
	if err != nil {
		return nil, err
	}
	fw.Write(content)
	w.Close()

	url := fmt.Sprintf("%s/api/v0/add?pin=true&cid-version=1", b.apiURL)
	last, err := b.postAdd(ctx, url, w.FormDataContentType(), &buf)
	if err != nil {
		return nil, err
	}

	cid, _ := last["Hash"].(string)
	if cid == "" {
		return nil, fmt.Errorf("kubo returned no content identifier")
	}
	last["cid"] = cid
	last["gateway"] = b.gatewayURL + "/ipfs/" + cid
	last["name"] = name
	return last, nil
}

// AddFiles imports a directory tree into Kubo and returns the root CID. Kubo's
// multipart add API flattens filenames to their base name, so nested paths are
// assembled in MFS first; that keeps assets/... servable from /ipfs/<root>/.
func (b *Bridge) AddFiles(ctx context.Context, files []File) (map[string]any, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to add")
	}

	root := fmt.Sprintf("/racore-publish-%d", time.Now().UnixNano())
	for _, file := range files {
		name := strings.ReplaceAll(file.Name, "\\", "/")
		name = strings.TrimPrefix(path.Clean("/"+name), "/")
		if name == "" || name == "." || strings.HasPrefix(name, "../") {
			return nil, fmt.Errorf("invalid file path %q", file.Name)
		}
		if err := b.writeMFS(ctx, root+"/"+name, file.Data); err != nil {
			b.removeMFS(ctx, root)
			return nil, err
		}
	}

	cid, err := b.statMFS(ctx, root)
	if err != nil {
		b.removeMFS(ctx, root)
		return nil, err
	}
	if _, err := b.Pin(ctx, cid); err != nil {
		b.removeMFS(ctx, root)
		return nil, fmt.Errorf("pin published directory: %w", err)
	}
	b.removeMFS(ctx, root)

	return map[string]any{
		"cid":     cid,
		"gateway": b.gatewayURL + "/ipfs/" + cid + "/",
		"files":   len(files),
	}, nil
}

func (b *Bridge) writeMFS(ctx context.Context, mfsPath string, data []byte) error {
	query := url.Values{
		"arg":      {mfsPath},
		"create":   {"true"},
		"parents":  {"true"},
		"truncate": {"true"},
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("data", "blob")
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		b.apiURL+"/api/v0/files/write?"+query.Encode(),
		&body,
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("kubo files write returned %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (b *Bridge) statMFS(ctx context.Context, mfsPath string) (string, error) {
	query := url.Values{"arg": {mfsPath}, "hash": {"true"}}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		b.apiURL+"/api/v0/files/stat?"+query.Encode(),
		nil,
	)
	if err != nil {
		return "", err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("kubo files stat returned %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}
	var payload struct {
		Hash string `json:"Hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.Hash == "" {
		return "", fmt.Errorf("kubo returned no directory content identifier")
	}
	return payload.Hash, nil
}

func (b *Bridge) removeMFS(ctx context.Context, mfsPath string) {
	query := url.Values{"arg": {mfsPath}, "recursive": {"true"}}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		b.apiURL+"/api/v0/files/rm?"+query.Encode(),
		nil,
	)
	if err != nil {
		return
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (b *Bridge) postAdd(
	ctx context.Context,
	url string,
	contentType string,
	body io.Reader,
) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("kubo add returned %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}

	decoder := json.NewDecoder(resp.Body)
	var last map[string]any
	for {
		var entry map[string]any
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		last = entry
	}
	if last == nil {
		return nil, fmt.Errorf("kubo add returned an empty response")
	}
	return last, nil
}

func (b *Bridge) Cat(ctx context.Context, cid string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v0/cat?arg=%s", b.apiURL, cid)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (b *Bridge) Pin(ctx context.Context, cid string) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/v0/pin/add?arg=%s", b.apiURL, cid)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (b *Bridge) PublishIPNS(ctx context.Context, cid string) (map[string]any, error) {
	url := fmt.Sprintf("%s/api/v0/name/publish?arg=%s&allow-offline=true", b.apiURL, cid)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}
