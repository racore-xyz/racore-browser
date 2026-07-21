package providers

import (
	"context"
	"errors"
	"testing"
)

func keyFunc(secrets map[string]string) KeyFunc {
	return func(id string) (string, error) {
		if v, ok := secrets[id]; ok {
			return v, nil
		}
		return "", errors.New("not connected")
	}
}

func TestCatalog(t *testing.T) {
	g := NewGateway(keyFunc(map[string]string{}))
	cat := g.Catalog()

	if len(cat) != 9 {
		t.Fatalf("expected 9 providers, got %d", len(cat))
	}

	// Check specific providers exist
	ids := make(map[string]bool)
	for _, p := range cat {
		ids[p.ID] = true
	}
	for _, id := range []string{"openai", "anthropic", "gemini", "ollama", "opencode"} {
		if !ids[id] {
			t.Fatalf("missing provider: %s", id)
		}
	}
}

func TestCatalogConnected(t *testing.T) {
	g := NewGateway(keyFunc(map[string]string{"openai": "sk-test"}))
	cat := g.Catalog()

	for _, p := range cat {
		if p.ID == "openai" && !p.Connected {
			t.Fatal("openai should be connected")
		}
		if p.ID == "ollama" && !p.Connected {
			t.Fatal("ollama should be connected (no auth)")
		}
		if p.ID == "opencode" && !p.Connected {
			t.Fatal("opencode should be connected (local)")
		}
	}
}

func TestCatalogMaskedKey(t *testing.T) {
	g := NewGateway(keyFunc(map[string]string{"openai": "sk-test123456789"}))
	cat := g.Catalog()

	for _, p := range cat {
		if p.ID == "openai" && p.MaskedKey != "sk-...6789" {
			t.Fatalf("expected sk-...6789, got %s", p.MaskedKey)
		}
	}
}

func TestHealthNoKey(t *testing.T) {
	g := NewGateway(keyFunc(map[string]string{}))
	ctx := context.Background()

	result, err := g.Health(ctx, "openai")
	if err != nil {
		t.Fatal(err)
	}
	ok, _ := result["ok"].(bool)
	if ok {
		t.Fatal("expected health to fail without key")
	}
}

func TestHealthUnknownProvider(t *testing.T) {
	g := NewGateway(keyFunc(map[string]string{}))
	ctx := context.Background()

	_, err := g.Health(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestChatUnknownProvider(t *testing.T) {
	g := NewGateway(keyFunc(map[string]string{}))
	ctx := context.Background()

	_, err := g.Chat(ctx, "nonexistent", "", nil, "")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestBuildMessagesWithSystem(t *testing.T) {
	msgs := []map[string]string{
		{"role": "user", "content": "hello"},
	}
	result := buildMessages(msgs, "Be helpful")
	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}
	if result[0]["role"] != "system" {
		t.Fatal("first message should be system")
	}
	if result[1]["role"] != "user" {
		t.Fatal("second message should be user")
	}
}

func TestBuildMessagesWithoutSystem(t *testing.T) {
	msgs := []map[string]string{
		{"role": "user", "content": "hi"},
		{"role": "assistant", "content": "hello"},
	}
	result := buildMessages(msgs, "")
	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}
}

func TestBuildMessagesEmpty(t *testing.T) {
	result := buildMessages(nil, "system only")
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if result[0]["role"] != "system" {
		t.Fatal("should be system message")
	}
}

func TestExtractTextOpenAI(t *testing.T) {
	result := map[string]any{
		"choices": []any{
			map[string]any{
				"message": map[string]any{
					"content": "Hello from AI",
				},
			},
		},
	}
	text := extractText(result)
	if text != "Hello from AI" {
		t.Fatalf("expected 'Hello from AI', got '%s'", text)
	}
}

func TestExtractTextOpenAICompat(t *testing.T) {
	result := map[string]any{
		"choices": []any{
			map[string]any{
				"text": "Simple text response",
			},
		},
	}
	text := extractText(result)
	if text != "Simple text response" {
		t.Fatalf("expected 'Simple text response', got '%s'", text)
	}
}

func TestExtractTextEmpty(t *testing.T) {
	result := map[string]any{}
	text := extractText(result)
	if text != "" {
		t.Fatalf("expected empty, got '%s'", text)
	}
}

func TestProviderCatalogImmutability(t *testing.T) {
	orig := len(Catalog)
	if orig != 9 {
		t.Fatalf("expected 9 providers, got %d", orig)
	}

	// Verify all providers have required fields
	for _, p := range Catalog {
		if p.ID == "" {
			t.Fatal("provider missing ID")
		}
		if p.Name == "" {
			t.Fatal("provider missing Name")
		}
		if p.Kind == "" {
			t.Fatal("provider missing Kind")
		}
		if p.DefaultModel == "" {
			t.Fatal("provider missing DefaultModel")
		}
		switch p.Auth {
		case "bearer", "x-api-key", "query", "none", "local":
		default:
			t.Fatalf("provider %s has unknown auth type: %s", p.ID, p.Auth)
		}
	}
}
