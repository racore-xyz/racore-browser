package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/racore/god/pkg/api"
)

type Provider struct {
	ID          string
	Name        string
	Kind        string
	BaseURL     string
	DefaultModel string
	Auth        string
	Free        bool
	Local       bool
}

var Catalog = []Provider{
	{ID: "openai", Name: "OpenAI", Kind: "responses", BaseURL: "https://api.openai.com/v1", DefaultModel: "gpt-5.6-terra", Auth: "bearer"},
	{ID: "anthropic", Name: "Anthropic", Kind: "anthropic", BaseURL: "https://api.anthropic.com/v1", DefaultModel: "claude-sonnet-4-5", Auth: "x-api-key"},
	{ID: "gemini", Name: "Google Gemini", Kind: "gemini", BaseURL: "https://generativelanguage.googleapis.com/v1beta", DefaultModel: "gemini-2.5-flash", Auth: "query"},
	{ID: "openrouter", Name: "OpenRouter", Kind: "openai", BaseURL: "https://openrouter.ai/api/v1", DefaultModel: "openrouter/auto", Auth: "bearer"},
	{ID: "kimi", Name: "Kimi / Moonshot", Kind: "openai", BaseURL: "https://api.moonshot.ai/v1", DefaultModel: "kimi-k2.5", Auth: "bearer"},
	{ID: "ollama", Name: "Ollama", Kind: "ollama", BaseURL: "http://127.0.0.1:11434", DefaultModel: "qwen3:8b", Auth: "none", Free: true, Local: true},
	{ID: "opencode", Name: "OpenCode", Kind: "cli", DefaultModel: "configured", Auth: "local", Free: true, Local: true},
	{ID: "claude-code", Name: "Claude Code", Kind: "cli", DefaultModel: "sonnet", Auth: "local", Local: true},
	{ID: "kimi-code", Name: "Kimi Code", Kind: "cli", DefaultModel: "default", Auth: "local", Free: true, Local: true},
}

type KeyFunc func(providerID string) (string, error)

type Gateway struct {
	httpClient *http.Client
	getKey     KeyFunc
}

func NewGateway(getKey KeyFunc) *Gateway {
	return &Gateway{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		getKey:     getKey,
	}
}

func (g *Gateway) Catalog() []api.Provider {
	out := make([]api.Provider, len(Catalog))
	for i, p := range Catalog {
		key, _ := g.getKey(p.ID)
		masked := ""
		if key != "" {
			if len(key) < 8 {
				masked = "***"
			} else {
				masked = key[:3] + "..." + key[len(key)-4:]
			}
		}
		out[i] = api.Provider{
			ID:           p.ID,
			Name:         p.Name,
			Kind:         p.Kind,
			BaseURL:      p.BaseURL,
			DefaultModel: p.DefaultModel,
			Auth:         p.Auth,
			Free:         p.Free,
			Local:        p.Local,
			Connected:    key != "" || p.Auth == "none" || p.Auth == "local",
			MaskedKey:    masked,
		}
	}
	return out
}

func (g *Gateway) health(ctx context.Context, p Provider) (map[string]any, error) {
	switch p.Kind {
	case "ollama":
		return g.ollamaHealth(ctx)
	case "cli":
		return g.cliHealth(p)
	default:
		key, err := g.getKey(p.ID)
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error()}, nil
		}
		if key == "" {
			return map[string]any{"ok": false, "error": "not connected"}, nil
		}
		return map[string]any{"ok": true}, nil
	}
}

func (g *Gateway) Health(ctx context.Context, id string) (map[string]any, error) {
	for _, p := range Catalog {
		if p.ID == id {
			return g.health(ctx, p)
		}
	}
	return nil, fmt.Errorf("provider not found: %s", id)
}

func (g *Gateway) ollamaHealth(ctx context.Context) (map[string]any, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:11434/api/tags", nil)
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return map[string]any{"ok": false, "error": err.Error()}, nil
	}
	defer resp.Body.Close()
	return map[string]any{"ok": resp.StatusCode == 200}, nil
}

func (g *Gateway) cliHealth(p Provider) (map[string]any, error) {
	cmd := p.ID
	if p.ID == "claude-code" {
		cmd = "claude"
	}
	if _, err := exec.LookPath(cmd); err != nil {
		return map[string]any{"ok": false, "error": fmt.Sprintf("%s not found", cmd)}, nil
	}
	return map[string]any{"ok": true}, nil
}

func (g *Gateway) Chat(ctx context.Context, providerID, model string, messages []map[string]string, system string) (*api.ChatResponse, error) {
	var p Provider
	found := false
	for _, pr := range Catalog {
		if pr.ID == providerID {
			p = pr
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown provider: %s", providerID)
	}

	if model == "" {
		model = p.DefaultModel
	}

	switch p.Kind {
	case "responses":
		return g.openaiResponses(ctx, p, model, messages, system)
	case "openai":
		return g.openaiCompat(ctx, p, model, messages, system)
	case "anthropic":
		return g.anthropicChat(ctx, p, model, messages, system)
	case "gemini":
		return g.geminiChat(ctx, p, model, messages, system)
	case "ollama":
		return g.ollamaChat(ctx, p, model, messages, system)
	case "cli":
		return g.cliChat(ctx, p, model, messages, system)
	default:
		return nil, fmt.Errorf("unsupported provider kind: %s", p.Kind)
	}
}

func (g *Gateway) openaiResponses(ctx context.Context, p Provider, model string, messages []map[string]string, system string) (*api.ChatResponse, error) {
	key, err := g.getKey(p.ID)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"model": model,
		"input": buildMessages(messages, system),
	}
	return g.doOpenAI(ctx, p.BaseURL+"/responses", key, body, p.ID, model)
}

func (g *Gateway) openaiCompat(ctx context.Context, p Provider, model string, messages []map[string]string, system string) (*api.ChatResponse, error) {
	key, err := g.getKey(p.ID)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"model":    model,
		"messages": buildMessages(messages, system),
		"stream":   false,
	}
	return g.doOpenAI(ctx, p.BaseURL+"/chat/completions", key, body, p.ID, model)
}

func (g *Gateway) doOpenAI(ctx context.Context, url, key string, body any, providerID, model string) (*api.ChatResponse, error) {
	data, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(raw, &result)

	text := extractText(result)
	usage, _ := result["usage"]
	rawID, _ := result["id"].(string)

	return &api.ChatResponse{
		Text: text, Usage: usage, Provider: providerID,
		Model: model, RawID: rawID,
	}, nil
}

func (g *Gateway) anthropicChat(ctx context.Context, p Provider, model string, messages []map[string]string, system string) (*api.ChatResponse, error) {
	key, err := g.getKey(p.ID)
	if err != nil {
		return nil, err
	}

	anthropicMsgs := make([]map[string]any, 0)
	sysMsg := ""
	for _, m := range messages {
		role := m["role"]
		if role == "system" {
			sysMsg = m["content"]
			continue
		}
		anthropicMsgs = append(anthropicMsgs, map[string]any{
			"role":    role,
			"content": m["content"],
		})
	}
	if system != "" {
		sysMsg = system
	}

	body := map[string]any{
		"model":      model,
		"messages":   anthropicMsgs,
		"max_tokens": 4096,
	}
	if sysMsg != "" {
		body["system"] = sysMsg
	}

	data, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/messages", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(raw, &result)

	text := ""
	if content, ok := result["content"].([]any); ok && len(content) > 0 {
		if first, ok := content[0].(map[string]any); ok {
			text, _ = first["text"].(string)
		}
	}
	usage, _ := result["usage"]

	return &api.ChatResponse{Text: text, Usage: usage, Provider: p.ID, Model: model}, nil
}

func (g *Gateway) geminiChat(ctx context.Context, p Provider, model string, messages []map[string]string, system string) (*api.ChatResponse, error) {
	key, err := g.getKey(p.ID)
	if err != nil {
		return nil, err
	}

	geminiContents := make([]map[string]any, 0)
	for _, m := range messages {
		role := "user"
		if m["role"] == "assistant" || m["role"] == "model" {
			role = "model"
		}
		geminiContents = append(geminiContents, map[string]any{
			"role": role,
			"parts": []map[string]any{
				{"text": m["content"]},
			},
		})
	}

	body := map[string]any{"contents": geminiContents}
	url := fmt.Sprintf("%s/models/%s:generateContent", p.BaseURL, model)

	data, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", key)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(raw, &result)

	text := ""
	if candidates, ok := result["candidates"].([]any); ok && len(candidates) > 0 {
		if first, ok := candidates[0].(map[string]any); ok {
			if content, ok := first["content"].(map[string]any); ok {
				if parts, ok := content["parts"].([]any); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]any); ok {
						text, _ = part["text"].(string)
					}
				}
			}
		}
	}

	return &api.ChatResponse{Text: text, Provider: p.ID, Model: model}, nil
}

func (g *Gateway) ollamaChat(ctx context.Context, p Provider, model string, messages []map[string]string, system string) (*api.ChatResponse, error) {
	ollamaMsgs := buildMessages(messages, system)
	body := map[string]any{
		"model":    model,
		"messages": ollamaMsgs,
		"stream":   false,
	}

	data, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/api/chat", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(raw, &result)

	text := ""
	if msg, ok := result["message"].(map[string]any); ok {
		text, _ = msg["content"].(string)
	}

	return &api.ChatResponse{Text: text, Provider: p.ID, Model: model}, nil
}

func (g *Gateway) cliChat(ctx context.Context, p Provider, model string, messages []map[string]string, system string) (*api.ChatResponse, error) {
	prompt := system
	if prompt != "" {
		prompt += "\n\n"
	}
	for _, m := range messages {
		prompt += m["role"] + ": " + m["content"] + "\n"
	}

	var c *exec.Cmd
	switch p.ID {
	case "opencode":
		c = exec.CommandContext(ctx, "opencode", "run", prompt)
	case "claude-code":
		c = exec.CommandContext(ctx, "claude", "-p", prompt)
	case "kimi-code":
		c = exec.CommandContext(ctx, "kimi", "--print", prompt)
	default:
		return nil, fmt.Errorf("unknown CLI provider: %s", p.ID)
	}

	out, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("CLI error: %w", err)
	}

	return &api.ChatResponse{
		Text: strings.TrimSpace(string(out)), Provider: p.ID, Model: model,
	}, nil
}

func buildMessages(messages []map[string]string, system string) []map[string]any {
	out := make([]map[string]any, 0)
	if system != "" {
		out = append(out, map[string]any{"role": "system", "content": system})
	}
	for _, m := range messages {
		out = append(out, map[string]any{"role": m["role"], "content": m["content"]})
	}
	return out
}

func extractText(result map[string]any) string {
	if output, ok := result["output"].(map[string]any); ok {
		if text, ok := output["text"].(string); ok {
			return text
		}
	}
	if choices, ok := result["choices"].([]any); ok && len(choices) > 0 {
		if first, ok := choices[0].(map[string]any); ok {
			if msg, ok := first["message"].(map[string]any); ok {
				if text, ok := msg["content"].(string); ok {
					return text
				}
			}
			if text, ok := first["text"].(string); ok {
				return text
			}
		}
	}
	return ""
}
