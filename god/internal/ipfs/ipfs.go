package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Bridge struct {
	apiURL     string
	gatewayURL string
	client     *http.Client
}

func New(apiURL, gatewayURL string) *Bridge {
	return &Bridge{
		apiURL:     apiURL,
		gatewayURL: gatewayURL,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (b *Bridge) Status(ctx context.Context) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", b.apiURL+"/api/v0/id", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]any{"online": false}, nil
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	result["online"] = true
	return result, nil
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
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	cid, _ := result["Hash"].(string)
	result["cid"] = cid
	result["gateway"] = b.gatewayURL + "/ipfs/" + cid
	result["name"] = name
	return result, nil
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
