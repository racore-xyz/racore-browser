package localai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/racore/god/pkg/api"
)

const (
	ModelID     = "MadeAgents/Hammer2.0-0.5b"
	ModelLabel  = "Hammer 2.0 0.5B"
	Engine      = "llama.cpp"
	defaultPort = 47834
)

type Manager struct {
	serverPath string
	modelPath  string
	baseURL    string
	client     *http.Client

	mu      sync.Mutex
	command *exec.Cmd
	lastErr string
}

func New() *Manager {
	baseURL := strings.TrimRight(os.Getenv("RACORE_LOCAL_AI_URL"), "/")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://127.0.0.1:%d", defaultPort)
	}
	return &Manager{
		serverPath: os.Getenv("RACORE_LLAMA_SERVER_PATH"),
		modelPath:  os.Getenv("RACORE_LOCAL_MODEL_PATH"),
		baseURL:    baseURL,
		client:     &http.Client{Timeout: 90 * time.Second},
	}
}

func NewClient(baseURL string, client *http.Client) *Manager {
	return &Manager{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	if m.ready(ctx) {
		return nil
	}
	if m.serverPath == "" || m.modelPath == "" {
		return m.fail("bundled local AI paths are not configured")
	}
	if _, err := os.Stat(m.serverPath); err != nil {
		return m.fail("llama.cpp runtime is unavailable: " + err.Error())
	}
	if _, err := os.Stat(m.modelPath); err != nil {
		return m.fail("local model is unavailable: " + err.Error())
	}

	threads := max(1, min(4, runtime.NumCPU()/2))
	command := exec.Command(
		m.serverPath,
		"--model", m.modelPath,
		"--host", "127.0.0.1",
		"--port", fmt.Sprint(defaultPort),
		"--ctx-size", "2048",
		"--threads", fmt.Sprint(threads),
		"--threads-batch", fmt.Sprint(threads),
		"--parallel", "1",
		"--batch-size", "128",
		"--ubatch-size", "64",
		"--no-webui",
	)
	command.Dir = filepath.Dir(m.serverPath)
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	configureBackgroundProcess(command)
	if err := command.Start(); err != nil {
		return m.fail("start llama.cpp: " + err.Error())
	}

	m.mu.Lock()
	m.command = command
	m.lastErr = ""
	m.mu.Unlock()
	go func() {
		err := command.Wait()
		m.mu.Lock()
		if m.command == command {
			m.command = nil
			if err != nil {
				m.lastErr = err.Error()
			}
		}
		m.mu.Unlock()
	}()

	for attempt := 0; attempt < 240; attempt++ {
		if m.ready(ctx) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	m.Stop()
	return m.fail("local model did not become ready within 60 seconds")
}

func (m *Manager) Stop() {
	m.mu.Lock()
	command := m.command
	m.command = nil
	m.mu.Unlock()
	if command != nil && command.Process != nil {
		_ = command.Process.Kill()
	}
}

func (m *Manager) Status(ctx context.Context) api.LocalAIStatus {
	ready := m.ready(ctx)
	m.mu.Lock()
	lastErr := m.lastErr
	m.mu.Unlock()
	state := "offline"
	if ready {
		state = "ready"
	} else if m.command != nil {
		state = "starting"
	} else if lastErr != "" {
		state = "error"
	}
	return api.LocalAIStatus{
		Ready:        ready,
		State:        state,
		Model:        ModelID,
		Label:        ModelLabel,
		Engine:       Engine,
		Parameters:   500_000_000,
		Quantization: "Q8_0",
		LocalOnly:    true,
		Error:        lastErr,
	}
}

func (m *Manager) Chat(ctx context.Context, messages []map[string]string, system string) (*api.ChatResponse, error) {
	if !m.ready(ctx) {
		return nil, errors.New("the bundled local model is not ready")
	}
	prompt := buildPrompt(messages, system)
	body, _ := json.Marshal(map[string]any{
		"prompt":       prompt,
		"n_predict":    320,
		"temperature":  0,
		"cache_prompt": true,
		"stop":         []string{"<|im_end|>"},
	})
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		m.baseURL+"/completion",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	started := time.Now()
	response, err := m.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("local inference failed: %w", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("local inference returned status %d", response.StatusCode)
	}
	var completion struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(raw, &completion); err != nil {
		return nil, fmt.Errorf("invalid local inference response: %w", err)
	}
	actions, text := parsePlan(completion.Content)
	return &api.ChatResponse{
		Text:      text,
		Model:     ModelID,
		Engine:    Engine,
		LatencyMs: time.Since(started).Milliseconds(),
		Actions:   actions,
	}, nil
}

func (m *Manager) ready(ctx context.Context) bool {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+"/health", nil)
	if err != nil {
		return false
	}
	response, err := m.client.Do(request)
	if err != nil {
		return false
	}
	_ = response.Body.Close()
	return response.StatusCode >= 200 && response.StatusCode < 300
}

func (m *Manager) fail(message string) error {
	m.mu.Lock()
	m.lastErr = message
	m.mu.Unlock()
	return errors.New(message)
}

func buildPrompt(messages []map[string]string, system string) string {
	queryParts := make([]string, 0, len(messages))
	for _, message := range messages {
		role := message["role"]
		content := strings.TrimSpace(message["content"])
		if content != "" {
			queryParts = append(queryParts, role+": "+content)
		}
	}
	instruction := `You are Racore's private on-device browser planner. Select one or more tools that satisfy the newest user request. Never claim an action happened; you only create a plan. Use answer_user for a concise direct answer, ask_user when essential information is missing, open_url for an explicit address, search_web for current web research, and resolve_racore_domain for names ending in .racore, .rac, .core, or .ra.`
	if strings.TrimSpace(system) != "" {
		instruction += " Additional policy: " + strings.TrimSpace(system)
	}
	tools := `[
{"name":"answer_user","description":"Answer the user locally without a browser action","parameters":{"text":{"type":"string","required":true}}},
{"name":"ask_user","description":"Ask for essential missing information","parameters":{"question":{"type":"string","required":true}}},
{"name":"open_url","description":"Open an explicit public HTTP or HTTPS URL","parameters":{"url":{"type":"string","required":true}}},
{"name":"search_web","description":"Search the public web","parameters":{"query":{"type":"string","required":true}}},
{"name":"resolve_racore_domain","description":"Resolve a Racore mesh domain","parameters":{"domain":{"type":"string","required":true}}}
]`
	format := `Output only a JSON array. Every item must be {"name":"tool_name","arguments":{...}}. Do not include markdown or other text.`
	return "<|im_start|>user\n[BEGIN OF TASK INSTRUCTION]\n" + instruction +
		"\n[END OF TASK INSTRUCTION]\n\n[BEGIN OF AVAILABLE TOOLS]\n" + tools +
		"\n[END OF AVAILABLE TOOLS]\n\n[BEGIN OF FORMAT INSTRUCTION]\n" + format +
		"\n[END OF FORMAT INSTRUCTION]\n\n[BEGIN OF QUERY]\n" +
		strings.Join(queryParts, "\n") +
		"\n[END OF QUERY]<|im_end|>\n<|im_start|>assistant\n"
}

func parsePlan(raw string) ([]api.BrowserAction, string) {
	raw = strings.TrimSpace(strings.TrimSuffix(raw, "<|im_end|>"))
	type toolCall struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	var calls []toolCall
	if start, end := strings.Index(raw, "["), strings.LastIndex(raw, "]"); start >= 0 && end >= start {
		raw = raw[start : end+1]
		_ = json.Unmarshal([]byte(raw), &calls)
	} else if start, end := strings.Index(raw, "{"), strings.LastIndex(raw, "}"); start >= 0 && end >= start {
		raw = raw[start : end+1]
		var call toolCall
		if json.Unmarshal([]byte(raw), &call) == nil {
			calls = []toolCall{call}
		}
	}
	if len(calls) == 0 {
		return nil, "The local planner returned an invalid plan. Please rephrase the request."
	}
	actions := make([]api.BrowserAction, 0, len(calls))
	var texts []string
	for _, call := range calls {
		switch call.Name {
		case "answer_user":
			if value := strings.TrimSpace(call.Arguments["text"]); value != "" {
				texts = append(texts, value)
			}
		case "ask_user":
			if value := strings.TrimSpace(call.Arguments["question"]); value != "" {
				texts = append(texts, value)
			}
		case "open_url":
			actions = append(actions, api.BrowserAction{Type: call.Name, Value: call.Arguments["url"]})
		case "search_web":
			actions = append(actions, api.BrowserAction{Type: call.Name, Value: call.Arguments["query"]})
		case "resolve_racore_domain":
			actions = append(actions, api.BrowserAction{Type: call.Name, Value: call.Arguments["domain"]})
		}
	}
	if len(texts) > 0 {
		return actions, strings.Join(texts, "\n")
	}
	if len(actions) > 0 {
		return actions, "The local browser plan is ready."
	}
	return nil, "This request does not map to an available local browser action."
}
